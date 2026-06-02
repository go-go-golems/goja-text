package sanitize

import (
	fmt "fmt"
	"sort"
	"strings"

	jsonsanitize "github.com/go-go-golems/sanitize/pkg/json"
	yamlsanitize "github.com/go-go-golems/sanitize/pkg/yaml"
)

const (
	defaultMaxIterations = 10
	defaultYamlTabWidth  = 2
)

// YamlOptionsBuilder is a Go-backed builder for YAML sanitize configuration.
type YamlOptionsBuilder struct {
	maxIterations int
	tabWidth      int
	onlyRules     []string
	disabledRules []string
	unknownPolicy UnknownOptionPolicy
	unknown       []string
	errors        []string
}

// NewYamlOptionsBuilder returns a YAML builder with sanitize defaults.
func NewYamlOptionsBuilder() *YamlOptionsBuilder {
	return &YamlOptionsBuilder{maxIterations: defaultMaxIterations, tabWidth: defaultYamlTabWidth, unknownPolicy: UnknownOptionReject}
}

func (b *YamlOptionsBuilder) MaxIterations(n int) *YamlOptionsBuilder {
	if n <= 0 {
		b.errors = append(b.errors, "maxIterations must be > 0")
		return b
	}
	b.maxIterations = n
	return b
}

func (b *YamlOptionsBuilder) TabWidth(n int) *YamlOptionsBuilder {
	if n <= 0 {
		b.errors = append(b.errors, "tabWidth must be > 0")
		return b
	}
	b.tabWidth = n
	return b
}

func (b *YamlOptionsBuilder) OnlyRules(rules ...string) *YamlOptionsBuilder {
	b.onlyRules = normalizeRules(rules)
	return b
}

func (b *YamlOptionsBuilder) DisabledRules(rules ...string) *YamlOptionsBuilder {
	b.disabledRules = normalizeRules(rules)
	return b
}

func (b *YamlOptionsBuilder) RejectUnknownOptions() *YamlOptionsBuilder {
	b.unknownPolicy = UnknownOptionReject
	return b
}

func (b *YamlOptionsBuilder) AllowUnknownOptions() *YamlOptionsBuilder {
	b.unknownPolicy = UnknownOptionAllow
	return b
}

func (b *YamlOptionsBuilder) CollectUnknownOptions() *YamlOptionsBuilder {
	b.unknownPolicy = UnknownOptionCollect
	return b
}

// FromObject imports a plain JavaScript options object according to the builder's unknown-option policy.
func (b *YamlOptionsBuilder) FromObject(options map[string]any) *YamlOptionsBuilder {
	for key, value := range options {
		switch key {
		case "maxIterations":
			n, err := numberToPositiveInt(key, value)
			if err != nil {
				b.errors = append(b.errors, err.Error())
				continue
			}
			b.MaxIterations(n)
		case "tabWidth":
			n, err := numberToPositiveInt(key, value)
			if err != nil {
				b.errors = append(b.errors, err.Error())
				continue
			}
			b.TabWidth(n)
		case "onlyRules":
			rules, err := stringSliceOption(key, value)
			if err != nil {
				b.errors = append(b.errors, err.Error())
				continue
			}
			b.OnlyRules(rules...)
		case "disabledRules":
			rules, err := stringSliceOption(key, value)
			if err != nil {
				b.errors = append(b.errors, err.Error())
				continue
			}
			b.DisabledRules(rules...)
		default:
			b.handleUnknown(key)
		}
	}
	return b
}

func (b *YamlOptionsBuilder) Validate() ValidationResult {
	errs := append([]string(nil), b.errors...)
	errs = append(errs, validateCommon(b.maxIterations, b.onlyRules, b.disabledRules, yamlsanitize.ValidateRuleNames)...)
	if b.tabWidth <= 0 {
		errs = append(errs, "tabWidth must be > 0")
	}
	if b.unknownPolicy == UnknownOptionReject && len(b.unknown) > 0 {
		errs = append(errs, fmt.Sprintf("unknown option(s): %s", strings.Join(sortedCopy(b.unknown), ", ")))
	}
	return ValidationResult{Valid: len(errs) == 0, Errors: errs, Unknown: sortedCopy(b.unknown)}
}

