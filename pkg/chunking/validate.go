package chunking

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func invalidUTF8Error() error {
	return fmt.Errorf("chunking: invalid_utf8: source must be valid UTF-8")
}

func validatePackingSpans(spans []Span) error {
	previousEndByte := 0
	previousEndRune := 0
	for i, span := range spans {
		if span.Ordinal != i {
			return fmt.Errorf("chunking: invalid_range: span %d has ordinal %d", i, span.Ordinal)
		}
		if !utf8.ValidString(span.Text) {
			return fmt.Errorf("chunking: invalid_utf8: span %d text must be valid UTF-8", i)
		}
		if span.StartByte != previousEndByte || span.EndByte < span.StartByte || span.EndByte-span.StartByte != len(span.Text) {
			return fmt.Errorf("chunking: invalid_range: span %d byte range [%d,%d) does not match its text or previous span", i, span.StartByte, span.EndByte)
		}
		if span.StartRune != previousEndRune || span.EndRune < span.StartRune || span.EndRune-span.StartRune != utf8.RuneCountInString(span.Text) {
			return fmt.Errorf("chunking: invalid_range: span %d rune range [%d,%d) does not match its text or previous span", i, span.StartRune, span.EndRune)
		}
		previousEndByte = span.EndByte
		previousEndRune = span.EndRune
	}
	return nil
}

// ValidatePartition verifies that spans form a gapless, exact source partition.
func ValidatePartition(source string, spans []Span) error {
	position := 0
	var joined strings.Builder
	for ordinal, span := range spans {
		if span.Ordinal != ordinal {
			return fmt.Errorf("chunking: invalid_range: span ordinal %d is %d", ordinal, span.Ordinal)
		}
		if span.StartByte != position || span.EndByte < span.StartByte || span.EndByte > len(source) {
			return fmt.Errorf("chunking: invalid_range: span %d has range [%d,%d), expected start %d", ordinal, span.StartByte, span.EndByte, position)
		}
		if source[span.StartByte:span.EndByte] != span.Text {
			return fmt.Errorf("chunking: source_range_mismatch: span %d text does not match source", ordinal)
		}
		joined.WriteString(span.Text)
		position = span.EndByte
	}
	if position != len(source) || joined.String() != source {
		return fmt.Errorf("chunking: source_range_mismatch: spans cover %d of %d bytes", position, len(source))
	}
	return nil
}

func segmentResult(name string, source string, spans []Span, options map[string]any) (*SegmentResult, error) {
	if err := ValidatePartition(source, spans); err != nil {
		return nil, err
	}
	return &SegmentResult{
		Spec:        strategy(name, options),
		SourceBytes: len(source),
		SourceRunes: len([]rune(source)),
		Spans:       spans,
		Diagnostics: []Diagnostic{},
	}, nil
}
