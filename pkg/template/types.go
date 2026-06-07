// Package template exposes Go text/template and html/template rendering helpers
// as Go-backed objects suitable for goja JavaScript runtimes.
package template

type Mode string

const (
	ModeText Mode = "text"
	ModeHTML Mode = "html"
)

const (
	MissingKeyDefault = "default"
	MissingKeyInvalid = "invalid"
	MissingKeyZero    = "zero"
	MissingKeyError   = "error"
)

type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

type TemplateConfig struct {
	Mode       Mode     `json:"mode"`
	Name       string   `json:"name"`
	FuncSets   []string `json:"funcSets,omitempty"`
	MissingKey string   `json:"missingKey"`
	LeftDelim  string   `json:"leftDelim,omitempty"`
	RightDelim string   `json:"rightDelim,omitempty"`
}

type RenderResult struct {
	Text         string `json:"text"`
	TemplateName string `json:"templateName"`
	Mode         Mode   `json:"mode"`
	Bytes        int    `json:"bytes"`
}

type TemplateInfo struct {
	Name    string `json:"name"`
	Defined bool   `json:"defined"`
	Mode    Mode   `json:"mode"`
}
