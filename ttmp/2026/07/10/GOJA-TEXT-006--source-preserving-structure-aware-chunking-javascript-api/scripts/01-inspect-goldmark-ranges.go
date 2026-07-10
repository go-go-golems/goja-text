package main

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const sample = `# Heading

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

func main() {
	source := []byte(sample)
	doc := goldmark.New().Parser().Parse(text.NewReader(source))
	walk(source, doc, 0)
}

func walk(source []byte, node ast.Node, depth int) {
	lineRanges := make([]string, 0)
	if node.Type() == ast.TypeBlock && node.Lines() != nil {
		for i := 0; i < node.Lines().Len(); i++ {
			segment := node.Lines().At(i)
			lineRanges = append(lineRanges, fmt.Sprintf("%d:%d", segment.Start, segment.Stop))
		}
	}
	fmt.Printf(
		"%s%-20s type=%v pos=%d lines=[%s]\n",
		strings.Repeat("  ", depth),
		node.Kind().String(),
		node.Type(),
		node.Pos(),
		strings.Join(lineRanges, ","),
	)
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		walk(source, child, depth+1)
	}
}
