package extract

import "strings"

// MarkdownCodeBlocks extracts fenced Markdown code blocks with exact wrapper and payload spans.
func MarkdownCodeBlocks(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) {
	options = optionsOrDefault(options)
	li := newLineIndex(input)
	lines := splitSourceLines(input)
	var candidates []*ExtractionCandidate

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		indent, fenceChar, fenceLen, info, ok := openingFence(line.Text)
		if !ok {
			continue
		}
		startByte := line.StartByte + indent
		payloadStart := lineFullEnd(line)
		label := ""
		if fields := strings.Fields(info); len(fields) > 0 {
			label = fields[0]
		}
		format := "unknown"
		if options.InferFormat {
			format = inferFormatFromLabel(info)
		}
		candidate := &ExtractionCandidate{
			Kind:       "markdownCodeBlock",
			Format:     format,
			Wrapper:    "markdownFence",
			Label:      strings.ToLower(label),
			Info:       info,
			Confidence: 0.95,
		}

		closed := false
		for j := i + 1; j < len(lines); j++ {
			if closingFence(lines[j].Text, fenceChar, fenceLen) {
				payloadEnd := lines[j].StartByte
				endByte := lineFullEnd(lines[j])
				candidate.Text = input[payloadStart:payloadEnd]
				candidate.Raw = input[startByte:endByte]
				li.fillSpan(candidate, startByte, endByte, payloadStart, payloadEnd)
				candidates = append(candidates, candidate)
				i = j
				closed = true
				break
			}
		}
		if !closed {
			payloadEnd := len(input)
			candidate.Text = input[payloadStart:payloadEnd]
			candidate.Raw = input[startByte:]
			candidate.Confidence = 0.6
			candidate.Diagnostics = append(candidate.Diagnostics, "unterminated markdown fence")
			li.fillSpan(candidate, startByte, len(input), payloadStart, payloadEnd)
			candidates = append(candidates, candidate)
			break
		}
	}
	return filterCandidates(candidates, options), nil
}

func openingFence(line string) (int, byte, int, string, bool) {
	indent := 0
	for indent < len(line) && line[indent] == ' ' {
		indent++
	}
	if indent > 3 || indent >= len(line) {
		return 0, 0, 0, "", false
	}
	ch := line[indent]
	if ch != '`' && ch != '~' {
		return 0, 0, 0, "", false
	}
	pos := indent
	for pos < len(line) && line[pos] == ch {
		pos++
	}
	fenceLen := pos - indent
	if fenceLen < 3 {
		return 0, 0, 0, "", false
	}
	return indent, ch, fenceLen, strings.TrimSpace(line[pos:]), true
}

func closingFence(line string, fenceChar byte, fenceLen int) bool {
	indent := 0
	for indent < len(line) && line[indent] == ' ' {
		indent++
	}
	if indent > 3 || indent >= len(line) {
		return false
	}
	pos := indent
	for pos < len(line) && line[pos] == fenceChar {
		pos++
	}
	if pos-indent < fenceLen {
		return false
	}
	return strings.TrimSpace(line[pos:]) == ""
}
