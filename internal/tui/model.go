package tui

import (
	"fmt"
	"github.com/eliooooooot/picky/internal/domain"
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
)

// Model implements the BubbleTea model interface
type Model struct {
	tree               *domain.Tree
	state              domain.ViewState
	viewportHeight     int
	viewportOffset     int
	requestedGenerate  bool
	newIgnores         map[string]struct{}
	existingIgnores    *map[string]struct{}
	statusMessage      string
	statusMessageTimer int
	tokens             map[string]int
	settings           Settings
	isSettingsOpen     bool
	settingsCursorIdx  int
	prompt             textarea.Model
	inPromptMode       bool
}

// NewModel creates a new TUI model
func NewModel(tree *domain.Tree, existingIgnores *map[string]struct{}) *Model {
	ta := textarea.New()
	ta.Placeholder = "Enter LLM prompt..."
	ta.Prompt = "» "
	ta.CharLimit = 4096
	ta.ShowLineNumbers = false
	ta.SetWidth(80)
	
	return &Model{
		tree:           tree,
		state:          domain.NewViewState(tree.Root.Path),
		viewportHeight: 20,
		viewportOffset: 0,
		newIgnores:     make(map[string]struct{}),
		existingIgnores: existingIgnores,
		settings:       defaultSettings(),
		prompt:         ta,
	}
}

// Tree returns the internal tree (for app layer to access after TUI exits)
func (m *Model) Tree() *domain.Tree {
	return m.tree
}

// State returns the view state (for app layer to access after TUI exits)
func (m *Model) State() domain.ViewState {
	return m.state
}

// RequestedGenerate returns whether the user requested file generation
func (m *Model) RequestedGenerate() bool {
	return m.requestedGenerate
}

// NewIgnores returns the set of newly ignored paths
func (m *Model) NewIgnores() map[string]struct{} {
	return m.newIgnores
}

// SetTokens injects the file-level token map
func (m *Model) SetTokens(t map[string]int) { m.tokens = t }

// Prompt returns the current prompt text
func (m *Model) Prompt() string {
	return m.prompt.Value()
}

// returns file tokens, or aggregated directory tokens
func (m *Model) tokenCount(node *domain.Node) int {
	if m.tokens == nil {
		return 0
	}
	if !node.IsDir {
		return m.tokens[node.Path]
	}
	// directory: sum tokens of descendant files still in tree
	var sum int
	var stack []*domain.Node
	stack = append(stack, node)
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if cur.IsDir {
			stack = append(stack, cur.Children...)
		} else {
			sum += m.tokens[cur.Path]
		}
	}
	return sum
}

func (m *Model) selectedTokens() int {
	if m.tokens == nil {
		return 0
	}
	total := 0
	for _, p := range domain.GetSelectedPaths(m.tree.Root, m.state) {
		total += m.tokens[p]
	}
	return total
}

// Init implements tea.Model
func (m *Model) Init() tea.Cmd {
	// Open the root directory by default
	m.state = m.state.SetOpen(m.tree.Root.Path, true)
	return nil
}

