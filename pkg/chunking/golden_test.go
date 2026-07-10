package chunking

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type markdownBlocksGolden struct {
	Source string                    `json:"source"`
	Spans  []markdownBlockGoldenSpan `json:"spans"`
}

type markdownBlockGoldenSpan struct {
	Kind         string `json:"kind"`
	Text         string `json:"text"`
	StartByte    int    `json:"startByte"`
	EndByte      int    `json:"endByte"`
	HeadingLevel int    `json:"headingLevel"`
	Atomic       bool   `json:"atomic"`
	Language     string `json:"language"`
}

func TestMarkdownBlocksGolden(t *testing.T) {
	data, err := os.ReadFile("testdata/markdown_blocks.golden.json")
	if err != nil {
		t.Fatalf("read golden fixture: %v", err)
	}
	var golden markdownBlocksGolden
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("decode golden fixture: %v", err)
	}

	result, err := MarkdownBlocks(golden.Source, MarkdownBlockOptions{Atomic: []string{"fencedCodeBlock"}})
	if err != nil {
		t.Fatalf("MarkdownBlocks: %v", err)
	}
	got := make([]markdownBlockGoldenSpan, 0, len(result.Spans))
	for _, span := range result.Spans {
		got = append(got, markdownBlockGoldenSpan{
			Kind:         span.Kind,
			Text:         span.Text,
			StartByte:    span.StartByte,
			EndByte:      span.EndByte,
			HeadingLevel: span.HeadingLevel,
			Atomic:       span.Atomic,
			Language:     span.Language,
		})
	}
	if !reflect.DeepEqual(got, golden.Spans) {
		t.Fatalf("Markdown block golden mismatch\n got: %#v\nwant: %#v", got, golden.Spans)
	}
}
