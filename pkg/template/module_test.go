package template_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/pkg/engine"
	_ "github.com/go-go-golems/goja-text/pkg/template"
)

func TestRequireTemplateTextBuilder(t *testing.T) {
	rt := newTemplateRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "template.text.builder", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const template = require("template");
			const result = template.text()
				.Name("greeting")
				.Funcs("sprig", "glazed")
				.Parse("Hello {{ .Name | upper }}")
				.Render({ Name: "intern" });
			({ text: result.Text, name: result.TemplateName, mode: result.Mode, bytes: result.Bytes });
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
	if got["text"] != "Hello INTERN" || got["name"] != "greeting" || fmt.Sprint(got["mode"]) != "text" {
		t.Fatalf("unexpected result: %#v", got)
	}
	if got["bytes"] == int64(0) || got["bytes"] == float64(0) {
		t.Fatalf("bytes not set: %#v", got)
	}
}

func TestRequireTemplateHTMLBuilderEscaping(t *testing.T) {
	rt := newTemplateRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "template.html.builder", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const template = require("template");
			template.html()
				.Parse('<p>{{ .Name }}</p><a href="{{ .URL }}">open</a>')
				.Render({ Name: '<Ada>', URL: 'javascript:alert(1)' })
				.Text;
		`)
		if runErr != nil {
			return nil, runErr
		}
		return value.Export(), nil
	})
	if err != nil {
		t.Fatalf("runtime call error = %v", err)
	}
	out, ok := ret.(string)
	if !ok {
		t.Fatalf("ret = %T, want string", ret)
	}
	if !strings.Contains(out, "&lt;Ada&gt;") || strings.Contains(out, "javascript:alert") {
		t.Fatalf("html output was not escaped: %s", out)
	}
}

func TestRequireTemplateConvenienceAndNamedTemplates(t *testing.T) {
	rt := newTemplateRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "template.named", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const template = require("template");
			const one = template.renderText("{{ .Name | trim }}", { Name: " Ada " });
			const set = template.text().Name("report").Parse('{{ define "body" }}# {{ .Title }}{{ end }}');
			const two = set.RenderTemplate("body", { Title: "Demo" });
			({ one: one.Text, two: two.Text, hasBody: set.Lookup("body").Name === "body", templates: set.Templates().length });
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
	if got["one"] != "Ada" || got["two"] != "# Demo" || got["hasBody"] != true {
		t.Fatalf("unexpected result: %#v", got)
	}
}

func TestRequireTemplateJSFunc(t *testing.T) {
	rt := newTemplateRuntime(t)

	ret, err := rt.Owner.Call(context.Background(), "template.jsfunc", func(_ context.Context, vm *goja.Runtime) (any, error) {
		value, runErr := vm.RunString(`
			const template = require("template");
			const text = template.text()
				.JSFunc("surround", (left, value, right) => String(left) + String(value).toUpperCase() + String(right))
				.Parse('{{ surround "[" .Name "]" }}')
				.Render({ Name: "ada" })
				.Text;
			const html = template.html()
				.JSFunc("rawish", () => '<script>alert(1)</script>')
				.Parse('<div>{{ rawish }}</div>')
				.Render({})
				.Text;
			({ text, html });
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
	if got["text"] != "[ADA]" {
		t.Fatalf("text = %#v", got["text"])
	}
	html := got["html"].(string)
	if strings.Contains(html, "<script>") || !strings.Contains(html, "&lt;script&gt;") {
		t.Fatalf("JSFunc HTML return was not escaped: %s", html)
	}
}

func TestRequireTemplateJSFuncErrors(t *testing.T) {
	rt := newTemplateRuntime(t)

	_, err := rt.Owner.Call(context.Background(), "template.jsfunc.error", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, runErr := vm.RunString(`
			(() => {
				const template = require("template");
				template.text().JSFunc("bad-name", () => "x").Parse("{{ . }}");
			})();
		`)
		return nil, runErr
	})
	if err == nil || !strings.Contains(err.Error(), "invalid JS function name") {
		t.Fatalf("error = %v, want invalid JS function name", err)
	}

	_, err = rt.Owner.Call(context.Background(), "template.jsfunc.throw", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, runErr := vm.RunString(`
			(() => {
				const template = require("template");
				template.text().JSFunc("boom", () => { throw new Error("kaboom"); }).Parse("{{ boom }}").Render({});
			})();
		`)
		return nil, runErr
	})
	if err == nil || !strings.Contains(err.Error(), "kaboom") {
		t.Fatalf("error = %v, want thrown JS error", err)
	}
}

func TestRequireTemplateValidationErrors(t *testing.T) {
	rt := newTemplateRuntime(t)

	_, err := rt.Owner.Call(context.Background(), "template.validation", func(_ context.Context, vm *goja.Runtime) (any, error) {
		_, runErr := vm.RunString(`
			const template = require("template");
			template.text().Funcs("none", "sprig").MissingKey("bogus").Parse("{{ .Name }}");
		`)
		return nil, runErr
	})
	if err == nil || !strings.Contains(err.Error(), "none") || !strings.Contains(err.Error(), "missingKey") {
		t.Fatalf("error = %v, want validation error mentioning func set and missingKey", err)
	}
}

func newTemplateRuntime(t *testing.T) *engine.Runtime {
	t.Helper()
	factory, err := engine.NewRuntimeFactoryBuilder().UseModuleMiddleware(engine.MiddlewareOnly("template")).Build()
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
