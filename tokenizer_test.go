package rtengo

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSentencePieceTokenizer_EncodeJapanese(t *testing.T) {
	tokenizer, err := loadTestTokenizer()
	if err != nil {
		t.Fatalf("Failed to load tokenizer: %v", err)
	}

	text := "川べりでサーフボードを持った人たちがいます"
	encoding, err := tokenizer.Encode(text, false)
	if err != nil {
		t.Fatalf("Failed to encode text: %v", err)
	}

	if len(encoding.IDs) == 0 {
		t.Error("Expected at least one token, got 0")
	}
	if len(encoding.Tokens) == 0 {
		t.Error("Expected at least one token string, got 0")
	}
	if len(encoding.AttentionMask) != len(encoding.IDs) {
		t.Fatalf("Attention mask length (%d) doesn't match IDs length (%d)",
			len(encoding.AttentionMask), len(encoding.IDs))
	}

	for i, mask := range encoding.AttentionMask {
		if mask != 1 {
			t.Errorf("Expected attention mask[%d] to be 1, got %d", i, mask)
		}
	}

	decoded := tokenizer.Decode(encoding.IDs)
	if decoded == "" {
		t.Error("Decoded text should not be empty")
	}
}

func TestSentencePieceTokenizer_EncodeWithSpecialTokens(t *testing.T) {
	tokenizer, err := loadTestTokenizer()
	if err != nil {
		t.Fatalf("Failed to load tokenizer: %v", err)
	}

	text := "川べりでサーフボードを持った人たちがいます"
	encoding, err := tokenizer.Encode(text, true)
	if err != nil {
		t.Fatalf("Failed to encode text: %v", err)
	}

	if len(encoding.IDs) < 2 {
		t.Fatalf("Expected at least two special tokens, got %d", len(encoding.IDs))
	}
	if encoding.Tokens[0] != "<s>" {
		t.Errorf("Expected first token to be <s>, got %q", encoding.Tokens[0])
	}
	if encoding.Tokens[len(encoding.Tokens)-1] != "</s>" {
		t.Errorf("Expected last token to be </s>, got %q", encoding.Tokens[len(encoding.Tokens)-1])
	}
}

func TestSentencePieceTokenizer_DecodeEmpty(t *testing.T) {
	tokenizer, err := loadTestTokenizer()
	if err != nil {
		t.Fatalf("Failed to load tokenizer: %v", err)
	}

	if got := tokenizer.Decode(nil); got != "" {
		t.Fatalf("Expected empty decode, got %q", got)
	}
}

// LoadTokenizerFromDefaultPath loads a tokenizer from the default testdata path.
func loadTestTokenizer() (*SentencePieceTokenizer, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	tokenizerPath := filepath.Join(wd, "testdata", "tokenizer.model")
	if _, err := os.Stat(tokenizerPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tokenizer file not found at %s", tokenizerPath)
	}

	return NewSentencePiece(tokenizerPath)
}
