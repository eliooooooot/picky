package main

import (
	"flag"
	"fmt"
	"os"
	"github.com/eliooooooot/picky/internal/app"
	"github.com/eliooooooot/picky/internal/fs"
)

func main() {
	var (
		outputPath = flag.String("o", "selected.txt", "output file path")
	)
	flag.Parse()
	
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: picky [options] <directory>")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nInteractive controls:")
		fmt.Fprintln(os.Stderr, "  ↑/↓ or j/k   Navigate up/down")
		fmt.Fprintln(os.Stderr, "  ←/→ or h/l   Collapse/expand directories")
		fmt.Fprintln(os.Stderr, "  Space        Toggle selection")
		fmt.Fprintln(os.Stderr, "  x            Exclude file/directory permanently")
		fmt.Fprintln(os.Stderr, "  s            Settings pane")
		fmt.Fprintln(os.Stderr, "  g            Generate output file")
		fmt.Fprintln(os.Stderr, "  q            Quit")
		fmt.Fprintln(os.Stderr, "\nExcluded paths are saved to .pickyignore in the target directory.")
		os.Exit(1)
	}
	
	rootPath := args[0]
	
	// Create app with OS filesystem
	application := &app.App{
		FS:         fs.NewOSFileSystem(),
		OutputPath: *outputPath,
	}
	
	// Run the application
	if err := application.Run(rootPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}