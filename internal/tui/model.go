package tui

import (
	"fmt"
	"github.com/eliooooooot/picky/internal/domain"
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
)

// Model implements the BubbleTea model interface
type Model struct {
	tree               *domain.Tree
	state              domain.ViewState
	vp                 viewport.Model
	paperHeight        int
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
	ta.Placeholder = "press p to add a prompt"
	ta.Prompt = "» "
	ta.CharLimit = 4096
	ta.ShowLineNumbers = false
	ta.SetWidth(80)
	
	vp := viewport.New(0, 0) // Initialize with zero size, will be set on WindowSizeMsg
	
	return &Model{
		tree:           tree,
		state:          domain.NewViewState(tree.Root.Path),
		vp:             vp,
		paperHeight:    20,
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
	// Set default size if not set yet
	if m.vp.Width == 0 {
		m.vp.Width = 80
	}
	if m.vp.Height == 0 {
		m.vp.Height = 20
	}
	// Initialize viewport content
	m.vp.SetContent(m.renderWholeTree())
	m.ensureCursorVisible()
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
				m.prompt.Placeholder = "press p to add a prompt"
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
			m.prompt.Placeholder = ""
			return m, nil
			
		case "up", "k":
			m.state = domain.NavigateUp(m.tree.Root, m.state)
			m.ensureCursorVisible()
			
		case "down", "j":
			m.state = domain.NavigateDown(m.tree.Root, m.state)
			m.ensureCursorVisible()
			
		case "left", "h":
			m.state = domain.NavigateOut(m.tree.Root, m.state)
			// Re-render tree when closing directories
			m.vp.SetContent(m.renderWholeTree())
			m.ensureCursorVisible()
			
		case "right", "l", "enter":
			m.state = domain.NavigateIn(m.tree.Root, m.state)
			// Re-render tree when opening directories
			m.vp.SetContent(m.renderWholeTree())
			m.ensureCursorVisible()
			
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
				m.ensureCursorVisible()
			}
		}
		
	case tea.WindowSizeMsg:
		m.prompt.SetWidth(msg.Width - 4) // Leave room for borders
		// Calculate prompt height
		promptLines := strings.Count(m.prompt.View(), "\n") + 3 // +3 for border and title
		// Calculate paper height for the tree
		m.paperHeight = msg.Height - promptLines - 5 // -5 for existing header/footer
		
		// Update viewport dimensions
		m.vp.Width = msg.Width
		m.vp.Height = m.paperHeight
		
		// Re-render the tree and set content
		m.vp.SetContent(m.renderWholeTree())
		
		// Clamp scroll offset if necessary
		if m.vp.TotalLineCount() > 0 && m.vp.YOffset > m.vp.TotalLineCount()-m.vp.Height {
			m.vp.GotoBottom()
		}
		
		// Ensure cursor remains visible
		m.ensureCursorVisible()
	}
	
	return m, nil
}

// updateSettings handles keyboard input when the settings modal is open
func (m *Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		// Wrap around navigation (2 items now)
		m.settingsCursorIdx = (m.settingsCursorIdx - 1 + 2) % 2
	case "down", "j":
		// Wrap around navigation (2 items now)
		m.settingsCursorIdx = (m.settingsCursorIdx + 1) % 2
	case " ", "enter":
		// Toggle the highlighted setting
		if m.settingsCursorIdx == 0 {
			m.settings = m.settings.ToggleEmoji()
		}
		// Color scheme doesn't use space/enter, it uses left/right
	case "left", "h":
		// Change color scheme (previous)
		if m.settingsCursorIdx == 1 {
			m.settings = m.settings.PrevColorScheme()
		}
	case "right", "l":
		// Change color scheme (next)
		if m.settingsCursorIdx == 1 {
			m.settings = m.settings.NextColorScheme()
		}
	case "esc", "s":
		m.isSettingsOpen = false
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) promptCollapsedView() string {
	// Faint border
	border := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(PromptBorderDim).
		Width(m.vp.Width - 2) // Account for padding

	value := strings.TrimSpace(m.prompt.Value())
	if value == "" {
		value = m.prompt.Placeholder
	} else {
		// Split into lines and limit to 3 lines
		lines := strings.Split(value, "\n")
		if len(lines) > 3 {
			lines = lines[:3]
			// Add ellipsis to the last line
			if len(lines[2]) > 0 {
				lines[2] = lines[2] + "…"
			}
		}
		
		// Rejoin and check total length
		value = strings.Join(lines, "\n")
		
		// If still too long, truncate the last line
		if len(value) > m.vp.Width*3 {
			runes := []rune(value)
			maxRunes := m.vp.Width * 3
			if len(runes) > maxRunes {
				value = string(runes[:maxRunes-1]) + "…"
			}
		}
	}
	value = lipgloss.NewStyle().Faint(true).Render(value)

	return border.Render(value)
}

