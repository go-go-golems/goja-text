---
Title: "goja-text template JavaScript API reference"
Slug: goja-text-template-api-reference
Short: "Reference for require(\"template\") in goja-text xgoja runtimes."
Topics:
- goja-text
- template
- templating
- javascript
- xgoja
Commands:
- goja-text
- goja-text eval
- goja-text run
- help
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

Use `require("template")` when JavaScript needs Go `text/template` or `html/template` rendering with Go-backed builders and result objects.

The module exposes Go-backed objects. JavaScript therefore uses exported Go method and field names such as `Name`, `Funcs`, `MissingKey`, `Parse`, `Render`, `Text`, `TemplateName`, and `Bytes`.

## Loading

```js
const template = require("template");
```

## Top-level functions

### text()

Creates a Go-backed `text/template` builder.

```js
const set = template.text()
  .Name("greeting")
  .Funcs("sprig", "glazed")
  .MissingKey("error")
  .Parse("Hello {{ .Name | upper }}");

console.log(set.Render({ Name: "intern" }).Text);
```

### html()

Creates a Go-backed `html/template` builder. HTML mode uses Go's contextual escaping rules.

```js
const result = template.html()
  .Parse('<p>{{ .Name }}</p><a href="{{ .URL }}">open</a>')
  .Render({ Name: '<Ada>', URL: 'javascript:alert(1)' });

console.log(result.Text);
```

### renderText(source, data?)

Parses and renders a text template in one call. This is convenience sugar over `template.text().Parse(source).Render(data)`.

### renderHTML(source, data?)

Parses and renders an HTML template in one call. This is convenience sugar over `template.html().Parse(source).Render(data)`.

## TemplateBuilder methods

- `Name(name)` — set the root template name.
- `Funcs(...names)` — choose helper presets. Supported names are `"sprig"`, `"glazed"`, and `"none"`.
- `MissingKey(policy)` — choose Go template missing-key behavior: `"default"`, `"invalid"`, `"zero"`, or `"error"`.
- `Delims(left, right)` — set custom delimiters.
- `Validate()` — return `{ Valid, Errors }` without parsing.
- `BuildConfig()` — return a frozen Go-backed config or throw if invalid.
- `Parse(source)` — parse source using the current builder name.
- `ParseNamed(name, source)` — parse source as a named template.

## TemplateSet methods and fields

- `Mode` — `"text"` or `"html"`.
- `Name` — default template name.
- `Render(data?)` — execute the default template.
- `RenderString(data?)` — execute and return only the rendered string.
- `RenderTemplate(name, data?)` — execute a named template from the set.
- `Templates()` — list template metadata.
- `Lookup(name)` — return metadata for one named template or `undefined`/`null`.

## RenderResult fields

- `Text` — rendered output.
- `TemplateName` — executed template name.
- `Mode` — `"text"` or `"html"`.
- `Bytes` — byte length of `Text`.

## Notes

- The default helper presets are `sprig` and `glazed`.
- The default missing-key policy is `error` for automation safety.
- JavaScript callback functions inside templates are intentionally not part of this phase.
