package markdown

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

// Parse parses Markdown and returns a Go-backed MarkdownNode tree.
func Parse(input string) (*MarkdownNode, error) {
	source := []byte(input)
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(source))
	if doc == nil {
		return nil, fmt.Errorf("markdown.parse: parser returned nil document")
	}
	return ConvertAST(source, doc), nil
}

// RenderHTML renders Markdown input to HTML.
func RenderHTML(input string) (string, error) {
	var buf bytes.Buffer
	if err := goldmark.New().Convert([]byte(input), &buf); err != nil {
		return "", fmt.Errorf("markdown.renderHTML: %w", err)
	}
	return buf.String(), nil
}

// ValidateInput parses input and validates the resulting MarkdownNode tree.
func ValidateInput(input string) ValidationResult {
	node, err := Parse(input)
	if err != nil {
		return ValidationResult{Valid: false, Errors: []string{err.Error()}}
	}
	return ValidateNode(node)
}

// ValidateNode validates Go-backed MarkdownNode invariants.
func ValidateNode(root *MarkdownNode) ValidationResult {
	var errors []string
	validateNode(root, "root", &errors)
	return ValidationResult{Valid: len(errors) == 0, Errors: errors}
}

func validateNode(node *MarkdownNode, path string, errors *[]string) {
	if node == nil {
		*errors = append(*errors, path+": nil node")
		return
	}
	if node.Type == "" {
		*errors = append(*errors, path+": Type is required")
	}
	if node.Type == "heading" && (node.Level < 1 || node.Level > 6) {
		*errors = append(*errors, fmt.Sprintf("%s: heading Level must be 1..6, got %d", path, node.Level))
	}
	for i, child := range node.Children {
		validateNode(child, fmt.Sprintf("%s.Children[%d]", path, i), errors)
	}
}