func (m *Model) dim(s string) string {
	return DimmedStyle.Render(s)
}

// formatInstructions formats instruction commands to fit terminal width
// Never wraps within a command, spans max 2 lines
func (m *Model) formatInstructions(commands []string) string {
	if m.vp.Width == 0 {
		return strings.Join(commands, " • ")
	}
	
	separator := " • "
	sepLen := len(separator)
	width := m.vp.Width
	
	var lines []string
	var currentLine []string
	currentLen := 0
	
	for _, cmd := range commands {
		cmdLen := len(cmd)
		
		// Check if adding this command would exceed width
		neededLen := cmdLen
		if len(currentLine) > 0 {
			neededLen += sepLen + cmdLen
		}
		
		if currentLen + neededLen > width && len(currentLine) > 0 {
			// Start new line
			lines = append(lines, strings.Join(currentLine, separator))
			currentLine = []string{cmd}
			currentLen = cmdLen
			
			// Stop if we already have 2 lines
			if len(lines) >= 2 {
				break
			}
		} else {
			// Add to current line
			currentLine = append(currentLine, cmd)
			if len(currentLine) == 1 {
				currentLen = cmdLen
			} else {
				currentLen += sepLen + cmdLen
			}
		}
	}
	
	// Add remaining commands if we haven't hit line limit
	if len(currentLine) > 0 && len(lines) < 2 {
		lines = append(lines, strings.Join(currentLine, separator))
	}
	
	// If we couldn't fit all commands, add ellipsis to last line
	totalCommands := len(commands)
	commandsFitted := 0
	for _, line := range lines {
		commandsFitted += strings.Count(line, separator) + 1
	}
	
	if commandsFitted < totalCommands && len(lines) > 0 {
		lines[len(lines)-1] += "..."
	}
	
	return strings.Join(lines, "\n")
}

// ensureCursorVisible scrolls the viewport to ensure the cursor is visible
func (m *Model) ensureCursorVisible() {
	flat := domain.Flatten(m.tree.Root, m.state)
	cursor := domain.FindNodeByPath(m.tree.Root, m.state.CursorPath)
	idx := findIndex(flat, cursor)
	
	// The line number in the rendered tree now directly matches the flattened index
	currentLine := idx
	
	// Ensure the cursor line is visible
	if currentLine < m.vp.YOffset {
		m.vp.SetYOffset(currentLine)
	} else if currentLine >= m.vp.YOffset+m.vp.Height {
		m.vp.SetYOffset(currentLine - m.vp.Height + 1)
	}
}

