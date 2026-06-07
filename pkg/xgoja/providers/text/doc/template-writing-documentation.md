---
Title: "Writing documentation with goja-text templates"
Slug: goja-text-template-writing-documentation
Short: "Use the template module and template jsverbs to generate Markdown and HTML documentation from structured data."
Topics:
- goja-text
- template
- templating
- documentation
- jsverbs
Commands:
- goja-text template text
- goja-text template html
- goja-text template inspect
- goja-text template check
- goja-text run
Flags:
- missingKey
- funcs
- templateName
- outputPath
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

Documentation templates turn structured facts into repeatable prose. The `template` module uses Go `text/template` for Markdown and plain text, `html/template` for contextually escaped HTML, and Glazed/Sprig helper functions for common formatting. Use it when the same document shape appears repeatedly: release notes, ticket summaries, API tables, prompt packs, fixture reports, or generated README sections.

The generated `goja-text` binary exposes the same renderer in two ways. Scripts can call `require("template")` for full control, and root-mounted jsverbs under `goja-text template ...` provide practical command-line rendering for template files plus YAML or JSON data. The command names are deliberately short: `template text`, `template html`, `template inspect`, `template check`, `template examples`, `template example`, and `template helper-demo`.

## Try the embedded examples first

The generated binary bundles reusable template examples as read-only xgoja assets mounted at `/templates` through `require("fs:assets")`. Use them as copyable starting points before writing your own templates.

```bash
./dist/goja-text template examples
./dist/goja-text template example report
./dist/goja-text template example api-reference
./dist/goja-text template example page --output-path page.html
```

The bundled examples demonstrate a Markdown status report, a Markdown API reference table, and an HTML page that filters unsafe URLs through `html/template`.

## Start with a small Markdown template

A documentation template should be boring to read and strict to render. Keep template logic shallow, name the data fields clearly, and let `missingkey=error` catch incomplete data before a broken document is published.

Create `doc.tmpl.md`:

```gotemplate
# {{ .Title }}

{{ .Summary }}

## Items

{{ range .Items -}}
- **{{ .Name }}** — {{ .Description }}
{{ end }}
```

Create `doc.yaml`:

```yaml
title: Template Module
summary: Go-backed template rendering for goja-text.
items:
  - name: text
    description: Renders Markdown, prompts, and plain text.
  - name: html
    description: Renders HTML with contextual escaping.
```

Render it from the generated binary:

```bash
./dist/goja-text template text doc.tmpl.md --data-file doc.yaml
```

If the YAML data uses lowercase keys, address them as lowercase in the template (`.title`) or normalize the data keys before rendering. Go-backed result objects from goja-text modules expose PascalCase fields such as `Text`, `TemplateName`, and `Bytes`; plain YAML/JSON data keeps the keys supplied by the file.

## Validate and inspect before rendering

Validation catches builder option problems and parse errors without producing output. Inspection lists named templates so you can confirm which `define` blocks are available before selecting one with `--template-name`.

```bash
./dist/goja-text template check doc.tmpl.md --missing-key error
./dist/goja-text template inspect doc.tmpl.md
```

Named templates make larger documentation packs easier to organize:

```gotemplate
{{ define "title" }}{{ .Title }}{{ end }}
{{ define "body" -}}
# {{ template "title" . }}

{{ .Summary }}
{{- end }}
```

Render the named body template:

```bash
./dist/goja-text template text report.tmpl.md --data-file report.yaml --template-name body
```

## Use helpers deliberately

The default helper presets are `sprig` and `glazed`. They provide string functions, formatting helpers, YAML serialization, padding, and other conveniences that are useful in generated documentation.

```gotemplate
# {{ .Title | upper }}

{{ toYaml .Metadata }}
```

Disable helper presets when you want a minimal template surface:

```bash
./dist/goja-text template text doc.tmpl.md --data-file doc.yaml --funcs none
```

Use `JSFunc` from scripts when a template needs a small local helper that is easier to express in JavaScript than in template syntax:

```js
const template = require("template");

const result = template.text()
  .JSFunc("badge", (value) => `[[${String(value).toUpperCase()}]]`)
  .Parse("Status: {{ badge .Status }}")
  .Render({ Status: "ready" });
```

Keep `JSFunc` helpers synchronous and side-effect-light. They run during template execution, so slow helpers make rendering slow, and thrown errors fail the render.

## Render HTML only with html mode

HTML templates should use `template.html()` or the `template html` jsverb. Go's `html/template` escapes values according to where they appear in the document, including text nodes, attributes, and URLs.

```bash
./dist/goja-text template html page.tmpl.html --data-file page.yaml --output-path page.html
```

A value such as `<Ada>` becomes escaped text, and unsafe URLs are filtered by `html/template`. Do not render HTML with `template text` unless you explicitly want unescaped text substitution.

## Write output files in automation

Use `--output-path` when a script should create a document artifact instead of printing the rendered text to stdout.

```bash
./dist/goja-text template text changelog.tmpl.md --data-file changelog.yaml --output-path CHANGELOG.generated.md
```

The command returns a small object with the written path and byte count so Glazed output modes can still report what happened.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `map has no entry for key` | The default `missingkey=error` policy found a missing data field. | Add the field to the YAML/JSON data, fix the field name's casing, or choose `--missing-key default` intentionally. |
| Template parses but named render fails | `--template-name` does not match a `{{ define "..." }}` block. | Run `goja-text template inspect <file>` and copy the exact template name. |
| Helper function is not defined | The template uses a Sprig/Glazed helper but helpers were disabled. | Use the default `--funcs sprig,glazed` or include the needed preset. |
| HTML output contains `#ZgotmplZ` | `html/template` rejected an unsafe URL or context. | Check URL fields and avoid `javascript:` or malformed links. |
| JSFunc helper returns `[object Promise]` or unexpected output | `JSFunc` is synchronous and does not await promises. | Resolve asynchronous data before rendering and pass the final value into the template. |
| Data fields are unexpectedly lowercase or uppercase | YAML/JSON data preserves file keys, while Go-backed objects expose exported Go field names. | Match the template selector to the actual data shape, or normalize data in JavaScript before rendering. |
| Embedded example path cannot be written | xgoja embedded assets are read-only. | Use `--output-path` to write rendered output to the host filesystem instead of trying to modify `/templates`. |

## See Also

- `goja-text-template-api-reference` — exact JavaScript API for builders, render results, and `JSFunc`.
- `goja-text-template-user-guide` — guided introduction to the template module from scripts.
- `goja-text template examples` — list bundled reusable template assets.
- `goja-text-markdown-user-guide` — parse and inspect Markdown produced by documentation templates.
- `goja-text-extract-user-guide` — extract structured payloads from generated or source documents.
