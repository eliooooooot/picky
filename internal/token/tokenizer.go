package token

// Tokenizer can be swapped for GPT-2, tiktoken, etc.
type Tokenizer interface {
	// CountTokens returns the number of tokens in the given text.
	CountTokens(text string) int
}

type NaiveTokenizer struct{}

func (NaiveTokenizer) CountTokens(text string) int {
	// 4 UTF-8 characters per token (round up).
	return (len([]rune(text)) + 3) / 4
}