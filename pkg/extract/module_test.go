package extract_test

import (
	"context"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/engine"
	_ "github.com/go-go-golems/goja-text/pkg/extract"
)

func TestRequireExtractMarkdownXMLRawFrontmatter(t *testing.T) {
	rt := newExtractRuntime(t)
	ret, err := rt.Owner.Call(context.Background(), "extract.core", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const extract = require("extract");
			const text = [
				"---",
				"title: Demo",
				"---",
				"",
				"~~~json meta",
				"{\"ok\": true}",
				"~~~",
				"",
				"<yaml>name: Alice\nage: 30\n</yaml>",
			].join("\n");
			const opts = extract.options().IncludeDiagnostics(true).Build();
			const blocks = extract.markdownCodeBlocks(text, opts);
			const tags = extract.xmlTagged(text, opts);
			const fm = extract.frontmatter(text, opts);
			const all = extract.all(text, opts);
			const raw = extract.rawStructured("{\"ok\": true}");
			({
				blockKind: blocks[0].Kind,
				blockFormat: blocks[0].Format,
				blockText: blocks[0].Text.trim(),
				blockLowercaseMissing: blocks[0].kind === undefined,
				tagFormat: tags[0].Format,
				frontmatterText: fm[0].Text.trim(),
				allCount: all.length,
				rawFormat: raw[0].Format,
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
	if got["blockKind"] != "markdownCodeBlock" || got["blockFormat"] != "json" || got["tagFormat"] != "yaml" || got["rawFormat"] != "json" || got["blockLowercaseMissing"] != true {
		t.Fatalf("unexpected extract result: %#v", got)
	}
	if text, _ := got["frontmatterText"].(string); !strings.Contains(text, "title: Demo") {
		t.Fatalf("frontmatterText = %q", text)
	}
}

func TestRequireExtractValidate(t *testing.T) {
	rt := newExtractRuntime(t)
	ret, err := rt.Owner.Call(context.Background(), "extract.validate", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const extract = require("extract");
			const candidate = extract.rawStructured("{'ok': True,}")[0];
			const validation = extract.validate(candidate);
			({
				format: validation.Format,
				valid: validation.Valid,
				sanitized: validation.Sanitized,
				fixCount: validation.Fixes.length,
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
	if got["format"] != "json" || got["valid"] != true {
		t.Fatalf("validation = %#v", got)
	}
	if sanitized, _ := got["sanitized"].(string); !strings.Contains(sanitized, "\"ok\"") {
		t.Fatalf("sanitized = %q", sanitized)
	}
}

func newExtractRuntime(t *testing.T) *engine.Runtime {
	t.Helper()
	factory, err := engine.NewBuilder().UseModuleMiddleware(engine.MiddlewareOnly("extract")).Build()
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
