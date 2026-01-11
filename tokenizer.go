package rtengo

import (
	"fmt"
	"os"

	"github.com/masacento/go-sentencepiece"
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

// SentencePieceTokenizer implements Tokenizer using sentencepiece.
type SentencePieceTokenizer struct {
	processor *sentencepiece.Processor
}

// NewSentencePiece creates a new SentencePiece tokenizer from a model file.
func NewSentencePiece(modelPath string) (*SentencePieceTokenizer, error) {
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tokenizer file not found at %s", modelPath)
	}

	proc, err := sentencepiece.NewProcessorFromPath(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load sentencepiece model: %w", err)
	}

	return &SentencePieceTokenizer{processor: proc}, nil
}

// Encode tokenizes the input text and returns an Encoding.
func (t *SentencePieceTokenizer) Encode(text string, addSpecialTokens bool) (*Encoding, error) {
	tokenList := t.processor.Encode(text)

	tokenIDs := make([]int, len(tokenList))
	tokens := make([]string, len(tokenList))
	for i, token := range tokenList {
		tokenIDs[i] = token.ID
		tokens[i] = token.Text
	}

	if addSpecialTokens {
		tokenIDs = append([]int{1}, tokenIDs...)
		tokens = append([]string{"<s>"}, tokens...)
		tokenIDs = append(tokenIDs, 2)
		tokens = append(tokens, "</s>")
	}

	attentionMask := make([]int, len(tokenIDs))
	for i := range attentionMask {
		attentionMask[i] = 1
	}

	return &Encoding{
		Tokens:        tokens,
		IDs:           tokenIDs,
		AttentionMask: attentionMask,
	}, nil
}

// Decode converts token IDs back to text.
func (t *SentencePieceTokenizer) Decode(ids []int) string {
	return t.processor.Decode(ids)
}
