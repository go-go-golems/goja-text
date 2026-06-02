package sanitize

import (
	jsonsanitize "github.com/go-go-golems/sanitize/pkg/json"
	yamlsanitize "github.com/go-go-golems/sanitize/pkg/yaml"
)

// UnknownOptionPolicy controls how a builder handles unknown keys passed through FromObject.
type UnknownOptionPolicy string

const (
	UnknownOptionReject  UnknownOptionPolicy = "reject"
	UnknownOptionAllow   UnknownOptionPolicy = "allow"
	UnknownOptionCollect UnknownOptionPolicy = "collect"
)

// ValidationResult reports builder/config validation status.
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Unknown []string `json:"unknown,omitempty"`
}

// YamlConfig is an immutable Go-backed YAML sanitize configuration.
type YamlConfig struct {
	MaxIterations int                 `json:"maxIterations"`
	TabWidth      int                 `json:"tabWidth"`
	OnlyRules     []string            `json:"onlyRules,omitempty"`
	DisabledRules []string            `json:"disabledRules,omitempty"`
	UnknownPolicy UnknownOptionPolicy `json:"unknownPolicy"`
	Unknown       []string            `json:"unknown,omitempty"`
}

// Options converts the config to sanitize/pkg/yaml functional options.
func (c *YamlConfig) Options() []yamlsanitize.Option {
	if c == nil {
		return nil
	}
	opts := []yamlsanitize.Option{
		yamlsanitize.WithMaxIterations(c.MaxIterations),
		yamlsanitize.WithTabWidth(c.TabWidth),
	}
	if len(c.OnlyRules) > 0 {
		opts = append(opts, yamlsanitize.WithOnlyRules(c.OnlyRules...))
	}
	if len(c.DisabledRules) > 0 {
		opts = append(opts, yamlsanitize.WithDisabledRules(c.DisabledRules...))
	}
	return opts
}

// JsonConfig is an immutable Go-backed JSON sanitize configuration.
type JsonConfig struct {
	MaxIterations int                 `json:"maxIterations"`
	OnlyRules     []string            `json:"onlyRules,omitempty"`
	DisabledRules []string            `json:"disabledRules,omitempty"`
	UnknownPolicy UnknownOptionPolicy `json:"unknownPolicy"`
	Unknown       []string            `json:"unknown,omitempty"`
}

// Options converts the config to sanitize/pkg/json functional options.
func (c *JsonConfig) Options() []jsonsanitize.Option {
	if c == nil {
		return nil
	}
	opts := []jsonsanitize.Option{
		jsonsanitize.WithMaxIterations(c.MaxIterations),
	}
	if len(c.OnlyRules) > 0 {
		opts = append(opts, jsonsanitize.WithOnlyRules(c.OnlyRules...))
	}
	if len(c.DisabledRules) > 0 {
		opts = append(opts, jsonsanitize.WithDisabledRules(c.DisabledRules...))
	}
	return opts
}

// YamlParseTreeResult wraps yamlsanitize.ParseTree's multiple return values.
type YamlParseTreeResult struct {
	TreeText string                   `json:"treeText"`
	Errors   []yamlsanitize.ErrorNode `json:"errors"`
}

// JsonParseTreeResult wraps jsonsanitize.ParseTree's multiple return values.
type JsonParseTreeResult struct {
	TreeText string                   `json:"treeText"`
	Errors   []jsonsanitize.ErrorNode `json:"errors"`
}

// StrictParseResult reports JSON strict parser validation status.
type StrictParseResult struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}
