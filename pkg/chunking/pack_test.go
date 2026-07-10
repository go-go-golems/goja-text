package chunking

import (
	"strings"
	"testing"
)

func TestPackGreedyBudgetAndWholeSpanOverlap(t *testing.T) {
	segmented, err := Lines("aa\nbb\ncc\n", LineOptions{KeepTerminators: true})
	if err != nil {
		t.Fatal(err)
	}
	result, err := Pack(segmented.Spans, PackOptions{
		MaxUnits: 6,
		Measure:  "bytes",
		Overlap:  OverlapOptions{Unit: "spans", Value: 1},
	})
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}
	if len(result.Chunks) != 2 {
		t.Fatalf("chunks = %d, want 2", len(result.Chunks))
	}
	if result.Chunks[0].Text != "aa\nbb\n" || result.Chunks[1].Text != "bb\ncc\n" {
		t.Fatalf("chunk texts = %q, %q", result.Chunks[0].Text, result.Chunks[1].Text)
	}
	for _, chunk := range result.Chunks {
		if chunk.Weight > 6 || chunk.Text == "" {
			t.Fatalf("invalid chunk %#v", chunk)
		}
	}
}

func TestPackOversizedPolicy(t *testing.T) {
	spans := []Span{{Ordinal: 0, Text: "oversized", StartByte: 0, EndByte: 9, EndRune: 9}}
	allowed, err := Pack(spans, PackOptions{MaxUnits: 3, Measure: "runes", Oversized: "allow"})
	if err != nil {
		t.Fatal(err)
	}
	if !allowed.Chunks[0].Oversized || len(allowed.Diagnostics) != 1 {
		t.Fatalf("allowed = %#v", allowed)
	}
	_, err = Pack(spans, PackOptions{MaxUnits: 3, Measure: "runes", Oversized: "error"})
	if err == nil || !strings.Contains(err.Error(), "span_exceeds_budget") {
		t.Fatalf("error = %v", err)
	}
}

func TestPackWeightedUsesCallerWeights(t *testing.T) {
	items := []WeightedSpan{
		{Span: Span{Ordinal: 0, Text: "large text", StartByte: 0, EndByte: 10}, Weight: 1},
		{Span: Span{Ordinal: 1, Text: "x", StartByte: 10, EndByte: 11}, Weight: 2},
	}
	result, err := PackWeighted(items, WeightedPackOptions{MaxWeight: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Chunks) != 2 || result.Chunks[0].Weight != 1 || result.Chunks[1].Weight != 2 {
		t.Fatalf("weighted result = %#v", result)
	}
}

func TestRecursiveFallsBackToRuneWindowsWithAbsoluteRanges(t *testing.T) {
	source := "# Heading\n\nabcdefghijklmnopqrstuvwxyz"
	result, err := Recursive(source, RecursiveOptions{
		MaxUnits: 8,
		Measure:  "runes",
		Levels:   []string{"markdownBlocks", "paragraphs", "lines", "runes"},
	})
	if err != nil {
		t.Fatalf("Recursive: %v", err)
	}
	if len(result.Chunks) < 4 {
		t.Fatalf("chunks = %d, want fallback windows", len(result.Chunks))
	}
	for _, chunk := range result.Chunks {
		if chunk.Weight > 8 || source[chunk.StartByte:chunk.EndByte] != chunk.Text {
			t.Fatalf("invalid absolute chunk %#v", chunk)
		}
	}
}