// renderWholeTree renders the complete tree structure without any viewport cropping
func (m *Model) renderWholeTree() string {
	// Build the complete tree starting from root
	items := m.buildTreeItems(m.tree.Root)
	
	// Create tree with items
	t := tree.New().
		EnumeratorStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("241"))).
		Child(items...)
	
	return t.String()
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
	
	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	header := headerStyle.Render(
		fmt.Sprintf("⛏️  Picky   •   Tokens selected: ~%s",
			formatTokenCount(m.selectedTokens())))
	if m.inPromptMode {
		header = m.dim(header)
	}
	b.WriteString(header)
	b.WriteString("\n")
	
	// Instructions
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	
	var instructionText string
	if m.inPromptMode {
		// No instructions shown in prompt mode
		instructionText = ""
	} else {
		instructionText = m.formatInstructions([]string{
			"↑/↓ navigate",
			"←/→ collapse/expand", 
			"space select",
			"x exclude",
			"p prompt",
			"s settings",
			"g generate",
			"q quit",
		})
	}
	
	if instructionText != "" {
		b.WriteString(helpStyle.Render(instructionText))
		b.WriteString("\n")
	}
	
	b.WriteString("\n") // Empty line above prompt box
	
	// Render prompt box after instructions
	b.WriteString(m.renderPrompt())
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
	
	// Set viewport content with the rendered tree
	m.vp.SetContent(m.renderWholeTree())
	
	// Get the tree view from viewport
	treeView := m.vp.View()
	
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
		Width(50)
	
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
	
	// 1. Emoji setting
	emojiSetting := "[x] Emoji icons"
	if !m.settings.Emoji {
		emojiSetting = "[ ] Emoji icons"
	}
	
	if m.settingsCursorIdx == 0 {
		content.WriteString(selectedStyle.Render(emojiSetting))
	} else {
		content.WriteString(normalStyle.Render(emojiSetting))
	}
	
	content.WriteString("\n\n")
	
	// 2. Color scheme setting
	colorSchemeSetting := fmt.Sprintf("Color scheme: ← %s →", m.settings.ColorScheme.Name)
	
	if m.settingsCursorIdx == 1 {
		content.WriteString(selectedStyle.Render(colorSchemeSetting))
	} else {
		content.WriteString(normalStyle.Render(colorSchemeSetting))
	}
	
	content.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	if m.settingsCursorIdx == 1 {
		content.WriteString(helpStyle.Render("↑/↓ navigate  ←/→ change  esc close"))
	} else {
		content.WriteString(helpStyle.Render("↑/↓ navigate  space/enter toggle  esc close"))
	}
	
	// Apply modal styling
	return modalStyle.Render(content.String())
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


// buildTreeItems recursively builds tree items for the entire tree
func (m *Model) buildTreeItems(node *domain.Node) []any {
	// Get the formatted label with cursor/selection styling applied
	label := m.formatNodeLabelWithStyle(node)
	
	// Handle directories with children
	if node.IsDir && m.state.IsOpen(node.Path) && len(node.Children) > 0 {
		childItems := []any{}
		for _, child := range node.Children {
			childItems = append(childItems, m.buildTreeItems(child)...)
		}
		return []any{tree.Root(label).Child(childItems...)}
	}
	
	// Leaf node or closed directory
	return []any{label}
}

// formatNodeLabelWithStyle formats a node's label and applies appropriate styling
func (m *Model) formatNodeLabelWithStyle(node *domain.Node) string {
	// First get the plain formatted label
	label := m.formatNodeLabel(node)
	
	// Determine the appropriate color based on selection state
	var color lipgloss.Color
	if node.IsDir {
		if domain.HasFullSelection(node, m.state) {
			color = m.settings.ColorScheme.Selected
		} else if domain.HasPartialSelection(node, m.state) {
			color = m.settings.ColorScheme.PartiallySelected
		} else {
			color = m.settings.ColorScheme.Unselected
		}
	} else {
		if m.state.IsSelected(node.Path) {
			color = m.settings.ColorScheme.Selected
		} else {
			color = m.settings.ColorScheme.Unselected
		}
	}
	
	// Apply styling based on cursor position
	if node.Path == m.state.CursorPath && !m.inPromptMode {
		// Cursor is on this item - use block background
		// For Classic scheme with dark backgrounds, use white text
		textColor := lipgloss.Color("0") // Default black text
		if m.settings.ColorScheme.Name == "Classic" {
			textColor = lipgloss.Color("255") // White text for Classic
		}
		
		cursorStyle := lipgloss.NewStyle().
			Background(color).
			Foreground(textColor)
		return cursorStyle.Render(label)
	} else {
		// Cursor is elsewhere - use text color only
		textStyle := lipgloss.NewStyle().Foreground(color)
		return textStyle.Render(label)
	}
}

// formatNodeLabel formats a node's label with selection and directory indicators
func (m *Model) formatNodeLabel(node *domain.Node) string {
	// Selection indicator
	selected := " "
	if node.IsDir {
		if domain.HasFullSelection(node, m.state) {
			selected = "✓"
		} else if domain.HasPartialSelection(node, m.state) {
			selected = "~"
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

// VP returns the viewport for testing purposes
func (m *Model) VP() *viewport.Model {
	return &m.vp
}