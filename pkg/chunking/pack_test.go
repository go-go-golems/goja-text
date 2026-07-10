package chunking

import (
	"reflect"
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

func TestPackMarksOversizedSpanAfterFlushingCurrentChunk(t *testing.T) {
	spans := []Span{
		{Ordinal: 0, Text: "ok", StartByte: 0, EndByte: 2, EndRune: 2},
		{Ordinal: 1, Text: "oversized", StartByte: 2, EndByte: 11, StartRune: 2, EndRune: 11},
	}
	result, err := Pack(spans, PackOptions{MaxUnits: 3, Measure: "runes", Oversized: "allow"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Chunks) != 2 || result.Chunks[0].Oversized || !result.Chunks[1].Oversized {
		t.Fatalf("chunks = %#v", result.Chunks)
	}
	if result.Chunks[1].Weight != 9 || len(result.Chunks[1].Diagnostics) != 1 {
		t.Fatalf("oversized chunk = %#v", result.Chunks[1])
	}
}

func TestPackWeightedUsesCallerWeights(t *testing.T) {
	items := []WeightedSpan{
		{Span: Span{Ordinal: 0, Text: "large text", StartByte: 0, EndByte: 10, EndRune: 10}, Weight: 1},
		{Span: Span{Ordinal: 1, Text: "x", StartByte: 10, EndByte: 11, StartRune: 10, EndRune: 11}, Weight: 2},
	}
	result, err := PackWeighted(items, WeightedPackOptions{MaxWeight: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Chunks) != 2 || result.Chunks[0].Weight != 1 || result.Chunks[1].Weight != 2 {
		t.Fatalf("weighted result = %#v", result)
	}
}

func TestPackWeightedOverlapRemainsContiguous(t *testing.T) {
	items := []WeightedSpan{
		{Span: Span{Ordinal: 0, Text: "a", StartByte: 0, EndByte: 1, EndRune: 1}, Weight: 1},
		{Span: Span{Ordinal: 1, Text: "b", StartByte: 1, EndByte: 2, StartRune: 1, EndRune: 2}, Weight: 100},
		{Span: Span{Ordinal: 2, Text: "c", StartByte: 2, EndByte: 3, StartRune: 2, EndRune: 3}, Weight: 1},
		{Span: Span{Ordinal: 3, Text: "d", StartByte: 3, EndByte: 4, StartRune: 3, EndRune: 4}, Weight: 1},
	}
	result, err := PackWeighted(items, WeightedPackOptions{MaxWeight: 102, OverlapWeight: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Chunks) != 2 {
		t.Fatalf("chunks = %#v", result.Chunks)
	}
	second := result.Chunks[1]
	wantOrdinals := []int{2, 3}
	if !reflect.DeepEqual(second.SpanOrdinals, wantOrdinals) {
		t.Fatalf("overlap ordinals = %v, want contiguous suffix %v", second.SpanOrdinals, wantOrdinals)
	}
	if second.Text != "cd" || second.StartByte != 2 || second.EndByte != 4 {
		t.Fatalf("overlap chunk = %#v", second)
	}
}

func TestPackRejectsForgedRanges(t *testing.T) {
	_, err := Pack([]Span{{Ordinal: 0, Text: "abc", StartByte: 0, EndByte: 2, EndRune: 3}}, PackOptions{MaxUnits: 10})
	if err == nil || !strings.Contains(err.Error(), "invalid_range") {
		t.Fatalf("error = %v", err)
	}
}

func TestRecursiveFallsBackToRuneWindowsWithAbsoluteRanges(t *testing.T) {
	source := "# Heading\n\nabcdefghijklmnopqrstuvwxyz"
	result, err := Recursive(source, RecursiveOptions{
		MaxUnits: 8,
		Measure:  "runes",
		Levels:   []string{"markdownSections", "markdownBlocks", "paragraphs", "lines", "runes"},
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
		if strings.Join(chunk.HeadingPath, "/") != "Heading" {
			t.Fatalf("chunk lost heading path: %#v", chunk)
		}
	}
}

func FuzzPackPreservesAllSpansWithoutOverlap(f *testing.F) {
	for _, seed := range []struct {
		source string
		budget int
	}{
		{"", 1},
		{"one\ntwo\n", 4},
		{"é🙂\nnext", 3},
	} {
		f.Add(seed.source, seed.budget)
	}
	f.Fuzz(func(t *testing.T, source string, budget int) {
		if budget <= 0 || budget > 1024 {
			return
		}
		segmented, err := Lines(source, LineOptions{KeepTerminators: true})
		if err != nil {
			return
		}
		result, err := Pack(segmented.Spans, PackOptions{MaxUnits: budget, Measure: "runes", Oversized: "allow"})
		if err != nil {
			t.Fatal(err)
		}
		var joined strings.Builder
		for _, chunk := range result.Chunks {
			if chunk.Text == "" {
				t.Fatal("empty chunk")
			}
			if chunk.Weight > budget && !chunk.Oversized {
				t.Fatalf("unmarked oversized chunk: %#v", chunk)
			}
			joined.WriteString(chunk.Text)
		}
		if joined.String() != source {
			t.Fatalf("joined chunks = %q, want %q", joined.String(), source)
		}
	})
}
