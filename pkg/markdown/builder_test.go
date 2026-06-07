package markdown

import (
	"strings"
	"testing"
)

func TestMarkdownBuilderBasicDocument(t *testing.T) {
	out, err := NewMarkdownBuilder().
		Title("Sprint report").
		Paragraph("Generated from structured data.").
		Heading(2, "Next steps").
		BulletList([]string{"Implement builder", "Add tests"}).
		RenderString()
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}
	want := "# Sprint report\n\nGenerated from structured data.\n\n## Next steps\n\n- Implement builder\n- Add tests\n"
	if out != want {
		t.Fatalf("out = %q, want %q", out, want)
	}
}

func TestMarkdownBuilderInlineEscapingAndRaw(t *testing.T) {
	inline := NewInlineFactory()
	out, err := NewMarkdownBuilder().
		Paragraph("Use *literal* brackets [x] and ", inline.Code("go test ./..."), ".").
		Raw("<details>raw markdown</details>").
		RenderString()
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}
	if !strings.Contains(out, `Use \*literal\* brackets \[x\] and `+"`go test ./...`") {
		t.Fatalf("escaped paragraph missing: %s", out)
	}
	if !strings.Contains(out, "<details>raw markdown</details>") {
		t.Fatalf("raw block missing: %s", out)
	}
}

func TestMarkdownBuilderTableAlignsAndEscapes(t *testing.T) {
	out, err := NewMarkdownBuilder().
		Table().
		Columns(map[string]any{"label": "Name", "align": "left"}, map[string]any{"label": "Score", "align": "right"}, "Note").
		Row("Ada", 42, "uses | pipe").
		Row("Linus", 100, "line 1\nline 2").
		End().
		RenderString()
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}
	for _, want := range []string{
		"| Name  | Score | Note             |",
		"| :---- | ----: | ---------------- |",
		`| Ada   | 42    | uses \| pipe     |`,
		"| Linus | 100   | line 1<br>line 2 |",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestMarkdownBuilderPreservesInlineWhitespace(t *testing.T) {
	inline := NewInlineFactory()
	out, err := NewMarkdownBuilder().
		Paragraph("Run ", inline.Code("cmd  --flag"), " exactly.").
		RenderString()
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}
	if !strings.Contains(out, "`cmd  --flag`") {
		t.Fatalf("code span whitespace was not preserved: %s", out)
	}
}

func TestMarkdownBuilderTableEscapesNestedInlineText(t *testing.T) {
	inline := NewInlineFactory()
	out, err := NewMarkdownBuilder().
		Table().
		Columns("Kind", "Value").
		Row("strong", inline.Strong("a|b\nc")).
		Row("em", inline.Em("x|y")).
		Row("link", inline.Link("docs|api", "https://example.com/a|b", "title|here")).
		End().
		RenderString()
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}
	for _, want := range []string{
		`**a\|b<br>c**`,
		`*x\|y*`,
		`[docs\|api](https://example.com/a\|b "title\|here")`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestMarkdownBuilderValidation(t *testing.T) {
	_, err := NewMarkdownBuilder().
		Heading(9, "bad").
		Table().Columns("A", "B").Row("only one").End().
		RenderString()
	if err == nil {
		t.Fatalf("RenderString() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "heading level") || !strings.Contains(err.Error(), "table row") {
		t.Fatalf("error = %v", err)
	}
}

func TestMarkdownBuilderListsChecklistCalloutAndCodeFence(t *testing.T) {
	out, err := NewMarkdownBuilder().
		OrderedList([]string{"first", "second"}, 3).
		Checklist([]map[string]any{{"text": "done", "checked": true}, {"text": "todo"}}).
		Callout("warning", "Careful", "line one\nline two").
		CodeBlock("go", "fmt.Println(```)").
		RenderString()
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}
	for _, want := range []string{"3. first", "4. second", "- [x] done", "- [ ] todo", "> [!WARNING] Careful", "> line two", "````go\nfmt.Println(```)\n````"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestMarkdownBuilderRenderHTML(t *testing.T) {
	html, err := NewMarkdownBuilder().Title("Hello").Paragraph("world").RenderHTML()
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}
	if !strings.Contains(html, "<h1>Hello</h1>") || !strings.Contains(html, "<p>world</p>") {
		t.Fatalf("html = %s", html)
	}
}
