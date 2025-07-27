package app

import (
	"fmt"
	"path/filepath"
	"github.com/eliotsamuelmiller/picky/internal/domain"
	"github.com/eliotsamuelmiller/picky/internal/generate"
	"github.com/eliotsamuelmiller/picky/internal/ignore"
	"github.com/eliotsamuelmiller/picky/internal/token"
	"github.com/eliotsamuelmiller/picky/internal/tui"
	
	tea "github.com/charmbracelet/bubbletea"
)

// App orchestrates the file selector application
type App struct {
	FS         domain.FileSystem
	OutputPath string
}

// Run executes the application
func (a *App) Run(rootPath string) error {
	// Load existing ignores
	ignores, err := ignore.Load(a.FS, rootPath)
	if err != nil {
		return fmt.Errorf("load ignores: %w", err)
	}
	
	// Build the file tree with filter
	keep := func(p string, isDir bool) bool {
		rel, err := filepath.Rel(rootPath, p)
		if err != nil {
			return true // Keep on error
		}
		// Normalize path for cross-platform compatibility
		normalizedRel := filepath.ToSlash(rel)
		_, skip := ignores[normalizedRel]
		return !skip
	}
	
	tree, err := domain.BuildTreeWithFilter(a.FS, rootPath, keep)
	if err != nil {
		return fmt.Errorf("build tree: %w", err)
	}
	
	// --- token counting --------------------------------------------------
	tc := token.NewCounter(a.FS, token.NaiveTokenizer{})
	tokensMap, err := tc.BuildTreeTokenMap(tree)
	if err != nil {
		return fmt.Errorf("token count: %w", err)
	}
	// ---------------------------------------------------------------------
	
	// Create and run the TUI
	model := tui.NewModel(tree, &ignores)
	model.SetTokens(tokensMap)
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("run tui: %w", err)
	}
	
	// Check if user requested generation
	m, ok := finalModel.(*tui.Model)
	if !ok {
		return fmt.Errorf("unexpected model type: %T", finalModel)
	}
	
	// Save new ignores if any
	if len(m.NewIgnores()) > 0 {
		for k := range m.NewIgnores() {
			ignores[k] = struct{}{}
		}
		if err := ignore.Save(a.FS, rootPath, ignores); err != nil {
			return fmt.Errorf("save ignores: %w", err)
		}
	}
	
	if m.RequestedGenerate() {
		outputPath := a.OutputPath
		if outputPath == "" {
			outputPath = "selected.txt"
		}
		
		if err := generate.Generate(outputPath, m.Prompt(), m.Tree(), m.State(), a.FS); err != nil {
			return fmt.Errorf("generate output: %w", err)
		}
		
		fmt.Printf("Output written to: %s\n", outputPath)
	}
	
	return nil
}