// Update implements tea.Model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Decrement status message timer
	if m.statusMessageTimer > 0 {
		m.statusMessageTimer--
	}
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global quit works regardless of mode
		if key := msg.String(); key == "ctrl+c" || (key == "q" && !m.inPromptMode) {
			return m, tea.Quit
		}
		
		// Handle prompt mode
		if m.inPromptMode {
			switch msg.String() {
			case "esc":
				m.inPromptMode = false
				m.prompt.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.prompt, cmd = m.prompt.Update(msg)
			return m, cmd
		}
		
		// Handle settings modal if open
		if m.isSettingsOpen {
			return m.updateSettings(msg)
		}
		
		switch msg.String() {
		case "p":
			m.inPromptMode = true
			m.prompt.Focus()
			return m, nil
			
		case "up", "k":
			m.state = domain.NavigateUp(m.tree.Root, m.state)
			m.updateViewport()
			
		case "down", "j":
			m.state = domain.NavigateDown(m.tree.Root, m.state)
			m.updateViewport()
			
		case "left", "h":
			m.state = domain.NavigateOut(m.tree.Root, m.state)
			
		case "right", "l", "enter":
			m.state = domain.NavigateIn(m.tree.Root, m.state)
			
		case " ":
			m.state = domain.ToggleSelection(m.tree.Root, m.state)
			
		case "g":
			m.requestedGenerate = true
			return m, tea.Quit
			
		case "s":
			if m.isSettingsOpen {
				m.isSettingsOpen = false
			} else {
				m.isSettingsOpen = true
				m.settingsCursorIdx = 0
			}
			return m, nil
			
		case "x":
			// Exclude current node
			// First, find the current node before excluding it
			currentNode := domain.FindNodeByPath(m.tree.Root, m.state.CursorPath)
			if currentNode == nil {
				break
			}
			
			// Get the flattened list before exclusion to find next cursor position
			flatBefore := domain.Flatten(m.tree.Root, m.state)
			currentIdx := -1
			for i, node := range flatBefore {
				if node.Path == m.state.CursorPath {
					currentIdx = i
					break
				}
			}
			
			// Now exclude the node
			if relPath, removedNode := m.tree.ExcludeNode(m.state.CursorPath); removedNode != nil {
				m.newIgnores[relPath] = struct{}{}
				m.statusMessage = fmt.Sprintf("Excluded: %s", relPath)
				m.statusMessageTimer = 30 // Show for ~1 second (30 frames)
				
				// Clean up ViewState by removing references to the excluded path
				m.state = m.state.Prune(removedNode.Path)
				
				// Get flattened list after exclusion
				flatAfter := domain.Flatten(m.tree.Root, m.state)
				
				if len(flatAfter) == 0 {
					// No nodes left, this shouldn't happen as root can't be excluded
					break
				}
				
				// Determine new cursor position using domain logic
				newCursorPath := domain.NextCursorAfterRemoval(flatBefore, currentIdx, flatAfter)
				m.state = m.state.SetCursor(newCursorPath)
				m.updateViewport()
			}
		}
		
	case tea.WindowSizeMsg:
		m.prompt.SetWidth(msg.Width - 4) // Leave room for borders
		// Calculate prompt height
		promptLines := strings.Count(m.prompt.View(), "\n") + 3 // +3 for border and title
		// Adjust viewport height to account for prompt
		m.viewportHeight = msg.Height - promptLines - 5 // -5 for existing header/footer
		m.updateViewport()
	}
	
	return m, nil
}

// updateSettings handles keyboard input when the settings modal is open
func (m *Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// Wrap around navigation (only 1 item for now)
		m.settingsCursorIdx = (m.settingsCursorIdx - 1 + 1) % 1
	case "down", "j":
		// Wrap around navigation (only 1 item for now)
		m.settingsCursorIdx = (m.settingsCursorIdx + 1) % 1
	case " ", "enter":
		// Toggle the highlighted setting
		if m.settingsCursorIdx == 0 {
			m.settings = m.settings.ToggleEmoji()
		}
	case "esc", "s":
		m.isSettingsOpen = false
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) promptCollapsedView() string {
	// One-liner, faint border
	border := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(PromptBorderDim)

	value := strings.TrimSpace(m.prompt.Value())
	if value == "" {
		value = m.prompt.Placeholder
	} else if runeCount := len([]rune(value)); runeCount > 60 {
		value = string([]rune(value)[:60]) + "…"
	}
	value = lipgloss.NewStyle().Faint(true).Render(value)

	return border.Render(" " + value + " ")
}

func (m *Model) dim(s string) string {
	return DimmedStyle.Render(s)
}

