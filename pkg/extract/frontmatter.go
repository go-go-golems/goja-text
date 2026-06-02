package extract

import "strings"

// Frontmatter extracts leading YAML frontmatter delimited by --- lines.
func Frontmatter(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) {
	options = optionsOrDefault(options)
	li := newLineIndex(input)
	lines := splitSourceLines(input)
	if len(lines) == 0 || strings.TrimSpace(lines[0].Text) != "---" {
		return nil, nil
	}
	payloadStart := lineFullEnd(lines[0])
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i].Text) == "---" {
			payloadEnd := lines[i].StartByte
			endByte := lineFullEnd(lines[i])
			c := &ExtractionCandidate{
				Kind:       "frontmatter",
				Format:     "yaml",
				Text:       input[payloadStart:payloadEnd],
				Raw:        input[0:endByte],
				Wrapper:    "frontmatter",
				Label:      "yaml",
				Confidence: 0.98,
			}
			li.fillSpan(c, 0, endByte, payloadStart, payloadEnd)
			return filterCandidates([]*ExtractionCandidate{c}, options), nil
		}
	}
	if options.IncludeDiagnostics {
		c := &ExtractionCandidate{Kind: "frontmatter", Format: "yaml", Text: input[payloadStart:], Raw: input, Wrapper: "frontmatter", Label: "yaml", Confidence: 0.4, Diagnostics: []string{"missing closing frontmatter delimiter"}}
		li.fillSpan(c, 0, len(input), payloadStart, len(input))
		return filterCandidates([]*ExtractionCandidate{c}, options), nil
	}
	return nil, nil
}