func (b *YamlOptionsBuilder) Build() (*YamlConfig, error) {
	result := b.Validate()
	if !result.Valid {
		return nil, fmt.Errorf("sanitize.yaml.options: %s", strings.Join(result.Errors, "; "))
	}
	return &YamlConfig{
		MaxIterations: b.maxIterations,
		TabWidth:      b.tabWidth,
		OnlyRules:     append([]string(nil), b.onlyRules...),
		DisabledRules: append([]string(nil), b.disabledRules...),
		UnknownPolicy: b.unknownPolicy,
		Unknown:       sortedCopy(b.unknown),
	}, nil
}

func (b *YamlOptionsBuilder) handleUnknown(key string) {
	switch b.unknownPolicy {
	case UnknownOptionAllow:
		return
	case UnknownOptionCollect, UnknownOptionReject:
		b.unknown = append(b.unknown, key)
	default:
		b.errors = append(b.errors, fmt.Sprintf("unsupported unknown option policy %q", b.unknownPolicy))
	}
}

// JsonOptionsBuilder is a Go-backed builder for JSON sanitize configuration.
type JsonOptionsBuilder struct {
	maxIterations int
	onlyRules     []string
	disabledRules []string
	unknownPolicy UnknownOptionPolicy
	unknown       []string
	errors        []string
}

// NewJsonOptionsBuilder returns a JSON builder with sanitize defaults.
func NewJsonOptionsBuilder() *JsonOptionsBuilder {
	return &JsonOptionsBuilder{maxIterations: defaultMaxIterations, unknownPolicy: UnknownOptionReject}
}

func (b *JsonOptionsBuilder) MaxIterations(n int) *JsonOptionsBuilder {
	if n <= 0 {
		b.errors = append(b.errors, "maxIterations must be > 0")
		return b
	}
	b.maxIterations = n
	return b
}

func (b *JsonOptionsBuilder) OnlyRules(rules ...string) *JsonOptionsBuilder {
	b.onlyRules = normalizeRules(rules)
	return b
}

func (b *JsonOptionsBuilder) DisabledRules(rules ...string) *JsonOptionsBuilder {
	b.disabledRules = normalizeRules(rules)
	return b
}

func (b *JsonOptionsBuilder) RejectUnknownOptions() *JsonOptionsBuilder {
	b.unknownPolicy = UnknownOptionReject
	return b
}

func (b *JsonOptionsBuilder) AllowUnknownOptions() *JsonOptionsBuilder {
	b.unknownPolicy = UnknownOptionAllow
	return b
}

func (b *JsonOptionsBuilder) CollectUnknownOptions() *JsonOptionsBuilder {
	b.unknownPolicy = UnknownOptionCollect
	return b
}

// FromObject imports a plain JavaScript options object according to the builder's unknown-option policy.
func (b *JsonOptionsBuilder) FromObject(options map[string]any) *JsonOptionsBuilder {
	for key, value := range options {
		switch key {
		case "maxIterations":
			n, err := numberToPositiveInt(key, value)
			if err != nil {
				b.errors = append(b.errors, err.Error())
				continue
			}
			b.MaxIterations(n)
		case "onlyRules":
			rules, err := stringSliceOption(key, value)
			if err != nil {
				b.errors = append(b.errors, err.Error())
				continue
			}
			b.OnlyRules(rules...)
		case "disabledRules":
			rules, err := stringSliceOption(key, value)
			if err != nil {
				b.errors = append(b.errors, err.Error())
				continue
			}
			b.DisabledRules(rules...)
		default:
			b.handleUnknown(key)
		}
	}
	return b
}