// renderPrompt renders the prompt box
func (m *Model) renderPrompt() string {
	if !m.inPromptMode {
		return m.promptCollapsedView()
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(PromptBorderLit).
		Width(m.prompt.Width() + 4)

	title := lipgloss.NewStyle().
		Foreground(PromptBorderLit).
		Bold(true).
		Render(" Prompt (esc to close)")

	return lipgloss.JoinVertical(
		lipgloss.Top,
		title,
		boxStyle.Render(m.prompt.View()),
	)
}

// View implements tea.Model
func (m *Model) View() string {
	var b strings.Builder
	
	// Render prompt box first
	b.WriteString(m.renderPrompt())
	b.WriteString("\n")
	
	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	header := headerStyle.Render(
		fmt.Sprintf("Picky - File Selector   •   Tokens selected: ~%s",
			formatTokenCount(m.selectedTokens())))
	if m.inPromptMode {
		header = m.dim(header)
	}
	b.WriteString(header)
	b.WriteString("\n\n")
	
	// Instructions
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	if m.inPromptMode {
		// Only esc + quit hints
		b.WriteString(helpStyle.Render("esc: close prompt • ctrl+c/q: quit"))
	} else {
		b.WriteString(helpStyle.Render("↑/↓ or j/k navigate • ←/→ collapse/expand • space select • x exclude • p prompt • s settings • g generate • q quit"))
	}
	b.WriteString("\n")
	
	// Status message
	if m.statusMessageTimer > 0 {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		statusLine := statusStyle.Render(m.statusMessage)
		if m.inPromptMode {
			statusLine = m.dim(statusLine)
		}
		b.WriteString(statusLine)
	}
	b.WriteString("\n")
	
	// Build and render tree
	treeView := m.buildTreeView()
	
	if m.inPromptMode {
		treeView = m.dim(treeView)
	}
	
	// Handle settings modal
	if m.isSettingsOpen {
		// Render dimmed tree as background
		dimmedTreeStyle := lipgloss.NewStyle().Faint(true)
		b.WriteString(dimmedTreeStyle.Render(treeView))
		
		// Render settings modal overlay
		settingsView := m.renderSettingsModal()
		// Position the modal over the tree
		// We'll overlay it by using ANSI cursor positioning or just append it
		b.WriteString("\n\n")
		b.WriteString(settingsView)
	} else {
		// Normal tree view
		b.WriteString(treeView)
	}
	
	return b.String()
}

// renderSettingsModal renders the settings modal overlay
func (m *Model) renderSettingsModal() string {
	// Define styles
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(40)
	
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6"))
	
	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("238")).
		Foreground(lipgloss.Color("255"))
	
	normalStyle := lipgloss.NewStyle()
	
	// Build content
	var content strings.Builder
	
	// Title
	content.WriteString(titleStyle.Render("Settings"))
	content.WriteString("\n\n")
	
	// Settings items
	emojiSetting := "[x] Emoji icons"
	if !m.settings.Emoji {
		emojiSetting = "[ ] Emoji icons"
	}
	
	// Apply style based on cursor position
	if m.settingsCursorIdx == 0 {
		content.WriteString(selectedStyle.Render(emojiSetting))
	} else {
		content.WriteString(normalStyle.Render(emojiSetting))
	}
	
	content.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	content.WriteString(helpStyle.Render("↑/↓ navigate  space/enter toggle  esc close"))
	
	// Apply modal styling
	return modalStyle.Render(content.String())
}

func (m *Model) updateViewport() {
	flat := domain.Flatten(m.tree.Root, m.state)
	cursor := domain.FindNodeByPath(m.tree.Root, m.state.CursorPath)
	cursorIdx := findIndex(flat, cursor)
	
	if cursorIdx < m.viewportOffset {
		m.viewportOffset = cursorIdx
	} else if cursorIdx >= m.viewportOffset+m.viewportHeight {
		m.viewportOffset = cursorIdx - m.viewportHeight + 1
	}
}

func findIndex(nodes []*domain.Node, target *domain.Node) int {
	for i, n := range nodes {
		if n == target {
			return i
		}
	}
	return 0
}

func getDepth(node *domain.Node) int {
	depth := 0
	current := node
	for current.Parent != nil {
		depth++
		current = current.Parent
	}
	return depth
}

// shouldRenderAsSelected determines if a node should be rendered with selected styling
func shouldRenderAsSelected(node *domain.Node, state domain.ViewState) bool {
	// Files are selected if explicitly selected or if any parent directory has full selection
	if !node.IsDir {
		if state.IsSelected(node.Path) {
			return true
		}
		// Check if any parent directory has full selection
		parent := node.Parent
		for parent != nil {
			if domain.HasFullSelection(parent, state) {
				return true
			}
			parent = parent.Parent
		}
		return false
	}
	
	// Directories are selected only if they have full selection
	return domain.HasFullSelection(node, state)
}

// buildTreeView builds the tree view using lipgloss/tree
func (m *Model) buildTreeView() string {
	// Get flattened view for viewport calculation
	flat := domain.Flatten(m.tree.Root, m.state)
	cursor := domain.FindNodeByPath(m.tree.Root, m.state.CursorPath)
	
	// Adjust viewport to ensure cursor is visible
	cursorIdx := findIndex(flat, cursor)
	if cursorIdx < m.viewportOffset {
		m.viewportOffset = cursorIdx
	} else if cursorIdx >= m.viewportOffset+m.viewportHeight {
		m.viewportOffset = cursorIdx - m.viewportHeight + 1
	}
	
	// Style configuration
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	cursorStyle := CursorStyle
	
	// Build tree starting from root's children
	items := m.buildTreeItems(m.tree.Root, flat, m.viewportOffset, m.viewportOffset+m.viewportHeight, selectedStyle, cursorStyle)
	
	// Create tree with items
	t := tree.New().
		EnumeratorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("241"))).
		Child(items...)
	
	return t.String()
}

