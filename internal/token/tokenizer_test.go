package token

import "testing"

func TestNaiveTokenizer_CountTokens(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "single character",
			text:     "a",
			expected: 1,
		},
		{
			name:     "exactly 4 characters",
			text:     "test",
			expected: 1,
		},
		{
			name:     "5 characters rounds up",
			text:     "hello",
			expected: 2,
		},
		{
			name:     "8 characters",
			text:     "hello123",
			expected: 2,
		},
		{
			name:     "unicode characters count correctly",
			text:     "你好世界", // 4 Chinese characters
			expected: 1,
		},
		{
			name:     "mixed ASCII and unicode",
			text:     "hello世界", // 7 characters total
			expected: 2,
		},
		{
			name:     "15 characters rounds up to 4",
			text:     "this is a test!",
			expected: 4,
		},
	}

	tokenizer := NaiveTokenizer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenizer.CountTokens(tt.text)
			if got != tt.expected {
				t.Errorf("CountTokens(%q) = %d, want %d", tt.text, got, tt.expected)
			}
		})
	}
}