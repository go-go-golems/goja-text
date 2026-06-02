package extract

// All runs enabled extractors and returns candidates sorted by source position.
func All(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) {
	options = optionsOrDefault(options)
	var out []*ExtractionCandidate
	if options.allowsExtractor("frontmatter") {
		candidates, err := Frontmatter(input, options)
		if err != nil {
			return nil, err
		}
		out = append(out, candidates...)
	}
	if options.allowsExtractor("markdowncodeblocks") {
		candidates, err := MarkdownCodeBlocks(input, options)
		if err != nil {
			return nil, err
		}
		out = append(out, candidates...)
	}
	if options.allowsExtractor("xmltagged") {
		candidates, err := XMLTagged(input, options)
		if err != nil {
			return nil, err
		}
		out = append(out, candidates...)
	}
	if options.allowsExtractor("rawstructured") {
		candidates, err := RawStructured(input, options)
		if err != nil {
			return nil, err
		}
		out = append(out, candidates...)
	}
	sortCandidates(out)
	return filterCandidates(out, options), nil
}
