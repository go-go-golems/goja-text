package extract

import (
	"fmt"
	"sort"
	"strings"
)

// ExtractionCandidate describes one structured-data payload candidate found in text.
type ExtractionCandidate struct {
	Kind             string   `json:"kind"`
	Format           string   `json:"format"`
	Text             string   `json:"text"`
	Raw              string   `json:"raw"`
	Wrapper          string   `json:"wrapper"`
	Label            string   `json:"label,omitempty"`
	Info             string   `json:"info,omitempty"`
	StartByte        int      `json:"startByte"`
	EndByte          int      `json:"endByte"`
	StartRow         int      `json:"startRow"`
	StartCol         int      `json:"startCol"`
	EndRow           int      `json:"endRow"`
	EndCol           int      `json:"endCol"`
	PayloadStartByte int      `json:"payloadStartByte"`
	PayloadEndByte   int      `json:"payloadEndByte"`
	Confidence       float64  `json:"confidence"`
	Diagnostics      []string `json:"diagnostics,omitempty"`
}

// ExtractOptions configures extractor behavior.
type ExtractOptions struct {
	Formats            []string `json:"formats,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	Extractors         []string `json:"extractors,omitempty"`
	IncludeDiagnostics bool     `json:"includeDiagnostics"`
	InferFormat        bool     `json:"inferFormat"`
	MinConfidence      float64  `json:"minConfidence"`
	MaxCandidates      int      `json:"maxCandidates"`
}

// ExtractionValidationResult reports builder validation status.
type ExtractionValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// CandidateValidationResult reports format validation/sanitization for a candidate.
type CandidateValidationResult struct {
	Candidate *ExtractionCandidate `json:"candidate"`
	Valid     bool                 `json:"valid"`
	Format    string               `json:"format"`
	Sanitized string               `json:"sanitized,omitempty"`
	Errors    []string             `json:"errors,omitempty"`
	Fixes     any                  `json:"fixes,omitempty"`
	Issues    any                  `json:"issues,omitempty"`
}

// ExtractOptionsBuilder is a Go-backed builder for extraction options.
type ExtractOptionsBuilder struct {
	formats            []string
	tags               []string
	extractors         []string
	includeDiagnostics bool
	inferFormat        bool
	minConfidence      float64
	maxCandidates      int
	errors             []string
}

var defaultTags = []string{"json", "yaml", "xml", "data", "result", "answer", "tool_call", "arguments", "payload"}
var defaultExtractors = []string{"frontmatter", "markdowncodeblocks", "xmltagged", "rawstructured"}

// NewExtractOptionsBuilder returns a builder with default extraction behavior.
func NewExtractOptionsBuilder() *ExtractOptionsBuilder {
	return &ExtractOptionsBuilder{
		formats:            nil,
		tags:               append([]string(nil), defaultTags...),
		extractors:         append([]string(nil), defaultExtractors...),
		includeDiagnostics: false,
		inferFormat:        true,
		minConfidence:      0,
		maxCandidates:      0,
	}
}

func DefaultOptions() *ExtractOptions {
	cfg, _ := NewExtractOptionsBuilder().Build()
	return cfg
}

func (b *ExtractOptionsBuilder) Formats(formats ...string) *ExtractOptionsBuilder {
	b.formats = normalizeStrings(formats)
	return b
}

func (b *ExtractOptionsBuilder) Tags(tags ...string) *ExtractOptionsBuilder {
	b.tags = normalizeStrings(tags)
	return b
}

func (b *ExtractOptionsBuilder) Extractors(extractors ...string) *ExtractOptionsBuilder {
	b.extractors = normalizeStrings(extractors)
	return b
}

func (b *ExtractOptionsBuilder) IncludeDiagnostics(enabled bool) *ExtractOptionsBuilder {
	b.includeDiagnostics = enabled
	return b
}

func (b *ExtractOptionsBuilder) InferFormat(enabled bool) *ExtractOptionsBuilder {
	b.inferFormat = enabled
	return b
}

func (b *ExtractOptionsBuilder) MinConfidence(n float64) *ExtractOptionsBuilder {
	if n < 0 || n > 1 {
		b.errors = append(b.errors, "minConfidence must be between 0 and 1")
		return b
	}
	b.minConfidence = n
	return b
}

func (b *ExtractOptionsBuilder) MaxCandidates(n int) *ExtractOptionsBuilder {
	if n < 0 {
		b.errors = append(b.errors, "maxCandidates must be >= 0")
		return b
	}
	b.maxCandidates = n
	return b
}

func (b *ExtractOptionsBuilder) Validate() ExtractionValidationResult {
	errs := append([]string(nil), b.errors...)
	for _, f := range b.formats {
		if !knownFormat(f) {
			errs = append(errs, fmt.Sprintf("unknown format %q", f))
		}
	}
	for _, e := range b.extractors {
		if !knownExtractor(e) {
			errs = append(errs, fmt.Sprintf("unknown extractor %q", e))
		}
	}
	for _, tag := range b.tags {
		if tag == "" || strings.ContainsAny(tag, " <>/\t\n\r") {
			errs = append(errs, fmt.Sprintf("invalid tag %q", tag))
		}
	}
	return ExtractionValidationResult{Valid: len(errs) == 0, Errors: errs}
}

func (b *ExtractOptionsBuilder) Build() (*ExtractOptions, error) {
	result := b.Validate()
	if !result.Valid {
		return nil, fmt.Errorf("extract.options: %s", strings.Join(result.Errors, "; "))
	}
	return &ExtractOptions{
		Formats:            append([]string(nil), b.formats...),
		Tags:               append([]string(nil), b.tags...),
		Extractors:         append([]string(nil), b.extractors...),
		IncludeDiagnostics: b.includeDiagnostics,
		InferFormat:        b.inferFormat,
		MinConfidence:      b.minConfidence,
		MaxCandidates:      b.maxCandidates,
	}, nil
}

func optionsOrDefault(options *ExtractOptions) *ExtractOptions {
	if options == nil {
		return DefaultOptions()
	}
	if len(options.Tags) == 0 {
		options.Tags = append([]string(nil), defaultTags...)
	}
	if len(options.Extractors) == 0 {
		options.Extractors = append([]string(nil), defaultExtractors...)
	}
	return options
}

func (o *ExtractOptions) allowsFormat(format string) bool {
	if o == nil || len(o.Formats) == 0 || format == "unknown" {
		return true
	}
	for _, f := range o.Formats {
		if f == format {
			return true
		}
	}
	return false
}

func (o *ExtractOptions) allowsExtractor(extractor string) bool {
	if o == nil || len(o.Extractors) == 0 {
		return true
	}
	for _, e := range o.Extractors {
		if e == extractor {
			return true
		}
	}
	return false
}

func filterCandidates(candidates []*ExtractionCandidate, options *ExtractOptions) []*ExtractionCandidate {
	options = optionsOrDefault(options)
	ret := make([]*ExtractionCandidate, 0, len(candidates))
	for _, c := range candidates {
		if c == nil || !options.allowsFormat(c.Format) || c.Confidence < options.MinConfidence {
			continue
		}
		if !options.IncludeDiagnostics {
			c = cloneCandidate(c)
			c.Diagnostics = nil
		}
		ret = append(ret, c)
		if options.MaxCandidates > 0 && len(ret) >= options.MaxCandidates {
			break
		}
	}
	return ret
}

func cloneCandidate(c *ExtractionCandidate) *ExtractionCandidate {
	copy := *c
	copy.Diagnostics = append([]string(nil), c.Diagnostics...)
	return &copy
}

func normalizeStrings(values []string) []string {
	ret := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(strings.ToLower(value))
		if trimmed != "" {
			ret = append(ret, trimmed)
		}
	}
	return ret
}

func knownFormat(format string) bool {
	switch format {
	case "json", "yaml", "yml", "xml", "toml", "text", "unknown":
		return true
	default:
		return false
	}
}

func knownExtractor(extractor string) bool {
	switch extractor {
	case "frontmatter", "markdowncodeblocks", "xmltagged", "rawstructured":
		return true
	default:
		return false
	}
}

func sortCandidates(candidates []*ExtractionCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].StartByte == candidates[j].StartByte {
			return candidates[i].EndByte < candidates[j].EndByte
		}
		return candidates[i].StartByte < candidates[j].StartByte
	})
}
