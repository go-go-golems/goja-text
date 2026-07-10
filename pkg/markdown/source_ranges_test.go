package markdown

import (
	"testing"
	"unicode/utf8"
)

func TestParseExposesExactStructuralSourceRanges(t *testing.T) {
	source := `# Heading

Paragraph with **strong** text.

- first
- second

~~~go title="demo"
fmt.Println("hello")
~~~

> quoted

<div>
html
</div>

---
`
	root, err := Parse(source)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if root.StartByte != 0 || root.EndByte != len(source) {
		t.Fatalf("document range = [%d,%d), want [0,%d)", root.StartByte, root.EndByte, len(source))
	}
	want := []string{
		"# Heading",
		"Paragraph with **strong** text.",
		"- first\n- second",
		"~~~go title=\"demo\"\nfmt.Println(\"hello\")\n~~~",
		"> quoted",
		"<div>\nhtml\n</div>",
		"---",
	}
	if len(root.Children) != len(want) {
		t.Fatalf("top-level blocks = %d, want %d", len(root.Children), len(want))
	}
	for index, node := range root.Children {
		if got := source[node.StartByte:node.EndByte]; got != want[index] {
			t.Fatalf("block %d (%s) source = %q, want %q", index, node.Type, got, want[index])
		}
		if node.StartRune != utf8.RuneCountInString(source[:node.StartByte]) || node.EndRune != utf8.RuneCountInString(source[:node.EndByte]) {
			t.Fatalf("block %d rune range = [%d,%d), inconsistent with source", index, node.StartRune, node.EndRune)
		}
	}
}

func TestParseSourceRangesHandleUnicodeAndInlineSyntax(t *testing.T) {
	source := "# Hé🙂\n\nA **bold** [link](https://example.com).\n"
	root, err := Parse(source)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	heading := root.Children[0]
	if got := source[heading.StartByte:heading.EndByte]; got != "# Hé🙂" {
		t.Fatalf("heading source = %q", got)
	}
	if heading.EndRune-heading.StartRune != utf8.RuneCountInString("# Hé🙂") {
		t.Fatalf("heading rune length = %d", heading.EndRune-heading.StartRune)
	}
	paragraph := root.Children[1]
	emphasis := paragraph.Children[1]
	if got := source[emphasis.StartByte:emphasis.EndByte]; got != "**bold**" {
		t.Fatalf("emphasis source = %q, want %q", got, "**bold**")
	}
	link := paragraph.Children[3]
	if got := source[link.StartByte:link.EndByte]; got != "[link](https://example.com)" {
		t.Fatalf("link source = %q", got)
	}
}

func TestParseSourceRangesExposeOneBasedEndPosition(t *testing.T) {
	root, err := Parse("# H\n\nBody")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	heading := root.Children[0]
	if heading.StartLine != 1 || heading.StartColumn != 1 || heading.EndLine != 1 || heading.EndColumn != 4 {
		t.Fatalf("heading positions = %d:%d-%d:%d, want 1:1-1:4", heading.StartLine, heading.StartColumn, heading.EndLine, heading.EndColumn)
	}
	if root.EndLine != 3 || root.EndColumn != 5 {
		t.Fatalf("document end = %d:%d, want 3:5", root.EndLine, root.EndColumn)
	}
}
