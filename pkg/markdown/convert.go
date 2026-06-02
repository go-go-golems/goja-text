package markdown

import (
	"fmt"
	"strings"

	goldast "github.com/yuin/goldmark/ast"
)

// ConvertAST converts a goldmark AST node into a Go-backed MarkdownNode tree.
func ConvertAST(source []byte, n goldast.Node) *MarkdownNode {
	if n == nil {
		return nil
	}

	sourcePos := sourcePosition(n, source)
	node := &MarkdownNode{
		Type:        nodeType(n),
		StartLine:   sourcePos[0],
		StartColumn: sourcePos[1],
		SourcePos:   sourcePos,
	}

	switch v := n.(type) {
	case *goldast.Heading:
		node.Level = v.Level
	case *goldast.Emphasis:
		node.Level = v.Level
	case *goldast.FencedCodeBlock:
		node.Language = string(v.Language(source))
		if v.Info != nil {
			node.Info = string(v.Info.Segment.Value(source))
		}
		node.Text = string(v.Lines().Value(source))
	case *goldast.CodeBlock:
		node.Text = string(v.Lines().Value(source))
	case *goldast.Text:
		node.Text = string(v.Value(source))
	case *goldast.String:
		node.Text = string(v.Value)
	case *goldast.Link:
		node.Destination = string(v.Destination)
		node.Title = string(v.Title)
	case *goldast.Image:
		node.Destination = string(v.Destination)
		node.Title = string(v.Title)
		node.Alt = plainTextFromGoldmarkChildren(source, v)
	case *goldast.List:
		node.Ordered = v.IsOrdered()
		if v.IsOrdered() {
			node.Start = v.Start
		}
		if v.Marker != 0 {
			node.Marker = string(v.Marker)
		}
	case *goldast.HTMLBlock:
		node.Raw = htmlBlockText(v, source)
	case *goldast.RawHTML:
		node.Raw = string(v.Segments.Value(source))
	case *goldast.AutoLink:
		node.Text = string(v.Label(source))
	}

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		node.Children = append(node.Children, ConvertAST(source, child))
	}

	return node
}

func nodeType(n goldast.Node) string {
	switch n.Kind() {
	case goldast.KindDocument:
		return "document"
	case goldast.KindHeading:
		return "heading"
	case goldast.KindParagraph:
		return "paragraph"
	case goldast.KindText:
		return "text"
	case goldast.KindString:
		return "string"
	case goldast.KindEmphasis:
		return "emphasis"
	case goldast.KindCodeSpan:
		return "codeSpan"
	case goldast.KindCodeBlock:
		return "codeBlock"
	case goldast.KindFencedCodeBlock:
		return "fencedCodeBlock"
	case goldast.KindLink:
		return "link"
	case goldast.KindImage:
		return "image"
	case goldast.KindList:
		return "list"
	case goldast.KindListItem:
		return "listItem"
	case goldast.KindBlockquote:
		return "blockquote"
	case goldast.KindThematicBreak:
		return "thematicBreak"
	case goldast.KindHTMLBlock:
		return "htmlBlock"
	case goldast.KindRawHTML:
		return "rawHTML"
	case goldast.KindTextBlock:
		return "textBlock"
	case goldast.KindAutoLink:
		return "autoLink"
	default:
		return strings.TrimSpace(n.Kind().String())
	}
}

func sourcePosition(n goldast.Node, source []byte) [2]int {
	pos := n.Pos()
	if pos < 0 {
		return [2]int{}
	}
	line, col := byteOffsetLineColumn(source, pos)
	return [2]int{line, col}
}

func htmlBlockText(node *goldast.HTMLBlock, source []byte) string {
	ret := node.Lines().Value(source)
	if node.HasClosure() {
		ret = append(ret, node.ClosureLine.Value(source)...)
	}
	return string(ret)
}

func byteOffsetLineColumn(source []byte, offset int) (int, int) {
	line, column := 1, 1
	if offset <= 0 {
		return line, column
	}
	if offset > len(source) {
		offset = len(source)
	}
	for _, b := range source[:offset] {
		if b == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return line, column
}

func plainTextFromGoldmarkChildren(source []byte, n goldast.Node) string {
	var b strings.Builder
	var visit func(goldast.Node)
	visit = func(node goldast.Node) {
		if node == nil {
			return
		}
		switch v := node.(type) {
		case *goldast.Text:
			b.Write(v.Value(source))
		case *goldast.String:
			b.Write(v.Value)
		}
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			visit(child)
		}
	}
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		visit(child)
	}
	return b.String()
}

// TextContent returns the concatenated textual content of node and its descendants.
func TextContent(node *MarkdownNode) (string, error) {
	if node == nil {
		return "", fmt.Errorf("markdown.textContent: node must be a MarkdownNode")
	}
	var b strings.Builder
	collectText(node, &b)
	return b.String(), nil
}

func collectText(node *MarkdownNode, b *strings.Builder) {
	if node == nil {
		return
	}
	if node.Text != "" {
		b.WriteString(node.Text)
	}
	for _, child := range node.Children {
		collectText(child, b)
	}
}
