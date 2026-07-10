package chunking

func Paragraphs(source string, options ParagraphOptions) (*SegmentResult, error) {
	index, err := newSourceIndex(source)
	if err != nil {
		return nil, err
	}
	mode := options.BlankLines
	if mode == "" {
		mode = "trailing"
	}
	if mode != "trailing" && mode != "separate" && mode != "leading" {
		return nil, errUnknownOption("blankLines", mode)
	}
	boundaries := paragraphBoundaries(source)
	spans := make([]Span, 0, len(boundaries)*2+1)
	cursor := 0
	for _, boundary := range boundaries {
		bodyEnd, separatorEnd := boundary[0], boundary[1]
		switch mode {
		case "trailing":
			spans = appendNonEmpty(spans, index.span(cursor, separatorEnd, len(spans), "paragraph"))
			cursor = separatorEnd
		case "separate":
			spans = appendNonEmpty(spans, index.span(cursor, bodyEnd, len(spans), "paragraph"))
			spans = appendNonEmpty(spans, index.span(bodyEnd, separatorEnd, len(spans), "paragraphSeparator"))
			cursor = separatorEnd
		case "leading":
			spans = appendNonEmpty(spans, index.span(cursor, bodyEnd, len(spans), "paragraph"))
			cursor = bodyEnd
		}
	}
	spans = appendNonEmpty(spans, index.span(cursor, len(source), len(spans), "paragraph"))
	return segmentResult("paragraphs", source, spans, map[string]any{"blankLines": mode})
}

func paragraphBoundaries(source string) [][2]int {
	var boundaries [][2]int
	lineStart := 0
	inBlankRun := false
	blankStart := 0
	for lineStart < len(source) {
		lineEnd := lineStart
		for lineEnd < len(source) && source[lineEnd] != '\n' {
			lineEnd++
		}
		fullEnd := lineEnd
		if fullEnd < len(source) {
			fullEnd++
		}
		contentEnd := lineEnd
		if contentEnd > lineStart && source[contentEnd-1] == '\r' {
			contentEnd--
		}
		blank := true
		for i := lineStart; i < contentEnd; i++ {
			if source[i] != ' ' && source[i] != '\t' {
				blank = false
				break
			}
		}
		if blank {
			if !inBlankRun {
				blankStart = lineStart
				inBlankRun = true
			}
		} else {
			if inBlankRun && blankStart > 0 {
				boundaries = append(boundaries, [2]int{blankStart, lineStart})
			}
			inBlankRun = false
		}
		lineStart = fullEnd
	}
	if inBlankRun && blankStart > 0 {
		boundaries = append(boundaries, [2]int{blankStart, len(source)})
	}
	return boundaries
}

func appendNonEmpty(spans []Span, span Span) []Span {
	if span.StartByte == span.EndByte {
		return spans
	}
	span.Ordinal = len(spans)
	return append(spans, span)
}

func errUnknownOption(name, value string) error {
	return &optionError{Name: name, Value: value}
}

type optionError struct {
	Name  string
	Value string
}

func (e *optionError) Error() string {
	return "chunking: unknown option value: " + e.Name + "=" + e.Value
}
