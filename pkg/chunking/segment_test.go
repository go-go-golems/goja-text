package chunking

import (
	"strings"
	"testing"
)

func TestLinesPreserveLFAndCRLFSource(t *testing.T) {
	for _, source := range []string{"", "one", "one\n", "one\r\ntwo\n", "é🙂\n"} {
		t.Run(source, func(t *testing.T) {
			result, err := Lines(source, LineOptions{KeepTerminators: true})
			if err != nil {
				t.Fatalf("Lines: %v", err)
			}
			assertPartition(t, source, result.Spans)
		})
	}
}

func TestLinesWithoutTerminatorsEmitsExplicitSeparatorSpans(t *testing.T) {
	result, err := Lines("one\r\ntwo", LineOptions{})
	if err != nil {
		t.Fatalf("Lines: %v", err)
	}
	wantKinds := []string{"line", "lineTerminator", "line"}
	if len(result.Spans) != len(wantKinds) {
		t.Fatalf("got %d spans, want %d", len(result.Spans), len(wantKinds))
	}
	for i, want := range wantKinds {
		if result.Spans[i].Kind != want {
			t.Errorf("span %d kind = %q, want %q", i, result.Spans[i].Kind, want)
		}
	}
	assertPartition(t, "one\r\ntwo", result.Spans)
}

func TestParagraphModesPreserveSeparators(t *testing.T) {
	source := "first\r\n\r\nsecond\n\n\nthird"
	for _, mode := range []string{"trailing", "separate", "leading"} {
		t.Run(mode, func(t *testing.T) {
			result, err := Paragraphs(source, ParagraphOptions{BlankLines: mode})
			if err != nil {
				t.Fatalf("Paragraphs: %v", err)
			}
			assertPartition(t, source, result.Spans)
		})
	}
	trailing, _ := Paragraphs("a\n\nb", ParagraphOptions{BlankLines: "trailing"})
	if got := trailing.Spans[0].Text; got != "a\n\n" {
		t.Fatalf("trailing separator = %q", got)
	}
}

func TestMarkdownBlocksPreserveSyntaxAndAtomicMetadata(t *testing.T) {
	source := "  # Title\n\nBody *with syntax*.\n\n```go\nfmt.Println(1)\n```\n"
	result, err := MarkdownBlocks(source, MarkdownBlockOptions{Atomic: []string{"fencedCodeBlock"}})
	if err != nil {
		t.Fatalf("MarkdownBlocks: %v", err)
	}
	assertPartition(t, source, result.Spans)
	if result.Spans[0].Text[:3] != "  #" {
		t.Fatalf("first block did not own leading source: %q", result.Spans[0].Text)
	}
	last := result.Spans[len(result.Spans)-1]
	if last.Kind != "fencedCodeBlock" || !last.Atomic || last.Language != "go" {
		t.Fatalf("fenced metadata = %#v", last)
	}
}

func TestMarkdownSectionsExposeHeadingPaths(t *testing.T) {
	source := "preamble\n\n# A\nBody\n\n## B\nNested\n\n# C\nEnd\n"
	result, err := MarkdownSections(source, MarkdownSectionOptions{})
	if err != nil {
		t.Fatalf("MarkdownSections: %v", err)
	}
	assertPartition(t, source, result.Spans)
	if result.Spans[0].Kind != "preamble" || len(result.Spans) != 4 {
		t.Fatalf("section kinds/count = %#v", result.Spans)
	}
	if got := strings.Join(result.Spans[2].HeadingPath, "/"); got != "A/B" {
		t.Fatalf("nested heading path = %q", got)
	}
}

func TestSegmentersRejectInvalidUTF8(t *testing.T) {
	_, err := Lines(string([]byte{0xff}), LineOptions{})
	if err == nil || !strings.Contains(err.Error(), "invalid_utf8") {
		t.Fatalf("error = %v", err)
	}
}

func assertPartition(t *testing.T, source string, spans []Span) {
	t.Helper()
	if err := ValidatePartition(source, spans); err != nil {
		t.Fatal(err)
	}
	var joined strings.Builder
	for _, span := range spans {
		joined.WriteString(span.Text)
	}
	if joined.String() != source {
		t.Fatalf("joined = %q, want %q", joined.String(), source)
	}
}

func FuzzLinesPreserveSource(f *testing.F) {
	for _, seed := range []string{"", "a", "a\n\n", "é🙂\r\nnext"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source string) {
		result, err := Lines(source, LineOptions{KeepTerminators: true})
		if err != nil {
			return
		}
		assertPartition(t, source, result.Spans)
	})
}
