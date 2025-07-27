package ignore

import (
	"bufio"
	"path/filepath"
	"sort"
	"strings"

	"github.com/eliooooooot/picky/internal/domain"
)

const ignoreFileName = ".pickyignore"

func Load(fs domain.FileSystem, root string) (map[string]struct{}, error) {
	ignoreFilePath := filepath.Join(root, ignoreFileName)
	ignores := make(map[string]struct{})

	info, err := fs.Stat(ignoreFilePath)
	if err != nil || info.IsDir() {
		return ignores, nil
	}

	data, err := fs.ReadFile(ignoreFilePath)
	if err != nil {
		return ignores, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Normalize path for cross-platform compatibility
		normalizedPath := filepath.ToSlash(line)
		ignores[normalizedPath] = struct{}{}
	}

	return ignores, scanner.Err()
}

func Save(fs domain.FileSystem, root string, set map[string]struct{}) error {
	if len(set) == 0 {
		return nil
	}

	ignoreFilePath := filepath.Join(root, ignoreFileName)
	
	// Collect and normalize paths
	paths := make([]string, 0, len(set))
	for path := range set {
		// Normalize path for cross-platform compatibility
		normalizedPath := filepath.ToSlash(path)
		paths = append(paths, normalizedPath)
	}
	
	// Sort for stable output
	sort.Strings(paths)
	
	// Join with newlines
	content := strings.Join(paths, "\n")
	if len(paths) > 0 {
		content += "\n"
	}

	return fs.WriteFile(ignoreFilePath, []byte(content), 0644)
}