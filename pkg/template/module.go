package template

import (
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

Go-backed builders and parsed template sets expose exported Go method and field names in JavaScript:
  template.text().Name("prompt").Funcs("sprig", "glazed").Parse("Hello {{ .Name }}").Render({ Name: "Ada" })
  result.Text, result.TemplateName, result.Mode, result.Bytes
`
}

func (mod module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
	exports := moduleObj.Get("exports").(*goja.Object)

	modules.SetExport(exports, mod.Name(), "text", func() *TemplateBuilder {
		return NewTextBuilder()
	})
	modules.SetExport(exports, mod.Name(), "html", func() *TemplateBuilder {
		return NewHTMLBuilder()
	})
	modules.SetExport(exports, mod.Name(), "renderText", func(source string, data goja.Value) (*RenderResult, error) {
		return RenderText(source, exportTemplateData(data))
	})
	modules.SetExport(exports, mod.Name(), "renderHTML", func(source string, data goja.Value) (*RenderResult, error) {
		return RenderHTML(source, exportTemplateData(data))
	})
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
