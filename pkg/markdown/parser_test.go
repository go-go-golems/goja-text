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

func walkNodes(node *MarkdownNode, fn func(*MarkdownNode)) {
	if node == nil {
		return
	}
	fn(node)
	for _, child := range node.Children {
		walkNodes(child, fn)
	}
}
