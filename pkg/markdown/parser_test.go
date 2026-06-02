package markdown

import (
	"strings"
	"testing"
)

func TestParseBuildsGoBackedTree(t *testing.T) {
	root, err := Parse("# Hello\n\nA [link](https://example.com).\n\n```go\nfmt.Println(1)\n```\n")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if root.Type != "document" {
		t.Fatalf("root.Type = %q, want document", root.Type)
	}
	if len(root.Children) < 3 {
		t.Fatalf("root.Children len = %d, want at least 3", len(root.Children))
	}
	heading := root.Children[0]
	if heading.Type != "heading" || heading.Level != 1 {
		t.Fatalf("heading = (%q,%d), want heading level 1", heading.Type, heading.Level)
	}
	text, err := TextContent(heading)
	if err != nil {
		t.Fatalf("TextContent(heading) error = %v", err)
	}
	if text != "Hello" {
		t.Fatalf("heading text = %q, want Hello", text)
	}

	var link *MarkdownNode
	var code *MarkdownNode
	walkNodes(root, func(node *MarkdownNode) {
		if node.Type == "link" {
			link = node
		}
		if node.Type == "fencedCodeBlock" {
			code = node
		}
	})
	if link == nil || link.Destination != "https://example.com" {
		t.Fatalf("link = %#v, want destination", link)
	}
	if code == nil || code.Language != "go" || !strings.Contains(code.Text, "fmt.Println") {
		t.Fatalf("code = %#v, want go fenced code block", code)
	}
}

func TestParseCapturesGoldmarkEdgeFields(t *testing.T) {
	source := strings.Join([]string{
		"# Title",
		"",
		"![Alt *em* text](https://img.example/p.png \"Image Title\")",
		"",
		"```go title=\"demo\"",
		"fmt.Println(1)",
		"```",
		"",
		"    indented code",
		"",
		"<div>raw</div>",
		"",
		"Text with <span>inline</span>.",
		"",
	}, "\n")

	root, err := Parse(source)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	image := firstNode(root, "image")
	if image == nil {
		t.Fatalf("image node not found in %#v", root)
	}
	if image.Destination != "https://img.example/p.png" || image.Title != "Image Title" || image.Alt != "Alt em text" {
		t.Fatalf("image fields = destination %q title %q alt %q", image.Destination, image.Title, image.Alt)
	}
	if image.SourcePos != [2]int{3, 1} {
		t.Fatalf("image.SourcePos = %#v, want [3 1]", image.SourcePos)
	}

	fenced := firstNode(root, "fencedCodeBlock")
	if fenced == nil {
		t.Fatal("fencedCodeBlock node not found")
	}
	if fenced.Language != "go" || fenced.Info != "go title=\"demo\"" || !strings.Contains(fenced.Text, "fmt.Println(1)") {
		t.Fatalf("fenced fields = language %q info %q text %q", fenced.Language, fenced.Info, fenced.Text)
	}
	if fenced.SourcePos != [2]int{5, 1} {
		t.Fatalf("fenced.SourcePos = %#v, want [5 1]", fenced.SourcePos)
	}

	code := firstNode(root, "codeBlock")
	if code == nil || !strings.Contains(code.Text, "indented code") {
		t.Fatalf("codeBlock = %#v, want indented code text", code)
	}
	if code.SourcePos != [2]int{9, 5} {
		t.Fatalf("code.SourcePos = %#v, want [9 5]", code.SourcePos)
	}

	htmlBlock := firstNode(root, "htmlBlock")
	if htmlBlock == nil || !strings.Contains(htmlBlock.Raw, "<div>raw</div>") {
		t.Fatalf("htmlBlock = %#v, want raw div", htmlBlock)
	}

	rawHTML := firstNode(root, "rawHTML")
	if rawHTML == nil || rawHTML.Raw != "<span>" {
		t.Fatalf("rawHTML = %#v, want inline span opener", rawHTML)
	}
}

func TestRenderHTML(t *testing.T) {
	html, err := RenderHTML("# Hello\n")
	if err != nil {
		t.Fatalf("RenderHTML() error = %v", err)
	}
	if !strings.Contains(html, "<h1>Hello</h1>") {
		t.Fatalf("html = %q, want h1", html)
	}
}

func TestValidateNode(t *testing.T) {
	valid := ValidateNode(&MarkdownNode{Type: "heading", Level: 2})
	if !valid.Valid {
		t.Fatalf("valid.Valid = false, errors = %#v", valid.Errors)
	}

	invalid := ValidateNode(&MarkdownNode{Type: "heading", Level: 9})
	if invalid.Valid || len(invalid.Errors) == 0 {
		t.Fatalf("invalid = %#v, want errors", invalid)
	}
}

func firstNode(root *MarkdownNode, typ string) *MarkdownNode {
	var found *MarkdownNode
	walkNodes(root, func(node *MarkdownNode) {
		if found == nil && node.Type == typ {
			found = node
		}
	})
	return found
}

func walkNodes(node *MarkdownNode, fn func(*MarkdownNode)) {
	if node == nil {
		return
	}
	fn(node)
	for _, child := range node.Children {
		walkNodes(child, fn)
	}
}
