package chunking

import (
	"fmt"
	"strings"

	markdownpkg "github.com/go-go-golems/goja-text/pkg/markdown"
)

func MarkdownBlocks(source string, options MarkdownBlockOptions) (*SegmentResult, error) {
	index, err := newSourceIndex(source)
	if err != nil {
		return nil, err
	}
	root, err := markdownpkg.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("chunking.markdownBlocks: %w", err)
	}
	atomic := stringSet(options.Atomic)
	spans := make([]Span, 0, len(root.Children))
	for childIndex, child := range root.Children {
		start := 0
		if childIndex > 0 {
			start = child.StartByte
		}
		end := len(source)
		if childIndex+1 < len(root.Children) {
			end = root.Children[childIndex+1].StartByte
		}
		span := index.span(start, end, len(spans), child.Type)
		span.Atomic = atomic[child.Type]
		span.HeadingLevel = headingLevel(child)
		span.Language = child.Language
		spans = appendNonEmpty(spans, span)
	}
	if len(root.Children) == 0 && source != "" {
		spans = append(spans, index.span(0, len(source), 0, "text"))
	}
	return segmentResult("markdownBlocks", source, spans, map[string]any{"atomic": append([]string(nil), options.Atomic...)})
}

func MarkdownSections(source string, options MarkdownSectionOptions) (*SegmentResult, error) {
	index, err := newSourceIndex(source)
	if err != nil {
		return nil, err
	}
	maxLevel := options.MaxHeadingLevel
	if maxLevel == 0 {
		maxLevel = 6
	}
	if maxLevel < 1 || maxLevel > 6 {
		return nil, fmt.Errorf("chunking.markdownSections: maxHeadingLevel must be between 1 and 6")
	}
	root, err := markdownpkg.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("chunking.markdownSections: %w", err)
	}
	type sectionStart struct {
		offset int
		level  int
		title  string
		path   []string
	}
	starts := make([]sectionStart, 0)
	path := make([]string, 6)
	for _, child := range root.Children {
		if child.Type != "heading" || child.Level > maxLevel {
			continue
		}
		title := markdownText(child)
		path[child.Level-1] = title
		for i := child.Level; i < len(path); i++ {
			path[i] = ""
		}
		starts = append(starts, sectionStart{
			offset: child.StartByte,
			level:  child.Level,
			title:  title,
			path:   compactStrings(path[:child.Level]),
		})
	}
	spans := make([]Span, 0, len(starts)+1)
	if len(starts) == 0 {
		if source != "" {
			spans = append(spans, index.span(0, len(source), 0, "preamble"))
		}
	} else {
		if starts[0].offset > 0 {
			spans = append(spans, index.span(0, starts[0].offset, len(spans), "preamble"))
		}
		for i, start := range starts {
			end := len(source)
			if i+1 < len(starts) {
				end = starts[i+1].offset
			}
			span := index.span(start.offset, end, len(spans), "markdownSection")
			span.HeadingLevel = start.level
			span.HeadingPath = append([]string(nil), start.path...)
			spans = appendNonEmpty(spans, span)
		}
	}
	return segmentResult("markdownSections", source, spans, map[string]any{"maxHeadingLevel": maxLevel})
}

func headingLevel(node *markdownpkg.MarkdownNode) int {
	if node.Type == "heading" {
		return node.Level
	}
	return 0
}

func markdownText(node *markdownpkg.MarkdownNode) string {
	var b strings.Builder
	var walk func(*markdownpkg.MarkdownNode)
	walk = func(current *markdownpkg.MarkdownNode) {
		if current == nil {
			return
		}
		b.WriteString(current.Text)
		for _, child := range current.Children {
			walk(child)
		}
	}
	walk(node)
	return strings.TrimSpace(b.String())
}

func compactStrings(values []string) []string {
	ret := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			ret = append(ret, value)
		}
	}
	return ret
}

func stringSet(values []string) map[string]bool {
	ret := make(map[string]bool, len(values))
	for _, value := range values {
		ret[value] = true
	}
	return ret
}
