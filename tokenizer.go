package rtengo

import (
	"fmt"
	"os"

	hftokenizer "github.com/masacento/tokenizer"
	"github.com/masacento/tokenizer/pretrained"
)

// Tokenizer defines the interface for tokenizers.
type Tokenizer interface {
	Encode(text string, addSpecial bool) (*Encoding, error)
	Decode(ids []int) string
}

// Encoding represents the result of tokenization.
type Encoding struct {
	Tokens        []string
	IDs           []int
	AttentionMask []int
}

// HuggingFaceTokenizer implements Tokenizer using a Hugging Face tokenizer.json file.
type HuggingFaceTokenizer struct {
	tokenizer *hftokenizer.Tokenizer
}

// NewTokenizerFromFile creates a tokenizer from a Hugging Face tokenizer.json file.
func NewTokenizerFromFile(tokenizerPath string) (*HuggingFaceTokenizer, error) {
	if _, err := os.Stat(tokenizerPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tokenizer file not found at %s", tokenizerPath)
	}

	tk, err := pretrained.FromFile(tokenizerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tokenizer: %w", err)
	}

	return &HuggingFaceTokenizer{tokenizer: tk}, nil
}

// Encode tokenizes the input text and returns an Encoding.
func (t *HuggingFaceTokenizer) Encode(text string, addSpecialTokens bool) (*Encoding, error) {
	encoding, err := t.tokenizer.EncodeSingle(text, addSpecialTokens)
	if err != nil {
		return nil, err
	}

	return &Encoding{
		Tokens:        encoding.Tokens,
		IDs:           encoding.Ids,
		AttentionMask: encoding.AttentionMask,
	}, nil
}

// Decode converts token IDs back to text.
func (t *HuggingFaceTokenizer) Decode(ids []int) string {
	return t.tokenizer.Decode(ids, true)
}
