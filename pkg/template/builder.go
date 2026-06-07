package template

import (
	"fmt"
	"strings"
	texttemplate "text/template"

	"github.com/dop251/goja"
)

const defaultTemplateName = "template"

type TemplateBuilder struct {
	cfg         TemplateConfig
	customFuncs texttemplate.FuncMap
	vm          *goja.Runtime
	errors      []string
}

func NewTextBuilder() *TemplateBuilder { return newBuilder(ModeText) }
func NewHTMLBuilder() *TemplateBuilder { return newBuilder(ModeHTML) }

func newBuilder(mode Mode) *TemplateBuilder {
	return &TemplateBuilder{cfg: TemplateConfig{Mode: mode, Name: defaultTemplateName, FuncSets: append([]string(nil), defaultFuncSets...), MissingKey: MissingKeyError}, customFuncs: texttemplate.FuncMap{}}
}

func (b *TemplateBuilder) Name(name string) *TemplateBuilder {
	name = strings.TrimSpace(name)
	if name == "" {
		b.errors = append(b.errors, "name must not be empty")
		return b
	}
	b.cfg.Name = name
	return b
}

func (b *TemplateBuilder) Funcs(names ...string) *TemplateBuilder {
	b.cfg.FuncSets = normalizeFuncSets(names)
	return b
}

func (b *TemplateBuilder) MissingKey(policy string) *TemplateBuilder {
	policy = strings.TrimSpace(policy)
	b.cfg.MissingKey = policy
	return b
}

func (b *TemplateBuilder) Delims(left, right string) *TemplateBuilder {
	b.cfg.LeftDelim = left
	b.cfg.RightDelim = right
	return b
}

func (b *TemplateBuilder) Validate() ValidationResult {
	cfg := b.cfg
	errs := append([]string(nil), b.errors...)
	if cfg.Mode != ModeText && cfg.Mode != ModeHTML {
		errs = append(errs, fmt.Sprintf("unsupported mode %q", cfg.Mode))
	}
	if strings.TrimSpace(cfg.Name) == "" {
		errs = append(errs, "name must not be empty")
	}
	switch cfg.MissingKey {
	case MissingKeyDefault, MissingKeyInvalid, MissingKeyZero, MissingKeyError:
	default:
		errs = append(errs, fmt.Sprintf("unsupported missingKey policy %q", cfg.MissingKey))
	}
	if (cfg.LeftDelim == "") != (cfg.RightDelim == "") {
		errs = append(errs, "left and right delimiters must be set together")
	}
	if cfg.LeftDelim != "" && cfg.LeftDelim == cfg.RightDelim {
		errs = append(errs, "left and right delimiters must differ")
	}
	errs = append(errs, validateFuncSets(cfg.FuncSets)...)
	return ValidationResult{Valid: len(errs) == 0, Errors: errs}
}

func (b *TemplateBuilder) BuildConfig() (*TemplateConfig, error) {
	result := b.Validate()
	if !result.Valid {
		return nil, fmt.Errorf("template.%s.options: %s", b.cfg.Mode, joinErrors(result.Errors))
	}
	cfg := b.cfg
	cfg.FuncSets = append([]string(nil), b.cfg.FuncSets...)
	return &cfg, nil
}

func (b *TemplateBuilder) Parse(source string) (*TemplateSet, error) {
	return b.ParseNamed(b.cfg.Name, source)
}

func (b *TemplateBuilder) ParseNamed(name, source string) (*TemplateSet, error) {
	if strings.TrimSpace(name) == "" {
		b.errors = append(b.errors, "parse name must not be empty")
	}
	cfg, err := b.BuildConfig()
	if err != nil {
		return nil, err
	}
	return parseTemplateSet(*cfg, name, source, b.customFuncs)
}

func joinErrors(errs []string) string {
	return strings.Join(errs, "; ")
}
