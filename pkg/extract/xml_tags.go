package extract

import (
	"regexp"
	"strings"
)

// XMLTagged extracts simple same-name XML-like tag wrappers. It is not a full XML parser.
func XMLTagged(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) {
	options = optionsOrDefault(options)
	li := newLineIndex(input)
	var candidates []*ExtractionCandidate
	occupied := [][2]int{}

	for _, tag := range options.Tags {
		if tag == "" {
			continue
		}
		openRe := regexp.MustCompile(`(?s)<` + regexp.QuoteMeta(tag) + `(?:\s[^>]*)?>`)
		matches := openRe.FindAllStringIndex(input, -1)
		for _, match := range matches {
			openStart, openEnd := match[0], match[1]
			if spanOverlaps(openStart, openEnd, occupied) {
				continue
			}
			closeToken := "</" + tag + ">"
			relClose := strings.Index(input[openEnd:], closeToken)
			if relClose < 0 {
				if options.IncludeDiagnostics {
					c := &ExtractionCandidate{Kind: "xmlTagged", Format: inferFormatFromLabel(tag), Wrapper: "xmlTag", Label: tag, Raw: input[openStart:openEnd], Confidence: 0.3, Diagnostics: []string{"missing closing tag"}}
					li.fillSpan(c, openStart, openEnd, openEnd, openEnd)
					candidates = append(candidates, c)
				}
				continue
			}
			payloadStart := openEnd
			payloadEnd := openEnd + relClose
			endByte := payloadEnd + len(closeToken)
			if spanOverlaps(openStart, endByte, occupied) {
				continue
			}
			format := "unknown"
			if options.InferFormat {
				format = inferFormatFromLabel(tag)
				if format == "unknown" {
					format = inferFormatFromPayload(input[payloadStart:payloadEnd])
				}
			}
			c := &ExtractionCandidate{
				Kind:       "xmlTagged",
				Format:     format,
				Text:       input[payloadStart:payloadEnd],
				Raw:        input[openStart:endByte],
				Wrapper:    "xmlTag",
				Label:      tag,
				Info:       input[openStart:openEnd],
				Confidence: 0.9,
			}
			li.fillSpan(c, openStart, endByte, payloadStart, payloadEnd)
			candidates = append(candidates, c)
			occupied = append(occupied, [2]int{openStart, endByte})
		}
	}
	sortCandidates(candidates)
	return filterCandidates(candidates, options), nil
}

func spanOverlaps(start, end int, occupied [][2]int) bool {
	for _, span := range occupied {
		if start < span[1] && end > span[0] {
			return true
		}
	}
	return false
}
