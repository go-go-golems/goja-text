package chunking

// Lines partitions source at LF boundaries and recognizes CRLF as one terminator.
func Lines(source string, options LineOptions) (*SegmentResult, error) {
	index, err := newSourceIndex(source)
	if err != nil {
		return nil, err
	}
	spans := make([]Span, 0)
	start := 0
	for start < len(source) {
		end := start
		for end < len(source) && source[end] != '\n' {
			end++
		}
		terminatedEnd := end
		if end < len(source) {
			terminatedEnd++
		}
		if options.KeepTerminators || terminatedEnd == end {
			spans = append(spans, index.span(start, terminatedEnd, len(spans), "line"))
		} else {
			terminatorStart := end
			if terminatorStart > start && source[terminatorStart-1] == '\r' {
				terminatorStart--
			}
			if start < terminatorStart {
				spans = append(spans, index.span(start, terminatorStart, len(spans), "line"))
			}
			spans = append(spans, index.span(terminatorStart, terminatedEnd, len(spans), "lineTerminator"))
		}
		start = terminatedEnd
	}
	return segmentResult("lines", source, spans, map[string]any{"keepTerminators": options.KeepTerminators})
}
