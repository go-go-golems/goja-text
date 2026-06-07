---
Title: "goja-text template user guide"
Slug: goja-text-template-user-guide
Short: "A guided introduction to rendering Go templates from JavaScript."
Topics:
- goja-text
- template
- templating
- html
- guide
Commands:
- goja-text eval
- goja-text run
- help
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

The template module lets JavaScript scripts render Go `text/template` and `html/template` documents without leaving the goja runtime. It is useful for prompts, Markdown reports, HTML snippets, generated configuration, and command output.

The important design choice is that template objects are Go-backed. JavaScript drives the workflow with a fluent builder, but the parsed template set, validation state, render result, and metadata remain typed Go objects.

## Render a text template

```js
const template = require("template");

const result = template.text()
  .Name("prompt")
  .Parse("Hello {{ .Name }}")
  .Render({ Name: "Ada" });

console.log(result.Text);
```

Read fields with PascalCase names because they are exported Go fields.

## Use helper functions

By default, builders include Sprig and Glazed helper functions. That means common string, date, formatting, YAML, and padding helpers are available inside templates.

```js
const result = template.text()
  .Parse("Project: {{ .Project | upper }}\n{{ toYaml .Items }}")
  .Render({ Project: "goja-text", Items: ["markdown", "sanitize", "template"] });

console.log(result.Text);
```

Use `.Funcs("none")` for a minimal template with no helper presets.

## Use named templates

Go templates can contain definitions that are executed by name.

```js
const set = template.text().Name("report").Parse(`
{{ define "title" }}Report for {{ .Project }}{{ end }}
{{ define "body" -}}
# {{ template "title" . }}
{{ range .Items }}- {{ . }}
{{ end }}
{{- end }}
`);

const result = set.RenderTemplate("body", {
  Project: "goja-text",
  Items: ["Markdown", "Sanitize", "Templates"],
});

console.log(result.Text);
console.log(set.Templates().map((t) => t.Name));
```

## Choose HTML mode when rendering HTML

HTML mode uses Go's `html/template`, which performs contextual escaping.

```js
const result = template.html()
  .Parse('<p>{{ .Name }}</p><a href="{{ .URL }}">open</a>')
  .Render({ Name: '<Ada>', URL: 'javascript:alert(1)' });

console.log(result.Text);
```

Use `html()` for HTML, SVG, or any document where escaping matters. Use `text()` for Markdown, prompts, plain text, and configuration files.

## Fail fast on missing data

The default missing-key policy is `error`. This makes automation fail loudly when a template expects data that the script did not provide.

```js
try {
  template.text().Parse("Hello {{ .Name }}").Render({});
} catch (err) {
  console.error("template failed:", err.message);
}
```

If you need standard Go behavior, choose a different policy explicitly:

```js
template.text().MissingKey("default").Parse("Hello {{ .Name }}");
```

## Running the included example

```bash
./dist/goja-text run examples/js/template-demo.js
```

## Key points

- Use `template.text()` for plain text and `template.html()` for contextual HTML escaping.
- Builders and results are Go-backed, so JS reads PascalCase methods and fields.
- Prefer builder methods over raw option objects.
- JavaScript functions inside templates are a future phase, not part of the current module contract.
