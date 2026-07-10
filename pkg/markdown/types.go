package markdown

// MarkdownNode is a Go-backed Markdown AST node projected into JavaScript by goja.
//
// JavaScript callers access exported Go field names by default, for example:
// node.Type, node.Children, node.Level, node.Destination.
type MarkdownNode struct {
	Type        string          `json:"type"`
	Children    []*MarkdownNode `json:"children,omitempty"`
	Text        string          `json:"text,omitempty"`
	Level       int             `json:"level,omitempty"`
	Language    string          `json:"language,omitempty"`
	Destination string          `json:"destination,omitempty"`
	Title       string          `json:"title,omitempty"`
	Alt         string          `json:"alt,omitempty"`
	Ordered     bool            `json:"ordered,omitempty"`
	Start       int             `json:"start,omitempty"`
	Marker      string          `json:"marker,omitempty"`
	Info        string          `json:"info,omitempty"`
	Raw         string          `json:"raw,omitempty"`
	StartByte   int             `json:"startByte,omitempty"`
	EndByte     int             `json:"endByte,omitempty"`
	StartRune   int             `json:"startRune,omitempty"`
	EndRune     int             `json:"endRune,omitempty"`
	StartLine   int             `json:"startLine,omitempty"`
	StartColumn int             `json:"startColumn,omitempty"`
	EndLine     int             `json:"endLine,omitempty"`
	EndColumn   int             `json:"endColumn,omitempty"`
	SourcePos   [2]int          `json:"sourcePos,omitempty"`
}

// WalkContext describes the current position during MarkdownNode traversal.
type WalkContext struct {
	Parent *MarkdownNode `json:"parent,omitempty"`
	Depth  int           `json:"depth"`
	Index  int           `json:"index"`
	Path   []int         `json:"path"`
}

// ValidationResult reports runtime validation status for Markdown inputs or AST nodes.
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

type walkState struct {
	Stopped bool
}
