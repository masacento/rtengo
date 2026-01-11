package rtengo

import "testing"

func TestSplitIntoSentences(t *testing.T) {
	input := "Hello world. How are you? I'm fine!"
	got := SplitIntoSentences(input)
	if len(got) != 3 {
		t.Fatalf("Expected 3 sentences, got %d", len(got))
	}
	if got[0] != "Hello world." {
		t.Fatalf("Unexpected first sentence: %q", got[0])
	}
}

func TestSplitIntoSentences_NoPunctuation(t *testing.T) {
	input := "  No punctuation here  "
	got := SplitIntoSentences(input)
	if len(got) != 1 {
		t.Fatalf("Expected 1 sentence, got %d", len(got))
	}
	if got[0] != "No punctuation here" {
		t.Fatalf("Unexpected sentence: %q", got[0])
	}
}

func TestSigmoid(t *testing.T) {
	if v := sigmoid(0); v < 0.49 || v > 0.51 {
		t.Fatalf("Expected sigmoid(0) ~ 0.5, got %f", v)
	}
}
