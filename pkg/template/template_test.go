package template

import (
	"strings"
	"testing"
)

func TestTextBuilderRender(t *testing.T) {
	set, err := NewTextBuilder().Parse("Hello {{ .Name | upper }}")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result, err := set.Render(map[string]any{"Name": "intern"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if result.Text != "Hello INTERN" || result.Mode != ModeText || result.TemplateName != defaultTemplateName {
		t.Fatalf("result = %#v", result)
	}
}

func TestHTMLBuilderEscapes(t *testing.T) {
	set, err := NewHTMLBuilder().Parse(`<p>{{ .Name }}</p><a href="{{ .URL }}">open</a>`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result, err := set.Render(map[string]any{"Name": "<Ada>", "URL": "javascript:alert(1)"})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(result.Text, "&lt;Ada&gt;") || strings.Contains(result.Text, "javascript:alert") {
		t.Fatalf("html output was not contextually escaped: %s", result.Text)
	}
}

func TestNamedTemplates(t *testing.T) {
	set, err := NewTextBuilder().Name("report").Parse(`{{ define "body" }}# {{ .Title }}{{ end }}`)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	result, err := set.RenderTemplate("body", map[string]any{"Title": "Demo"})
	if err != nil {
		t.Fatalf("RenderTemplate() error = %v", err)
	}
	if result.Text != "# Demo" {
		t.Fatalf("result.Text = %q", result.Text)
	}
	if set.Lookup("body") == nil {
		t.Fatalf("Lookup(body) = nil")
	}
}

func TestMissingKeyError(t *testing.T) {
	set, err := NewTextBuilder().Parse("Hello {{ .Name }}")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	_, err = set.Render(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "map has no entry for key") {
		t.Fatalf("Render() error = %v, want missing key error", err)
	}
}

func TestBuilderValidation(t *testing.T) {
	result := NewTextBuilder().Name("").Funcs("none", "sprig").MissingKey("bogus").Delims("[[", "[[").Validate()
	if result.Valid || len(result.Errors) < 4 {
		t.Fatalf("Validate() = %#v, want multiple errors", result)
	}
}

func TestRenderConvenience(t *testing.T) {
	result, err := RenderText("{{ . }}", "ok")
	if err != nil {
		t.Fatalf("RenderText() error = %v", err)
	}
	if result.Text != "ok" {
		t.Fatalf("Text = %q", result.Text)
	}
}
