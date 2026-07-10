package chunking

import (
	"fmt"
	"unicode/utf8"
)

func Recursive(source string, options RecursiveOptions) (*PackResult, error) {
	index, err := newSourceIndex(source)
	if err != nil {
		return nil, err
	}
	if options.MaxUnits <= 0 {
		return nil, fmt.Errorf("chunking.recursive: maxUnits must be greater than zero")
	}
	if options.Measure == "" {
		options.Measure = "runes"
	}
	if options.Oversized == "" {
		options.Oversized = "allow"
	}
	if len(options.Levels) == 0 {
		options.Levels = []string{"markdownSections", "markdownBlocks", "paragraphs", "lines", "runes"}
	}
	if _, err := measureText("", options.Measure); err != nil {
		return nil, err
	}
	root := index.span(0, len(source), 0, "source")
	leafSpans, err := refineSpan(root, options, 0)
	if err != nil {
		return nil, err
	}
	for i := range leafSpans {
		leafSpans[i].Ordinal = i
	}
	result, err := Pack(leafSpans, PackOptions{
		MaxUnits:  options.MaxUnits,
		Measure:   options.Measure,
		Overlap:   options.Overlap,
		Oversized: options.Oversized,
	})
	if err != nil {
		return nil, err
	}
	result.Spec = strategy("recursive", map[string]any{
		"maxUnits":  options.MaxUnits,
		"measure":   options.Measure,
		"levels":    append([]string(nil), options.Levels...),
		"overlap":   map[string]any{"unit": "spans", "value": options.Overlap.Value},
		"oversized": options.Oversized,
	})
	return result, nil
}

func refineSpan(parent Span, options RecursiveOptions, levelIndex int) ([]Span, error) {
	weight, err := measureText(parent.Text, options.Measure)
	if err != nil {
		return nil, err
	}
	if weight <= options.MaxUnits || parent.Atomic {
		return []Span{parent}, nil
	}
	if levelIndex >= len(options.Levels) {
		return []Span{parent}, nil
	}
	level := options.Levels[levelIndex]
	var result *SegmentResult
	switch level {
	case "markdownSections":
		result, err = MarkdownSections(parent.Text, MarkdownSectionOptions{})
	case "markdownBlocks":
		result, err = MarkdownBlocks(parent.Text, MarkdownBlockOptions{Atomic: options.Atomic})
	case "paragraphs":
		result, err = Paragraphs(parent.Text, ParagraphOptions{BlankLines: "trailing"})
	case "lines":
		result, err = Lines(parent.Text, LineOptions{KeepTerminators: true})
	case "runes":
		result, err = runeWindows(parent.Text, options.MaxUnits, options.Measure)
	default:
		return nil, fmt.Errorf("chunking: unknown_recursive_level: %q", level)
	}
	if err != nil {
		return nil, err
	}
	translated := make([]Span, 0, len(result.Spans))
	for _, child := range result.Spans {
		child.Level = level
		translated = append(translated, translateSpan(child, parent, len(translated)))
	}
	if len(translated) == 1 && translated[0].StartByte == parent.StartByte && translated[0].EndByte == parent.EndByte {
		translated[0].Atomic = parent.Atomic || translated[0].Atomic
		return refineSpan(translated[0], options, levelIndex+1)
	}
	leaves := make([]Span, 0, len(translated))
	for _, child := range translated {
		refined, refineErr := refineSpan(child, options, levelIndex+1)
		if refineErr != nil {
			return nil, refineErr
		}
		leaves = append(leaves, refined...)
	}
	return leaves, nil
}

func runeWindows(source string, maxUnits int, measure string) (*SegmentResult, error) {
	if measure == "words" {
		return Paragraphs(source, ParagraphOptions{BlankLines: "trailing"})
	}
	index, err := newSourceIndex(source)
	if err != nil {
		return nil, err
	}
	spans := make([]Span, 0)
	start := 0
	units := 0
	for offset := 0; offset < len(source); {
		_, size := utf8.DecodeRuneInString(source[offset:])
		unit := 1
		if measure == "bytes" {
			unit = size
		}
		if units > 0 && units+unit > maxUnits {
			spans = append(spans, index.span(start, offset, len(spans), "runeWindow"))
			start, units = offset, 0
		}
		units += unit
		offset += size
	}
	spans = appendNonEmpty(spans, index.span(start, len(source), len(spans), "runeWindow"))
	return segmentResult("runes", source, spans, map[string]any{"maxUnits": maxUnits, "measure": measure})
}