// buildTreeItems recursively builds tree items within the viewport range
func (m *Model) buildTreeItems(node *domain.Node, flat []*domain.Node, start, end int, selectedStyle, cursorStyle lipgloss.Style) []any {
	var items []any
	
	// Process root's children directly
	if node == m.tree.Root {
		for _, child := range node.Children {
			childItems := m.buildTreeItems(child, flat, start, end, selectedStyle, cursorStyle)
			items = append(items, childItems...)
		}
		return items
	}
	
	// Find node index in flat list
	nodeIdx := -1
	for i, n := range flat {
		if n.Path == node.Path {
			nodeIdx = i
			break
		}
	}
	
	// Skip if node is outside viewport
	if nodeIdx != -1 && (nodeIdx < start || nodeIdx >= end) {
		// But check if any children might be visible
		if node.IsDir && m.state.IsOpen(node.Path) {
			for _, child := range node.Children {
				childIdx := -1
				for i, n := range flat {
					if n.Path == child.Path {
						childIdx = i
						break
					}
				}
				if childIdx >= start && childIdx < end {
					// Add this node with its children
					label := m.formatNodeLabel(node)
					if node.Path == m.state.CursorPath && !m.inPromptMode {
						label = cursorStyle.Render(label)
					} else if shouldRenderAsSelected(node, m.state) {
						label = selectedStyle.Render(label)
					}
					
					childItems := []any{}
					for _, c := range node.Children {
						childItems = append(childItems, m.buildTreeItems(c, flat, start, end, selectedStyle, cursorStyle)...)
					}
					
					return []any{tree.Root(label).Child(childItems...)}
				}
			}
		}
		return nil
	}
	
	// Node is visible - render it
	label := m.formatNodeLabel(node)
	
	// Apply styles
	if node.Path == m.state.CursorPath && !m.inPromptMode {
		label = cursorStyle.Render(label)
	} else if shouldRenderAsSelected(node, m.state) {
		label = selectedStyle.Render(label)
	}
	
	// Handle directories with children
	if node.IsDir && m.state.IsOpen(node.Path) && len(node.Children) > 0 {
		childItems := []any{}
		for _, child := range node.Children {
			childItems = append(childItems, m.buildTreeItems(child, flat, start, end, selectedStyle, cursorStyle)...)
		}
		return []any{tree.Root(label).Child(childItems...)}
	}
	
	// Leaf node or closed directory
	return []any{label}
}

// formatNodeLabel formats a node's label with selection and directory indicators
func (m *Model) formatNodeLabel(node *domain.Node) string {
	// Selection indicator
	selected := " "
	if node.IsDir {
		if domain.HasFullSelection(node, m.state) {
			selected = "✓"
		} else if domain.HasPartialSelection(node, m.state) {
			selected = "-"
		}
	} else if m.state.IsSelected(node.Path) {
		selected = "✓"
	}
	
	// Directory indicator and emoji
	name := node.Name
	
	if node.IsDir {
		// For directories, either use arrows OR emoji
		if m.settings.Emoji {
			// In emoji mode, replace arrows with folder emoji
			if m.state.IsOpen(node.Path) {
				name = "📂 " + name  // Open folder emoji
			} else {
				name = "📁 " + name  // Closed folder emoji
			}
		} else {
			// Normal arrow mode
			if m.state.IsOpen(node.Path) {
				name = "▼ " + name
			} else {
				name = "▶ " + name
			}
		}
	} else {
		// For files, add emoji if enabled
		if m.settings.Emoji {
			name = "📄 " + name
		}
	}
	
	tok := m.tokenCount(node)
	// final label: "[✓] [▶ dir] (123)"
	return fmt.Sprintf("%s %s (%s)", selected, name, formatTokenCount(tok))
}

// formatTokenCount formats a token count with k/M suffixes for large numbers
func formatTokenCount(count int) string {
	if count < 1000 {
		return fmt.Sprintf("%d", count)
	}
	if count < 1000000 {
		return fmt.Sprintf("%.1fk", float64(count)/1000)
	}
	return fmt.Sprintf("%.1fM", float64(count)/1000000)
}