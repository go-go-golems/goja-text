package template

import (
	"fmt"
	"regexp"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-goja/modules"
)

type module struct{}

var _ modules.NativeModule = (*module)(nil)
var _ modules.TypeScriptDeclarer = (*module)(nil)

func (module) Name() string { return "template" }

func (module) Doc() string {
	return `
The template module renders Go text/template and html/template documents from JavaScript.

Functions:
  text(): Create a Go-backed text/template builder.
  html(): Create a Go-backed html/template builder with contextual escaping.
  renderText(source, data?): Render a text template in one call.
  renderHTML(source, data?): Render an HTML template in one call.
  builder.JSFunc(name, fn): Register a synchronous JavaScript template helper.

Go-backed builders and parsed template sets expose exported Go method and field names in JavaScript:
  template.text().Name("prompt").Funcs("sprig", "glazed").Parse("Hello {{ .Name }}").Render({ Name: "Ada" })
  result.Text, result.TemplateName, result.Mode, result.Bytes
`
}

func (mod module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)

	modules.SetExport(exports, mod.Name(), "text", func() *TemplateBuilder {
		builder := NewTextBuilder()
		builder.vm = vm
		return builder
	})
	modules.SetExport(exports, mod.Name(), "html", func() *TemplateBuilder {
		builder := NewHTMLBuilder()
		builder.vm = vm
		return builder
	})
	modules.SetExport(exports, mod.Name(), "renderText", func(source string, data goja.Value) (*RenderResult, error) {
		return RenderText(source, exportTemplateData(data))
	})
	modules.SetExport(exports, mod.Name(), "renderHTML", func(source string, data goja.Value) (*RenderResult, error) {
		return RenderHTML(source, exportTemplateData(data))
	})
}

var templateFuncNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// JSFunc registers a JavaScript function as a Go template helper.
//
// The wrapper is synchronous: the JavaScript function is called while the
// template executes. Returned JavaScript values are exported back to Go before
// text/template or html/template receives them. In HTML mode, ordinary returned
// strings remain untrusted strings and are escaped by html/template.
func (b *TemplateBuilder) JSFunc(name string, value goja.Value) *TemplateBuilder {
	if !templateFuncNamePattern.MatchString(name) {
		b.errors = append(b.errors, fmt.Sprintf("invalid JS function name %q", name))
		return b
	}
	fn, ok := goja.AssertFunction(value)
	if !ok {
		b.errors = append(b.errors, fmt.Sprintf("JSFunc %q must be a function", name))
		return b
	}
	if b.vm == nil {
		b.errors = append(b.errors, fmt.Sprintf("JSFunc %q requires a goja runtime-backed builder", name))
		return b
	}
	if b.customFuncs == nil {
		b.customFuncs = map[string]any{}
	}
	b.customFuncs[name] = func(args ...any) (any, error) {
		jsArgs := make([]goja.Value, 0, len(args))
		for _, arg := range args {
			jsArgs = append(jsArgs, b.vm.ToValue(arg))
		}
		ret, err := fn(goja.Undefined(), jsArgs...)
		if err != nil {
			return nil, err
		}
		return exportTemplateData(ret), nil
	}
	return b
}

func exportTemplateData(value goja.Value) any {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil
	}
	return value.Export()
}

func init() {
	modules.Register(&module{})
}
