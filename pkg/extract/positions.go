package extract

// SourcePosition is a zero-indexed row/column pair for a byte offset.
type SourcePosition struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// lineIndex maps byte offsets to zero-indexed row/column positions.
type lineIndex struct {
	source     string
	lineStarts []int
}

func newLineIndex(source string) lineIndex {
	starts := []int{0}
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			starts = append(starts, i+1)
		}
	}
	return lineIndex{source: source, lineStarts: starts}
}

func (li lineIndex) position(offset int) SourcePosition {
	if offset < 0 {
		offset = 0
	}
	if offset > len(li.source) {
		offset = len(li.source)
	}
	row := 0
	lo, hi := 0, len(li.lineStarts)
	for lo < hi {
		mid := (lo + hi) / 2
		if li.lineStarts[mid] <= offset {
			row = mid
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return SourcePosition{Row: row, Col: offset - li.lineStarts[row]}
}

func (li lineIndex) fillSpan(candidate *ExtractionCandidate, start, end, payloadStart, payloadEnd int) {
	candidate.StartByte = start
	candidate.EndByte = end
	candidate.PayloadStartByte = payloadStart
	candidate.PayloadEndByte = payloadEnd
	sp := li.position(start)
	ep := li.position(end)
	candidate.StartRow = sp.Row
	candidate.StartCol = sp.Col
	candidate.EndRow = ep.Row
	candidate.EndCol = ep.Col
}

type sourceLine struct {
	Text       string
	StartByte  int
	EndByte    int
	HasNewline bool
}

func splitSourceLines(source string) []sourceLine {
	if source == "" {
		return nil
	}
	var lines []sourceLine
	start := 0
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			lines = append(lines, sourceLine{Text: source[start:i], StartByte: start, EndByte: i, HasNewline: true})
			start = i + 1
		}
	}
	if start < len(source) {
		lines = append(lines, sourceLine{Text: source[start:], StartByte: start, EndByte: len(source)})
	} else if len(source) > 0 && source[len(source)-1] == '\n' {
		// No synthetic empty line after final newline; source spans remain easier to reason about.
	}
	return lines
}

func lineFullEnd(line sourceLine) int {
	if line.HasNewline {
		return line.EndByte + 1
	}
	return line.EndByte
}
