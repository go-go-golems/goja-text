package sanitize

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/modules"
	jsonsanitize "github.com/go-go-golems/sanitize/pkg/json"
	yamlsanitize "github.com/go-go-golems/sanitize/pkg/yaml"
)

type module struct{}

var _ modules.NativeModule = (*module)(nil)
var _ modules.TypeScriptDeclarer = (*module)(nil)

func (module) Name() string { return "sanitize" }

func (module) Doc() string {
	return `
The sanitize module provides YAML and JSON linting, sanitizing, parse-tree inspection, rule catalogs, and examples.

Functions are grouped by format:
  sanitize.yaml.options(): Create a Go-backed YAML options builder.
  sanitize.yaml.sanitize(input, config?): Sanitize YAML with defaults or a built config.
  sanitize.yaml.lint(input, config?): Lint YAML with defaults or a built config.
  sanitize.yaml.parseTree(input): Return tree-sitter tree text and parse errors.
  sanitize.yaml.rules(): Return YAML rule catalog.
  sanitize.yaml.examples(): Return YAML examples.

  sanitize.json.options(): Create a Go-backed JSON options builder.
  sanitize.json.sanitize(input, config?): Sanitize JSON with defaults or a built config.
  sanitize.json.lint(input, config?): Lint JSON with defaults or a built config.
  sanitize.json.parseTree(input): Return tree-sitter tree text and parse errors.
  sanitize.json.strictParse(input): Validate strict JSON.
  sanitize.json.rules(): Return JSON rule catalog.
  sanitize.json.examples(): Return JSON examples.

Go-backed config and result objects expose exported Go field and method names in JavaScript:
  builder.MaxIterations(5).OnlyRules("tab_indent").Build()
  result.Sanitized, result.Fixes, issue.Rule, issue.Description, config.UnknownPolicy
`
}

func (mod module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)

	yamlObj := vm.NewObject()
	mod.setYamlExports(yamlObj)
	if err := exports.Set("yaml", yamlObj); err != nil {
		panic(fmt.Errorf("sanitize: failed to set yaml namespace: %w", err))
	}

	jsonObj := vm.NewObject()
	mod.setJsonExports(jsonObj)
	if err := exports.Set("json", jsonObj); err != nil {
		panic(fmt.Errorf("sanitize: failed to set json namespace: %w", err))
	}
}

func (mod module) setYamlExports(obj *goja.Object) {
	modules.SetExport(obj, mod.Name(), "options", func() *YamlOptionsBuilder {
		return NewYamlOptionsBuilder()
	})

	modules.SetExport(obj, mod.Name(), "sanitize", func(input string, config *YamlConfig) (*yamlsanitize.Result, error) {
		result, err := yamlsanitize.SanitizeWithOptions(input, config.Options()...)
		if err != nil {
			return nil, fmt.Errorf("sanitize.yaml.sanitize: %w", err)
		}
		return &result, nil
	})

	modules.SetExport(obj, mod.Name(), "lint", func(input string, config *YamlConfig) ([]yamlsanitize.LintIssue, error) {
		issues, err := yamlsanitize.LintWithOptions(input, config.Options()...)
		if err != nil {
			return nil, fmt.Errorf("sanitize.yaml.lint: %w", err)
		}
		return issues, nil
	})

	modules.SetExport(obj, mod.Name(), "parseTree", func(input string) (*YamlParseTreeResult, error) {
		tree, errors, err := yamlsanitize.ParseTree(input)
		if err != nil {
			return nil, fmt.Errorf("sanitize.yaml.parseTree: %w", err)
		}
		return &YamlParseTreeResult{TreeText: tree, Errors: errors}, nil
	})

	modules.SetExport(obj, mod.Name(), "rules", func() []yamlsanitize.RuleSpec {
		return yamlsanitize.RuleCatalog()
	})

	modules.SetExport(obj, mod.Name(), "examples", func() []yamlsanitize.Example {
		return yamlsanitize.Examples
	})
}

func (mod module) setJsonExports(obj *goja.Object) {
	modules.SetExport(obj, mod.Name(), "options", func() *JsonOptionsBuilder {
		return NewJsonOptionsBuilder()
	})

	modules.SetExport(obj, mod.Name(), "sanitize", func(input string, config *JsonConfig) (*jsonsanitize.Result, error) {
		result, err := jsonsanitize.SanitizeWithOptions(input, config.Options()...)
		if err != nil {
			return nil, fmt.Errorf("sanitize.json.sanitize: %w", err)
		}
		return &result, nil
	})

	modules.SetExport(obj, mod.Name(), "lint", func(input string, config *JsonConfig) ([]jsonsanitize.LintIssue, error) {
		issues, err := jsonsanitize.LintWithOptions(input, config.Options()...)
		if err != nil {
			return nil, fmt.Errorf("sanitize.json.lint: %w", err)
		}
		return issues, nil
	})

	modules.SetExport(obj, mod.Name(), "parseTree", func(input string) (*JsonParseTreeResult, error) {
		tree, errors, err := jsonsanitize.ParseTree(input)
		if err != nil {
			return nil, fmt.Errorf("sanitize.json.parseTree: %w", err)
		}
		return &JsonParseTreeResult{TreeText: tree, Errors: errors}, nil
	})

	modules.SetExport(obj, mod.Name(), "strictParse", func(input string) StrictParseResult {
		if err := jsonsanitize.StrictParse(input); err != nil {
			return StrictParseResult{Valid: false, Error: err.Error()}
		}
		return StrictParseResult{Valid: true}
	})

	modules.SetExport(obj, mod.Name(), "rules", func() []jsonsanitize.RuleSpec {
		return jsonsanitize.RuleCatalog()
	})

	modules.SetExport(obj, mod.Name(), "examples", func() []jsonsanitize.Example {
		return jsonsanitize.Examples
	})
}

func init() {
	modules.Register(&module{})
}
