package markdown

import (
	"unicode"
	"unicode/utf8"

	goldast "github.com/yuin/goldmark/ast"
)

// nodeSourceRange returns the half-open byte range occupied by a Goldmark node.
// Leaf text nodes use their exact source segment. Structural nodes extend from
// their source position to the next sibling (or ancestor sibling), excluding
// only trailing whitespace between blocks.
func nodeSourceRange(node goldast.Node, source []byte) (int, int) {
	if node == nil {
		return 0, 0
	}
	if node.Kind() == goldast.KindDocument {
		return 0, len(source)
	}

	start, ok := nodeSourceStart(node)
	if !ok {
		return 0, 0
	}

	switch value := node.(type) {
	case *goldast.Text:
		return clampRange(value.Segment.Start, value.Segment.Stop, len(source))
	case *goldast.RawHTML:
		if value.Segments != nil && value.Segments.Len() > 0 {
			first := value.Segments.At(0)
			last := value.Segments.At(value.Segments.Len() - 1)
			return clampRange(first.Start, last.Stop, len(source))
		}
	}

	end := len(source)
	for current := node; current != nil; current = current.Parent() {
		next := current.NextSibling()
		if next == nil {
			continue
		}
		if nextStart, found := nodeSourceStart(next); found {
			end = nextStart
			break
		}
	}
	start, end = clampRange(start, end, len(source))
	for end > start {
		r, size := lastRune(source[start:end])
		if size == 0 || !unicode.IsSpace(r) {
			break
		}
		end -= size
	}
	return start, end
}

func nodeSourceStart(node goldast.Node) (int, bool) {
	if node == nil {
		return 0, false
	}
	if node.Kind() == goldast.KindDocument {
		return 0, true
	}
	if pos := node.Pos(); pos >= 0 {
		return pos, true
	}
	if node.Type() == goldast.TypeBlock && node.Lines() != nil && node.Lines().Len() > 0 {
		return node.Lines().At(0).Start, true
	}
	if raw, ok := node.(*goldast.RawHTML); ok && raw.Segments != nil && raw.Segments.Len() > 0 {
		return raw.Segments.At(0).Start, true
	}
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if start, found := nodeSourceStart(child); found {
			return start, true
		}
	}
	return 0, false
}

func clampRange(start, end, length int) (int, int) {
	if start < 0 {
		start = 0
	}
	if start > length {
		start = length
	}
	if end < start {
		end = start
	}
	if end > length {
		end = length
	}
	return start, end
}

func lastRune(value []byte) (rune, int) {
	if len(value) == 0 {
		return 0, 0
	}
	return utf8.DecodeLastRune(value)
}
