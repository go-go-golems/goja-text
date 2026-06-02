package extract

import (
	"strings"
	"testing"
)

func TestMarkdownCodeBlocks(t *testing.T) {
	input := strings.Join([]string{
		"before",
		"```json meta",
		"{\"ok\": true}",
		"```",
		"text",
		"~~~yaml",
		"name: Alice",
		"~~~",
		"after",
	}, "\n")
	got, err := MarkdownCodeBlocks(input, nil)
	if err != nil {
		t.Fatalf("MarkdownCodeBlocks() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2: %#v", len(got), got)
	}
	if got[0].Format != "json" || got[0].Label != "json" || got[0].Info != "json meta" || strings.TrimSpace(got[0].Text) != "{\"ok\": true}" {
		t.Fatalf("first candidate = %#v", got[0])
	}
	if got[1].Format != "yaml" || got[1].Label != "yaml" || strings.TrimSpace(got[1].Text) != "name: Alice" {
		t.Fatalf("second candidate = %#v", got[1])
	}
}

func TestMarkdownCodeBlocksUnterminatedWithDiagnostics(t *testing.T) {
	opts, _ := NewExtractOptionsBuilder().IncludeDiagnostics(true).Build()
	got, err := MarkdownCodeBlocks("```json\n{\"ok\": true}\n", opts)
	if err != nil {
		t.Fatalf("MarkdownCodeBlocks() error = %v", err)
	}
	if len(got) != 1 || len(got[0].Diagnostics) == 0 || got[0].Confidence >= 0.95 {
		t.Fatalf("got = %#v, want diagnostic unterminated candidate", got)
	}
}

func TestXMLTagged(t *testing.T) {
	input := "before <json>{\"ok\": true}</json> middle <payload type=\"yaml\">name: Alice\n</payload> after"
	opts, _ := NewExtractOptionsBuilder().Tags("json", "payload").Build()
	got, err := XMLTagged(input, opts)
	if err != nil {
		t.Fatalf("XMLTagged() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2: %#v", len(got), got)
	}
	if got[0].Kind != "xmlTagged" || got[0].Format != "json" || got[0].Label != "json" || got[0].Text != "{\"ok\": true}" {
		t.Fatalf("first candidate = %#v", got[0])
	}
	if got[1].Label != "payload" || got[1].Info != "<payload type=\"yaml\">" {
		t.Fatalf("second candidate = %#v", got[1])
	}
}

func TestXMLTaggedMissingCloseDiagnostics(t *testing.T) {
	opts, _ := NewExtractOptionsBuilder().Tags("json").IncludeDiagnostics(true).Build()
	got, err := XMLTagged("<json>{\"ok\": true}", opts)
	if err != nil {
		t.Fatalf("XMLTagged() error = %v", err)
	}
	if len(got) != 1 || len(got[0].Diagnostics) == 0 {
		t.Fatalf("got = %#v, want missing close diagnostic", got)
	}
}

func TestFrontmatter(t *testing.T) {
	input := "---\ntitle: Demo\ntags:\n  - goja\n---\n# Body\n"
	got, err := Frontmatter(input, nil)
	if err != nil {
		t.Fatalf("Frontmatter() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].Kind != "frontmatter" || got[0].Format != "yaml" || !strings.Contains(got[0].Text, "title: Demo") || !strings.HasPrefix(got[0].Raw, "---\n") {
		t.Fatalf("candidate = %#v", got[0])
	}
}

func TestFrontmatterMissingClose(t *testing.T) {
	opts, _ := NewExtractOptionsBuilder().IncludeDiagnostics(true).Build()
	got, err := Frontmatter("---\ntitle: Demo\n# Body\n", opts)
	if err != nil {
		t.Fatalf("Frontmatter() error = %v", err)
	}
	if len(got) != 1 || len(got[0].Diagnostics) == 0 {
		t.Fatalf("got = %#v, want missing close diagnostic", got)
	}

	withoutDiag, err := Frontmatter("---\ntitle: Demo\n# Body\n", nil)
	if err != nil {
		t.Fatalf("Frontmatter() error = %v", err)
	}
	if len(withoutDiag) != 0 {
		t.Fatalf("withoutDiag len = %d, want 0", len(withoutDiag))
	}
}
