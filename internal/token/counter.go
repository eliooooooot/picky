package token

import (
	"github.com/eliooooooot/picky/internal/domain"
)

// Counter turns a Tokenizer + domain.FileSystem into a path→count map.
type Counter struct {
	FS        domain.FileSystem
	Tokenizer Tokenizer
	cache     map[string]int // avoids re-reading files
}

func NewCounter(fs domain.FileSystem, tz Tokenizer) *Counter {
	return &Counter{
		FS: fs, Tokenizer: tz, cache: make(map[string]int),
	}
}

// BuildTreeTokenMap walks every *file* node and produces a map[path]tokens.
func (c *Counter) BuildTreeTokenMap(t *domain.Tree) (map[string]int, error) {
	out := make(map[string]int)
	for _, n := range t.Flatten() {
		if n.IsDir {
			continue
		}
		tokens, err := c.tokensForFile(n.Path)
		if err != nil {
			return nil, err
		}
		out[n.Path] = tokens
	}
	return out, nil
}

// tokensForFile is cached per-path.
func (c *Counter) tokensForFile(path string) (int, error) {
	if v, ok := c.cache[path]; ok {
		return v, nil
	}
	bytes, err := c.FS.ReadFile(path)
	if err != nil {
		return 0, err
	}
	v := c.Tokenizer.CountTokens(string(bytes))
	c.cache[path] = v
	return v, nil
}