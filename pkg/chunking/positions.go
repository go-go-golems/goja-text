package chunking

import "unicode/utf8"

type sourceIndex struct {
	source string
}

func newSourceIndex(source string) (sourceIndex, error) {
	if !utf8.ValidString(source) {
		return sourceIndex{}, invalidUTF8Error()
	}
	return sourceIndex{source: source}, nil
}

func (i sourceIndex) span(start, end int, ordinal int, kind string) Span {
	start = clamp(start, 0, len(i.source))
	end = clamp(end, start, len(i.source))
	startLine, startColumn := lineColumn(i.source, start)
	endLine, endColumn := lineColumn(i.source, end)
	return Span{
		Ordinal:     ordinal,
		Kind:        kind,
		Text:        i.source[start:end],
		StartByte:   start,
		EndByte:     end,
		StartRune:   utf8.RuneCountInString(i.source[:start]),
		EndRune:     utf8.RuneCountInString(i.source[:end]),
		StartLine:   startLine,
		StartColumn: startColumn,
		EndLine:     endLine,
		EndColumn:   endColumn,
	}
}

func lineColumn(source string, offset int) (int, int) {
	line, column := 1, 1
	offset = clamp(offset, 0, len(source))
	for _, r := range source[:offset] {
		if r == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return line, column
}

func clamp(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func translateSpan(span Span, parent Span, ordinal int) Span {
	span.Ordinal = ordinal
	span.StartByte += parent.StartByte
	span.EndByte += parent.StartByte
	span.StartRune += parent.StartRune
	span.EndRune += parent.StartRune
	span.StartLine += parent.StartLine - 1
	span.EndLine += parent.StartLine - 1
	if span.StartLine == parent.StartLine {
		span.StartColumn += parent.StartColumn - 1
	}
	if span.EndLine == parent.StartLine {
		span.EndColumn += parent.StartColumn - 1
	}
	return span
}
