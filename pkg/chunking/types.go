package chunking

// Span is an exact, half-open range in the original UTF-8 source.
type Span struct {
	Ordinal int
	Kind    string
	Text    string

	StartByte int
	EndByte   int
	StartRune int
	EndRune   int

	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int

	Atomic       bool
	HeadingLevel int
	HeadingPath  []string
	Language     string
	Level        string
}

// Diagnostic is stable, machine-readable evidence about a chunking result.
type Diagnostic struct {
	Code      string
	Severity  string
	Message   string
	StartByte int
	EndByte   int
}

// StrategySpec identifies the deterministic operation that produced a result.
type StrategySpec struct {
	Name    string
	Version string
	Options map[string]any
}

// SegmentResult contains a lossless source partition.
type SegmentResult struct {
	Spec        StrategySpec
	SourceBytes int
	SourceRunes int
	Spans       []Span
	Diagnostics []Diagnostic
}

// PackedChunk contains one or more complete source spans.
type PackedChunk struct {
	Ordinal      int
	Text         string
	StartByte    int
	EndByte      int
	StartRune    int
	EndRune      int
	SpanOrdinals []int
	HeadingPath  []string
	Weight       int
	Oversized    bool
	Level        string
	Diagnostics  []Diagnostic
}

// PackResult contains greedy budgeted chunks and aggregate diagnostics.
type PackResult struct {
	Spec        StrategySpec
	Chunks      []PackedChunk
	Diagnostics []Diagnostic
}

type LineOptions struct {
	KeepTerminators bool
}

type ParagraphOptions struct {
	BlankLines string
}

type MarkdownBlockOptions struct {
	Atomic []string
}

type MarkdownSectionOptions struct {
	MaxHeadingLevel int
}

type OverlapOptions struct {
	Unit  string
	Value int
}

type PackOptions struct {
	MaxUnits  int
	Measure   string
	Overlap   OverlapOptions
	Oversized string
}

type WeightedSpan struct {
	Span   Span
	Weight int
}

type WeightedPackOptions struct {
	MaxWeight     int
	OverlapWeight int
	Oversized     string
}

type RecursiveOptions struct {
	MaxUnits  int
	Measure   string
	Levels    []string
	Overlap   OverlapOptions
	Oversized string
	Atomic    []string
}

func strategy(name string, options map[string]any) StrategySpec {
	return StrategySpec{Name: name, Version: "1", Options: options}
}
