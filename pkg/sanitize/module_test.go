package sanitize_test

import (
	"context"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/engine"
	_ "github.com/go-go-golems/goja-text/pkg/sanitize"
)

func TestRequireSanitizeYamlBuilderAndSanitize(t *testing.T) {
	rt := newSanitizeRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "sanitize.yaml.builder", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const sanitize = require("sanitize");
			const cfg = sanitize.yaml.options()
				.MaxIterations(3)
				.TabWidth(4)
				.OnlyRules("missing_space_after_colon")
				.Build();
			const result = sanitize.yaml.sanitize("name:Alice\n", cfg);
			({
				maxIterations: cfg.MaxIterations,
				tabWidth: cfg.TabWidth,
				policy: cfg.UnknownPolicy,
				sanitized: result.Sanitized,
				fixCount: result.Fixes.length,
				firstFixRule: result.Fixes[0] && result.Fixes[0].Rule,
				lowercaseMissing: result.sanitized === undefined,
			});
		`)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call error = %v", err)
	}

	got := ret.(map[string]any)
	if got["sanitized"] != "name: Alice\n" || got["firstFixRule"] != "missing_space_after_colon" || got["lowercaseMissing"] != true {
		t.Fatalf("unexpected result: %#v", got)
	}
	if got["maxIterations"] != int64(3) && got["maxIterations"] != int(3) && got["maxIterations"] != float64(3) {
		t.Fatalf("maxIterations = %#v, want 3", got["maxIterations"])
	}
	if got["tabWidth"] != int64(4) && got["tabWidth"] != int(4) && got["tabWidth"] != float64(4) {
		t.Fatalf("tabWidth = %#v, want 4", got["tabWidth"])
	}
}

func TestRequireSanitizeJsonBuilderAndStrictParse(t *testing.T) {
	rt := newSanitizeRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "sanitize.json.builder", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const sanitize = require("sanitize");
			const cfg = sanitize.json.options()
				.MaxIterations(4)
				.DisabledRules("duplicate_key")
				.Build();
			const result = sanitize.json.sanitize("~~~json\n{'ok': True,}\n~~~\n", cfg);
			const strictBad = sanitize.json.strictParse("{'ok': true}");
			const strictGood = sanitize.json.strictParse("{\"ok\": true}");
			({
				maxIterations: cfg.MaxIterations,
				sanitized: result.Sanitized,
				fixRules: result.Fixes.map(f => f.Rule),
				strictClean: result.StrictParseClean,
				strictBadValid: strictBad.Valid,
				strictBadError: strictBad.Error,
				strictGoodValid: strictGood.Valid,
			});
		`)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call error = %v", err)
	}

	got := ret.(map[string]any)
	if got["strictBadValid"] != false || got["strictGoodValid"] != true {
		t.Fatalf("strict parse result = %#v", got)
	}
	if errText, _ := got["strictBadError"].(string); errText == "" {
		t.Fatalf("strictBadError = %#v, want non-empty", got["strictBadError"])
	}
	if sanitized, _ := got["sanitized"].(string); !strings.Contains(sanitized, "\"ok\"") {
		t.Fatalf("sanitized = %q, want repaired JSON-ish output", sanitized)
	}
}

func TestRequireSanitizeBuilderUnknownOptionPolicies(t *testing.T) {
	rt := newSanitizeRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "sanitize.options.unknown", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const sanitize = require("sanitize");
			let rejectError = "";
			try {
				sanitize.yaml.options().FromObject({ typo: true }).Build();
			} catch (e) {
				rejectError = String(e);
			}
			const collected = sanitize.yaml.options()
				.CollectUnknownOptions()
				.FromObject({ typo: true, maxIterations: 2 })
				.Build();
			({
				rejectError,
				unknown0: collected.Unknown[0],
				unknownLength: collected.Unknown.length,
				maxIterations: collected.MaxIterations,
			});
		`)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call error = %v", err)
	}

	got := ret.(map[string]any)
	if rejectError, _ := got["rejectError"].(string); !strings.Contains(rejectError, "unknown option") {
		t.Fatalf("rejectError = %q, want unknown option error", rejectError)
	}
	if got["unknown0"] != "typo" {
		t.Fatalf("unknown0 = %#v, want typo", got["unknown0"])
	}
	if got["unknownLength"] != int64(1) && got["unknownLength"] != int(1) && got["unknownLength"] != float64(1) {
		t.Fatalf("unknownLength = %#v, want 1", got["unknownLength"])
	}
}

func TestRequireSanitizeRulesExamplesAndParseTree(t *testing.T) {
	rt := newSanitizeRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "sanitize.metadata", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const sanitize = require("sanitize");
			const yamlRules = sanitize.yaml.rules();
			const jsonRules = sanitize.json.rules();
			const yamlExamples = sanitize.yaml.examples();
			const jsonExamples = sanitize.json.examples();
			const tree = sanitize.yaml.parseTree("name:Alice\n");
			({
				yamlRule: yamlRules[0].Name,
				jsonRule: jsonRules[0].Name,
				jsonParseAwarePresent: jsonRules.some(r => r.ParseAware === true),
				yamlExampleCount: yamlExamples.length,
				jsonExampleCount: jsonExamples.length,
				treeTextPresent: tree.TreeText.length > 0,
				errorCount: tree.Errors.length,
			});
		`)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call error = %v", err)
	}

	got := ret.(map[string]any)
	if got["yamlRule"] == "" || got["jsonRule"] == "" || got["treeTextPresent"] != true {
		t.Fatalf("metadata result = %#v", got)
	}
}

func newSanitizeRuntime(t *testing.T) *engine.Runtime {
	t.Helper()
	factory, err := engine.NewRuntimeFactoryBuilder().UseModuleMiddleware(engine.MiddlewareOnly("sanitize")).Build()
	if err != nil {
		t.Fatalf("build runtime factory: %v", err)
	}
	rt, err := factory.NewRuntime(engine.WithStartupContext(context.Background()), engine.WithLifetimeContext(context.Background()))
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	t.Cleanup(func() { _ = rt.Close(context.Background()) })
	return rt
}