func (b *JsonOptionsBuilder) Validate() ValidationResult {
	errs := append([]string(nil), b.errors...)
	errs = append(errs, validateCommon(b.maxIterations, b.onlyRules, b.disabledRules, jsonsanitize.ValidateRuleNames)...)
	if b.unknownPolicy == UnknownOptionReject && len(b.unknown) > 0 {
		errs = append(errs, fmt.Sprintf("unknown option(s): %s", strings.Join(sortedCopy(b.unknown), ", ")))
	}
	return ValidationResult{Valid: len(errs) == 0, Errors: errs, Unknown: sortedCopy(b.unknown)}
}

func (b *JsonOptionsBuilder) Build() (*JsonConfig, error) {
	result := b.Validate()
	if !result.Valid {
		return nil, fmt.Errorf("sanitize.json.options: %s", strings.Join(result.Errors, "; "))
	}
	return &JsonConfig{
		MaxIterations: b.maxIterations,
		OnlyRules:     append([]string(nil), b.onlyRules...),
		DisabledRules: append([]string(nil), b.disabledRules...),
		UnknownPolicy: b.unknownPolicy,
		Unknown:       sortedCopy(b.unknown),
	}, nil
}

func (b *JsonOptionsBuilder) handleUnknown(key string) {
	switch b.unknownPolicy {
	case UnknownOptionAllow:
		return
	case UnknownOptionCollect, UnknownOptionReject:
		b.unknown = append(b.unknown, key)
	default:
		b.errors = append(b.errors, fmt.Sprintf("unsupported unknown option policy %q", b.unknownPolicy))
	}
}

func normalizeRules(rules []string) []string {
	ret := make([]string, 0, len(rules))
	for _, rule := range rules {
		trimmed := strings.TrimSpace(rule)
		if trimmed != "" {
			ret = append(ret, trimmed)
		}
	}
	return ret
}

func validateCommon(maxIterations int, onlyRules, disabledRules []string, validateRules func(...string) error) []string {
	var errs []string
	if maxIterations <= 0 {
		errs = append(errs, "maxIterations must be > 0")
	}
	if err := validateRules(onlyRules...); err != nil {
		errs = append(errs, err.Error())
	}
	if err := validateRules(disabledRules...); err != nil {
		errs = append(errs, err.Error())
	}
	disabled := make(map[string]bool, len(disabledRules))
	for _, rule := range disabledRules {
		disabled[rule] = true
	}
	var overlap []string
	for _, rule := range onlyRules {
		if disabled[rule] {
			overlap = append(overlap, rule)
		}
	}
	if len(overlap) > 0 {
		errs = append(errs, fmt.Sprintf("rule(s) cannot be both enabled and disabled: %s", strings.Join(sortedCopy(overlap), ", ")))
	}
	return errs
}

func numberToPositiveInt(name string, value any) (int, error) {
	switch n := value.(type) {
	case int:
		if n <= 0 {
			return 0, fmt.Errorf("%s must be > 0", name)
		}
		return n, nil
	case int64:
		if n <= 0 {
			return 0, fmt.Errorf("%s must be > 0", name)
		}
		return int(n), nil
	case float64:
		if n <= 0 || n != float64(int(n)) {
			return 0, fmt.Errorf("%s must be a positive integer", name)
		}
		return int(n), nil
	default:
		return 0, fmt.Errorf("%s must be a number, got %T", name, value)
	}
}

func stringSliceOption(name string, value any) ([]string, error) {
	switch v := value.(type) {
	case []string:
		return normalizeRules(v), nil
	case []any:
		ret := make([]string, 0, len(v))
		for i, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("%s[%d] must be a string, got %T", name, i, item)
			}
			ret = append(ret, s)
		}
		return normalizeRules(ret), nil
	default:
		return nil, fmt.Errorf("%s must be an array of strings, got %T", name, value)
	}
}

func sortedCopy(values []string) []string {
	ret := append([]string(nil), values...)
	sort.Strings(ret)
	return ret
}
