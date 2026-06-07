---
Title: Template Module Analysis Design and Implementation Guide
Ticket: GOJA-TEXT-004
Status: active
Topics:
    - goja
    - goja-bindings
    - native-modules
    - text-algorithms
    - templating
    - html
    - xgoja
    - jsverbs
    - cli
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: glazed/pkg/helpers/templating/templating.go
      Note: Reusable text and html template constructors Sprig integration Glazed helper FuncMap and FS parsing helpers
    - Path: goja-text/cmd/goja-text/xgoja.yaml
      Note: Generated binary module selection that must expose the template module
    - Path: goja-text/pkg/extract/module.go
      Note: Simple native module loader and Go-backed result object pattern
    - Path: goja-text/pkg/markdown/module.go
      Note: Canonical Go-backed AST module TypeScriptDeclarer Loader SetExport and JS callback traversal pattern
    - Path: goja-text/pkg/sanitize/module.go
      Note: Namespace export and Go-backed options-builder pattern to mirror for template builders
    - Path: goja-text/pkg/sanitize/options.go
      Note: Builder validation and Build config pattern for fluent Go-backed options
    - Path: goja-text/pkg/xgoja/providers/text/text.go
      Note: xgoja provider registration point that must include the template module
ExternalSources: []
Summary: Design for adding Go text/template and html/template support to goja-text as an elegant Go-backed JavaScript API.
LastUpdated: 2026-06-07T16:20:00-04:00
WhatFor: Use when implementing the goja-text template module, xgoja provider wiring, docs, examples, and tests.
WhenToUse: Before writing code for GOJA-TEXT-004 or when reviewing the proposed JavaScript API and implementation plan.
---


# Template Module Analysis Design and Implementation Guide

## Executive Summary

`goja-text` should add a new native JavaScript module, tentatively `require("template")`, that exposes Go's `text/template` and `html/template` engines to goja scripts. The API should feel fluent in JavaScript while remaining Go-backed on the object side: scripts call builder methods such as `template.text().Name("invoice").Funcs("glazed", "sprig").Parse(src).Render(data)`, but the mutable template configuration, parsed template set, validation state, and render results live in Go structs rather than loosely shaped JavaScript maps.

This is a natural extension of the existing repository. `goja-text` already exposes Go parsers and text processors as go-go-goja `NativeModule` packages, and it intentionally uses Go-backed ASTs, candidates, builders, configs, and result objects where the Go representation is the domain model. The template module should follow the same pattern: keep parsing, function-map construction, safe data wrapping, and rendering in Go, while giving JavaScript a small set of composable calls for selecting text vs HTML behavior, parsing inline strings or named template sets, and executing templates.

The recommended first phase is synchronous and deterministic:

- Add `pkg/template` or `pkg/templates` in `goja-text` with a pure-ish service layer plus goja module adapter.
- Support `text/template` and `html/template` through separate builder entrypoints.
- Reuse Glazed's helper functions and Sprig integration by default or through explicit builder presets.
- Return Go-backed `TemplateSet`, `RenderResult`, `TemplateInfo`, `TemplateOptionsBuilder`, and `TemplateConfig` objects.
- Add TypeScript declarations, provider registration, xgoja buildspec entries, runtime help pages, JS examples, jsverbs, and integration tests.
- Defer JavaScript callback functions inside templates to a second phase after the data/escaping/function-call contract is reviewed carefully.

## Problem Statement and Scope

Today, `goja-text` can parse Markdown, sanitize YAML/JSON, and extract structured snippets, but it cannot render Go templates from JavaScript. That means scripts that already live in the xgoja runtime must either hand-roll string interpolation, shell out to another tool, or import a separate Go host capability to render templated prompts, Markdown reports, HTML snippets, SQL fragments, or CLI output.

The requested outcome is a polished goja JavaScript API over Go's template engines, not a thin `render(template, object)` wrapper. The API should be elegant, builder-oriented, and fluid. It should also avoid pushing all domain state into JavaScript maps and objects. In this repository, Go-backed objects are a deliberate pattern because they give the Go side validation, method chaining, stable field names, typed configs, and better error messages.

### In scope for phase 1

- Go native module exposed as `require("template")` or a carefully chosen alias.
- `text/template` rendering.
- `html/template` rendering with Go's contextual escaping semantics.
- Inline parse and render helpers for simple use cases.
- Fluent Go-backed builders and parsed template-set objects for real use cases.
- Glazed helper function integration, including Sprig and the Glazed `TemplateFuncs` helpers.
- Named template definitions and `ExecuteTemplate`-style rendering.
- Data conversion rules that can accept JavaScript data but normalize it into Go-owned values before execution.
- Tests at service and runtime levels.
- Runtime API help pages and xgoja build inclusion.

### Out of scope for phase 1

- Calling arbitrary JavaScript functions from inside templates.
- Asynchronous template functions or Promise-aware rendering.
- File-system template loading in the default module without reviewing xgoja host capability boundaries.
- Sandboxing beyond what Go templates already provide.
- Replacing Glazed's existing templating package.

Phase 2 should add carefully reviewed JavaScript callback exports to template function maps. That is explicitly called out later because it crosses runtime ownership, exception translation, value conversion, and potential reentrancy boundaries.

## Current-State Architecture

### Repository module pattern

`goja-text` exposes each text capability as a go-go-goja native module. The Markdown module implements `modules.NativeModule`, declares TypeScript metadata, registers itself during `init()`, and wires functions in `Loader` through `modules.SetExport` (`/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/markdown/module.go:13`, `:14`, `:78`, `:81`). The extract and sanitize modules use the same pattern (`/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/extract/module.go:12`, `:35`; `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/sanitize/module.go:14`, `:45`).

The go-go-goja contract is small and important. A module has `Name()`, `Doc()`, and `Loader(*goja.Runtime, *goja.Object)`, and the default registry registers each module with goja-nodejs `require` (`/home/manuel/workspaces/2026-06-07/goja-text-templates/go-go-goja/modules/common.go:28`, `:88`, `:96`). The go-go-goja docs describe the same pattern and show that `Name()` becomes the `require()` name, while `Loader` populates `module.exports` (`/home/manuel/workspaces/2026-06-07/goja-text-templates/go-go-goja/pkg/doc/02-creating-modules.md:18`, `:40`, `:44`).

```text
JavaScript script
    |
    | require("markdown"), require("sanitize"), require("extract")
    v
go-go-goja require.Registry
    |
    | registered NativeModule.Loader
    v
goja-text module package
    |
    | Go parser/sanitizer/extractor service code
    v
Go-backed result/config/domain objects visible to JS
```

### Existing Go-backed API style

The repository already documents the exact design preference requested for templates. Markdown returns Go-backed AST nodes, and JavaScript reads exported Go fields such as `node.Type` and `node.Children` (`/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/markdown/module.go:23`, `:29`). Sanitize exposes Go-backed options builders such as `sanitize.yaml.options().MaxIterations(5).Build()` (`/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/sanitize/module.go:24`, `:39`). Extract returns Go-backed candidates with source evidence (`/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/extract/module.go:22`, `:30`).

This matters because Go templates have a rich object model: a template set has names, definitions, parse trees, function maps, execution mode, escaping behavior, and errors. That state should not be represented as free-form JS maps. It should live in Go-backed builders and parsed template objects.

### xgoja provider and generated binary wiring

The generated `goja-text` binary does not automatically include every Go package. Its provider imports module packages for registration, finds them in the default registry, and exposes them as xgoja provider modules (`/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/xgoja/providers/text/text.go:15`, `:17`, `:23`, `:39`). The current xgoja buildspec selects `markdown`, `sanitize`, and `extract` under the `goja-text` provider (`/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/cmd/goja-text/xgoja.yaml:25`).

The xgoja docs explain the two boundaries:

- A provider package compiles modules into the generated binary.
- The top-level `modules` list chooses what JavaScript can `require()` (`/home/manuel/workspaces/2026-06-07/goja-text-templates/go-go-goja/cmd/xgoja/doc/04-tutorial-providing-package-and-modules.md:28`, `:68`, `:92`).

Therefore the template implementation must update both `pkg/xgoja/providers/text/text.go` and `cmd/goja-text/xgoja.yaml`, not just add a new module package.

### Glazed templating helpers

Glazed already has a reusable package at `github.com/go-go-golems/glazed/pkg/helpers/templating`. It builds on `text/template`, `html/template`, Sprig, and Glazed-specific helper functions. The implementation imports Sprig, defines `TemplateFuncs`, exposes `RenderTemplateString`, `RenderHtmlTemplateString`, `CreateTemplate`, `CreateHTMLTemplate`, `ParseFS`, and `ParseHTMLFS` (`/home/manuel/workspaces/2026-06-07/goja-text-templates/glazed/pkg/helpers/templating/templating.go:22`, `:56`, `:568`, `:577`, `:594`, `:606`, `:649`). The Glazed docs describe `CreateTemplate` as the primary way to construct a template enriched with Sprig and Glazed functions (`/home/manuel/workspaces/2026-06-07/goja-text-templates/glazed/pkg/doc/topics/22-templating-helpers.md:20`, `:35`, `:40`, `:49`).

This package should be reused rather than reimplemented. However, the goja API still needs its own service layer because it must expose fluent builders, JS data conversion, parsed template sets, TypeScript declarations, runtime help, and future JS callback support.

## Proposed JavaScript API

### Module shape

Use `require("template")` unless the team wants to avoid collision with the Go package name or with common JS names. The module exports two namespaces/entrypoints: `text` and `html`. Each returns a builder. Convenience helpers exist for simple one-shot rendering.

```js
const template = require("template");

const out = template.text()
  .Name("greeting")
  .Funcs("glazed", "sprig")
  .MissingKey("error")
  .Parse("Hello {{ .Name | upper }}")
  .Render({ Name: "intern" });

console.log(out.Text);          // "Hello INTERN"
console.log(out.TemplateName);  // "greeting"
```

A more complete named-template example:

```js
const template = require("template");

const set = template.text()
  .Name("report")
  .Funcs("glazed", "sprig")
  .Parse(`
{{ define "title" }}Report for {{ .Project }}{{ end }}
{{ define "body" -}}
# {{ template "title" . }}

Items:
{{ range .Items }}- {{ . | trim }}
{{ end }}
{{- end }}
`);

const result = set.RenderTemplate("body", {
  Project: "goja-text",
  Items: [" markdown ", " sanitize ", " templates "],
});

console.log(result.Text);
console.log(set.Templates().map(t => t.Name));
```

HTML rendering should be explicit so reviewers know where escaping semantics come from:

```js
const template = require("template");

const html = template.html()
  .Name("card")
  .Funcs("glazed", "sprig")
  .Parse(`<p>Hello {{ .Name }}</p><a href="{{ .URL }}">open</a>`)
  .Render({ Name: `<Manuel>`, URL: `javascript:alert(1)` });

console.log(html.Text); // escaped/sanitized according to html/template contextual rules
```

### Convenience helpers

The module should also expose one-shot helpers for scripts and examples:

```js
template.renderText("Hello {{ .Name }}", template.data().Set("Name", "Ada"));
template.renderHTML("<b>{{ .Name }}</b>", { Name: "Ada" });
```

These helpers should be documented as sugar over the builder path. They must not become the primary implementation location.

### Go-backed object model

Recommended Go-backed objects:

| Object | Created by | Purpose | JS-visible shape |
| --- | --- | --- | --- |
| `TemplateBuilder` | `template.text()`, `template.html()` | Accumulates immutable-or-copyable parse/render config | Methods: `Name`, `Funcs`, `MissingKey`, `Delims`, `Option`, `Parse`, `ParseNamed`, `ParseFiles?`, `Validate`, `BuildConfig` |
| `TemplateConfig` | `builder.BuildConfig()` | Frozen render/parse config | Fields: `Mode`, `Name`, `FuncSets`, `MissingKey`, `LeftDelim`, `RightDelim` |
| `TemplateSet` | `builder.Parse(...)` | Parsed text or HTML template set | Methods: `Render`, `RenderTemplate`, `Templates`, `Lookup`, `MustRender`; fields: `Mode`, `Name` |
| `RenderResult` | `Render...` methods | Evidence-rich render output | Fields: `Text`, `TemplateName`, `Mode`, `Bytes`, `DurationMillis?` |
| `TemplateInfo` | `Templates()`, `Lookup()` | Metadata for defined templates | Fields: `Name`, `Defined`, `Mode` |
| `DataBuilder` (optional) | `template.data()` | Avoids large anonymous maps for common scripts | Methods: `Set`, `SetPath?`, `Build`; fields only after build |

Why `RenderResult` instead of returning a bare string? Existing goja-text modules return evidence-rich objects where useful. Rendering has enough metadata to justify a result object. A convenience `RenderString` can return just text if needed.

### Fluent builder naming

Follow existing Go-backed builder style from `sanitize` and `extract`: exported Go methods produce PascalCase method names in JavaScript. The docs should explicitly show `builder.MissingKey("error")`, not lowerCamel, because the object is Go-backed and this is consistent with `sanitize.yaml.options().MaxIterations(5).Build()`.

```js
const cfg = template.text()
  .Name("prompt")
  .Funcs("glazed")
  .MissingKey("error")
  .Delims("[[", "]]")
  .BuildConfig();

console.log(cfg.Name, cfg.FuncSets, cfg.MissingKey);
```

## Proposed Go Package Layout

Use a separate service/adaptation split even if the first commit is small.

```text
goja-text/
  pkg/
    template/
      doc.go                    # package overview
      module.go                 # NativeModule + TypeScriptDeclarer + Loader wiring
      builder.go                # TemplateBuilder, TemplateConfig, validation
      render.go                 # TemplateSet, RenderResult, execution helpers
      funcs.go                  # function-set selection and Glazed/Sprig integration
      data.go                   # JS value normalization / optional DataBuilder
      jsfuncs.go                # phase-2 JS callback function adapter (stub/design only at first)
      typescript.go             # optional generated/declarative raw DTS helpers
      *_test.go                 # service tests and runtime integration tests
    xgoja/providers/text/
      text.go                   # add blank import + textModuleNames entry
      doc/
        template-api-reference.md
        template-user-guide.md
  cmd/goja-text/
    xgoja.yaml                  # add module alias
    jsverbs/template.js         # optional practical commands after API stabilizes
  examples/js/template-demo.js
```

This keeps module loading small and reviewable, which matches the go-go-goja module-authoring guidance: domain/service code belongs outside `Loader`, while `Loader` should mostly decode inputs and wire exports.

## Service-Layer API Sketch

The service layer should not import `goja` except where unavoidable for future callback adapters. It should know about Go templates, function sets, config validation, and data normalization.

```go
type Mode string

const (
    ModeText Mode = "text"
    ModeHTML Mode = "html"
)

type TemplateConfig struct {
    Mode       Mode
    Name       string
    FuncSets   []string
    MissingKey string // "default", "zero", "error", "invalid"
    LeftDelim  string
    RightDelim string
}

type TemplateBuilder struct {
    cfg    TemplateConfig
    errors []string
}

func NewTextBuilder() *TemplateBuilder { ... }
func NewHTMLBuilder() *TemplateBuilder { ... }

func (b *TemplateBuilder) Name(name string) *TemplateBuilder { ... }
func (b *TemplateBuilder) Funcs(names ...string) *TemplateBuilder { ... }
func (b *TemplateBuilder) MissingKey(policy string) *TemplateBuilder { ... }
func (b *TemplateBuilder) Delims(left, right string) *TemplateBuilder { ... }
func (b *TemplateBuilder) Validate() ValidationResult { ... }
func (b *TemplateBuilder) BuildConfig() (*TemplateConfig, error) { ... }
func (b *TemplateBuilder) Parse(src string) (*TemplateSet, error) { ... }
func (b *TemplateBuilder) ParseNamed(name, src string) (*TemplateSet, error) { ... }
```

The parsed set needs to hide `*template.Template` vs `*html/template.Template` behind one Go-backed type.

```go
type TemplateSet struct {
    Mode Mode
    Name string
    text *template.Template
    html *htmltemplate.Template
}

func (s *TemplateSet) Render(data any) (*RenderResult, error) {
    return s.RenderTemplate(s.Name, data)
}

func (s *TemplateSet) RenderTemplate(name string, data any) (*RenderResult, error) {
    // normalize data, execute into buffer, return RenderResult
}

func (s *TemplateSet) Templates() []TemplateInfo { ... }
func (s *TemplateSet) Lookup(name string) *TemplateInfo { ... }
```

## Function-Set Design

### Presets

The builder should select function sets by stable names:

- `"standard"`: optional minimal standard helpers if we define any.
- `"sprig"`: Sprig text or HTML function map depending on mode.
- `"glazed"`: Glazed `TemplateFuncs`.
- `"none"`: no helper functions other than Go's built-in template operations.

Recommended default: include `glazed` and `sprig` unless there is a compatibility reason to require explicit opt-in. Glazed's own `CreateTemplate` includes both Sprig and `TemplateFuncs`, and this module is likely used by go-go-golems automation where those helpers are expected.

A safe compromise is:

```js
const set = template.text().Parse("{{ .Name }}");                    // default: glazed+sprig
const strict = template.text().Funcs("none").Parse("{{ .Name }}");    // no helpers
const explicit = template.text().Funcs("glazed", "sprig").Parse(src);
```

### Implementation sketch

```go
func funcMapFor(mode Mode, names []string) (template.FuncMap, error) {
    out := template.FuncMap{}
    for _, name := range names {
    switch name {
    case "none":
        // only valid as sole entry
    case "glazed":
        merge(out, glazedtemplating.TemplateFuncs)
    case "sprig":
        if mode == ModeHTML {
            merge(out, sprig.HtmlFuncMap())
        } else {
            merge(out, sprig.TxtFuncMap())
        }
    default:
        return nil, fmt.Errorf("template.funcs: unknown function set %q", name)
    }
    }
    return out, nil
}
```

Prefer using `glazed/pkg/helpers/templating.CreateTemplate` and `CreateHTMLTemplate` for the default constructor path when it fits. If the builder needs fine-grained presets, construct directly with the same ingredients so `Funcs("none")` and future JS callbacks can be controlled precisely.

## Data Conversion Strategy

Go templates execute against Go data. goja can export JavaScript values into maps and slices, but the API should make conversion explicit enough that interns can debug it.

Recommended phase-1 rules:

1. Accept Go-backed values directly, such as Markdown nodes, extraction candidates, sanitize results, and config objects.
2. Accept JavaScript objects and arrays, then export them through goja into Go values before execution.
3. Keep object property names exactly as supplied by JS. If data is a Go-backed object, exported Go field names remain PascalCase.
4. Document that template selectors use Go template rules: `{{ .Name }}` for field/key `Name`, `{{ index . "lowercase" }}` for awkward map keys, and method calls only where Go templates allow them.
5. Default `missingkey` should be `error` for automation safety, or at minimum strongly recommend `.MissingKey("error")` in docs. Silent missing values are hard to debug.

The adapter layer is allowed to import goja and normalize `goja.Value` before passing it to the service:

```go
modules.SetExport(exports, mod.Name(), "renderText", func(src string, data goja.Value) (*RenderResult, error) {
    goData := exportTemplateData(data)
    return RenderText(src, goData, DefaultConfig())
})
```

For builder methods on Go-backed objects, a method signature can accept `any` or `goja.Value` depending on whether the object lives in a package that imports goja. If possible, keep goja-specific conversion in `module.go` and call service functions with `any`.

## Phase 2: JavaScript Functions in Templates

The user request mentions a later phase where JavaScript functions should be exported to the template renderer. This should not be bolted onto phase 1 because Go templates execute functions synchronously and expect normal Go return values or `(value, error)`. JavaScript functions can throw, capture runtime state, and may not be safe to call from arbitrary goroutines.

Proposed phase-2 API:

```js
const set = template.text()
  .Funcs("glazed", "sprig")
  .JSFunc("shout", (s) => String(s).toUpperCase() + "!")
  .JSFunc("link", (text, url) => `[${text}](${url})`)
  .Parse(`{{ shout .Name }} {{ link "docs" .URL }}`);
```

Adapter pseudocode:

```go
func (b *TemplateBuilder) JSFunc(name string, value goja.Value) *TemplateBuilder {
    fn, ok := goja.AssertFunction(value)
    if !ok { b.errors = append(b.errors, name+" must be a function"); return b }
    b.jsFuncs[name] = fn
    return b
}

func wrapJSFunc(vm *goja.Runtime, fn goja.Callable) any {
    return func(args ...any) (any, error) {
        jsArgs := make([]goja.Value, 0, len(args))
        for _, arg := range args { jsArgs = append(jsArgs, vm.ToValue(arg)) }
        ret, err := fn(goja.Undefined(), jsArgs...)
        if err != nil { return nil, err }
        return ret.Export(), nil
    }
}
```

Review risks before implementing:

- Calls must happen on the runtime owner goroutine if the runtime is owner-bound.
- Template execution must remain synchronous or the API must explicitly become async.
- Exceptions should become template execution errors with function names attached.
- Function names must be validated so JS cannot replace reserved helpers accidentally unless explicitly allowed.
- HTML mode must not allow helpers to bypass escaping unless returning trusted `template.HTML`, `template.URL`, etc.; those trusted types should probably be Go-only in phase 2.

## Module Loader Sketch

`module.go` should be small. It should register entrypoints and leave logic to builders/services.

```go
type module struct{}

var _ modules.NativeModule = (*module)(nil)
var _ modules.TypeScriptDeclarer = (*module)(nil)

func (module) Name() string { return "template" }

func (module) Doc() string { return `...` }

func (module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)

    modules.SetExport(exports, "template", "text", func() *TemplateBuilder {
        return NewTextBuilder()
    })
    modules.SetExport(exports, "template", "html", func() *TemplateBuilder {
        return NewHTMLBuilder()
    })
    modules.SetExport(exports, "template", "renderText", func(src string, data goja.Value) (*RenderResult, error) {
        return RenderTextValue(vm, src, data)
    })
    modules.SetExport(exports, "template", "renderHTML", func(src string, data goja.Value) (*RenderResult, error) {
        return RenderHTMLValue(vm, src, data)
    })
}

func init() { modules.Register(&module{}) }
```

## TypeScript Declaration Sketch

The existing modules implement `modules.TypeScriptDeclarer`, and the template module should too.

```ts
export type TemplateMode = "text" | "html";
export type MissingKeyPolicy = "default" | "invalid" | "zero" | "error";
export type FuncSetName = "none" | "sprig" | "glazed" | "standard";

export interface ValidationResult {
  Valid: boolean;
  Errors?: string[];
}

export interface TemplateConfig {
  Mode: TemplateMode;
  Name: string;
  FuncSets: string[];
  MissingKey: MissingKeyPolicy;
  LeftDelim?: string;
  RightDelim?: string;
}

export interface RenderResult {
  Text: string;
  TemplateName: string;
  Mode: TemplateMode;
  Bytes: number;
}

export interface TemplateInfo {
  Name: string;
  Defined: boolean;
  Mode: TemplateMode;
}

export interface TemplateBuilder {
  Name(name: string): TemplateBuilder;
  Funcs(...names: FuncSetName[]): TemplateBuilder;
  MissingKey(policy: MissingKeyPolicy): TemplateBuilder;
  Delims(left: string, right: string): TemplateBuilder;
  Validate(): ValidationResult;
  BuildConfig(): TemplateConfig;
  Parse(source: string): TemplateSet;
  ParseNamed(name: string, source: string): TemplateSet;
}

export interface TemplateSet {
  Mode: TemplateMode;
  Name: string;
  Render(data?: unknown): RenderResult;
  RenderTemplate(name: string, data?: unknown): RenderResult;
  Templates(): TemplateInfo[];
  Lookup(name: string): TemplateInfo | undefined;
}

export function text(): TemplateBuilder;
export function html(): TemplateBuilder;
export function renderText(source: string, data?: unknown): RenderResult;
export function renderHTML(source: string, data?: unknown): RenderResult;
```

## Implementation Plan

### Step 1: Add service package skeleton

Create `pkg/template/doc.go`, `builder.go`, `render.go`, `funcs.go`, and tests. Implement only Go service behavior first.

Acceptance checks:

- `NewTextBuilder().Parse("Hello {{ .Name }}").Render(map[string]any{"Name":"Ada"})` returns expected text.
- `NewHTMLBuilder()` escapes HTML-relevant values.
- Invalid function-set names, empty names, invalid delimiters, and invalid missing-key policies return useful errors.

### Step 2: Add NativeModule adapter

Create `pkg/template/module.go` with `Name() == "template"`, `Doc()`, `Loader`, `TypeScriptModule()`, and `init()` registration. Keep `Loader` small.

Acceptance checks:

- A runtime integration test can `require("template")`.
- JS can call `template.text().Parse(...).Render(...)`.
- JS can call `template.html().Parse(...).Render(...)` and observe escaping.
- Builder validation errors appear as JavaScript exceptions with `template.*` prefixes.

### Step 3: Wire xgoja provider and buildspec

Update `pkg/xgoja/providers/text/text.go`:

- Add blank import `_ "github.com/go-go-golems/goja-text/pkg/template"`.
- Add `"template"` to `textModuleNames`.
- Update help-source description to include templates.

Update `cmd/goja-text/xgoja.yaml`:

```yaml
  - package: goja-text
    name: template
    as: template
```

Regenerate the generated command if required by repository workflow:

```bash
cd goja-text/cmd/goja-text
GOWORK=off go generate
GOWORK=off go build -o ../../dist/goja-text .
```

### Step 4: Add docs and examples

Add provider help docs:

- `pkg/xgoja/providers/text/doc/template-api-reference.md`
- `pkg/xgoja/providers/text/doc/template-user-guide.md`

Add example script:

- `examples/js/template-demo.js`

Optional jsverbs after the API stabilizes:

- `cmd/goja-text/jsverbs/template.js`

The help pages should teach:

- Text vs HTML mode.
- Builder path first, convenience helpers second.
- Go-backed field/method naming.
- Function presets.
- Missing-key policy.
- How Go template selector rules interact with JS data.
- Why JS callbacks are phase 2.

### Step 5: Add tests

Recommended test matrix:

| Test | Layer | Purpose |
| --- | --- | --- |
| `TestTextBuilderRender` | service | basic text/template execution |
| `TestHTMLBuilderEscapes` | service | prove html/template contextual escaping |
| `TestBuilderValidation` | service | bad mode/function/missingkey/delims errors |
| `TestNamedTemplates` | service | `define` + `RenderTemplate` |
| `TestRequireTemplateTextBuilder` | runtime | JS builder chain works |
| `TestRequireTemplateHTMLBuilderEscaping` | runtime | JS observes HTML escaping |
| `TestTemplateUsesGlazedHelpers` | runtime/service | `upper`, `trim`, `toYaml`, padding helpers available |
| `TestTemplateMissingKeyError` | runtime/service | missing data fails when configured |
| `TestTemplateTypeScriptDeclarations` | docs/generator if present | declaration includes new module |

Commands before review:

```bash
cd goja-text
go test ./... -count=1
GOWORK=off go test ./... -count=1
GOWORK=off make lint
cd cmd/goja-text
GOWORK=off go generate
GOWORK=off go build -o ../../dist/goja-text .
```

## Design Decisions

### Decision 1: Use Go-backed builders and template sets

- **Context:** The user explicitly wants to avoid JS maps/objects for the object side, and existing goja-text APIs already expose Go-backed objects.
- **Options considered:** Plain JS object options; one-shot render function only; Go-backed builder and parsed set.
- **Decision:** Use Go-backed builder/config/template-set/result objects as the primary API.
- **Rationale:** This matches sanitize options builders and Markdown/extract Go-backed domain objects. It also centralizes validation and keeps parsed templates reusable.
- **Consequences:** JS methods use PascalCase exported Go method names; docs and TypeScript declarations must be explicit.
- **Status:** proposed.

### Decision 2: Separate text and HTML modes at construction time

- **Context:** `text/template` and `html/template` share concepts but differ materially in escaping and trusted-type behavior.
- **Options considered:** One `template.create({ mode: "html" })`; separate `text()`/`html()`; separate modules `textTemplate` and `htmlTemplate`.
- **Decision:** Export `template.text()` and `template.html()` builders.
- **Rationale:** The call site makes escaping mode visible. One module keeps related docs and helpers together.
- **Consequences:** Service layer needs a small abstraction around text vs HTML template types.
- **Status:** proposed.

### Decision 3: Reuse Glazed helper functions rather than copy them

- **Context:** Glazed already has Sprig plus custom helpers such as formatting, YAML rendering, padding, random helpers, and file/FS parsing helpers.
- **Options considered:** Reimplement helpers in goja-text; import Glazed helper package; provide no helpers by default.
- **Decision:** Import and reuse `glazed/pkg/helpers/templating` function maps/constructors.
- **Rationale:** It avoids drift and preserves go-go-golems template behavior.
- **Consequences:** The dependency is already present indirectly in `goja-text/go.mod`, but implementation may make it a direct dependency.
- **Status:** proposed.

### Decision 4: Defer JS template functions to phase 2

- **Context:** Calling JS functions from Go template execution is desirable but tricky.
- **Options considered:** Implement immediately; reject permanently; design now and defer implementation.
- **Decision:** Document the API shape and risks now, implement after phase-1 rendering is stable.
- **Rationale:** Runtime ownership, synchronous execution, exception translation, and HTML trusted-type behavior need focused review.
- **Consequences:** Phase 1 remains useful and safer; phase 2 has a clear extension seam.
- **Status:** proposed.

### Decision 5: Default to safe missing-key behavior

- **Context:** Automation and report generation fail confusingly when missing data silently renders as `<no value>`.
- **Options considered:** Go default behavior; `missingkey=zero`; `missingkey=error`.
- **Decision:** Prefer `missingkey=error` as module default, or at least expose it prominently and consider a strict builder preset.
- **Rationale:** goja-text is used for structured automation where silent data loss is dangerous.
- **Consequences:** Existing Go template users may need to supply optional fields or choose a less strict policy.
- **Status:** proposed.

## Alternatives Considered

### Alternative: `template.render(source, data, options)` only

This is easy to implement but does not fit the requested builder style and does not represent parsed template sets, function presets, or validation state well. It can exist as convenience sugar, but it should not be the core API.

### Alternative: expose raw Go `*template.Template`

Raw Go templates are not a stable JavaScript API. They expose too much implementation detail and do not unify text/html behavior cleanly. A `TemplateSet` wrapper gives a small, documented surface.

### Alternative: use JavaScript template literals instead of Go templates

That would not satisfy the request for Go `text/template` and `html/template`, would not reuse Glazed helpers, and would not give Go's HTML contextual escaping.

### Alternative: merge into the sanitize module

Templating is a separate concern. A top-level module keeps help, examples, TypeScript declarations, and future callback support cohesive.

## Risks and Review Points

- **Escaping correctness:** HTML mode must rely on `html/template`; helper functions must not accidentally mark untrusted JS data as trusted HTML.
- **Data conversion surprises:** JS maps, arrays, dates, Go-backed objects, and undefined/null values need tests and documentation.
- **Function-name collisions:** Sprig and Glazed helpers may overlap. Merge order must be documented.
- **Missing-key default:** Review whether strict-by-default is acceptable for all intended scripts.
- **File loading:** `ParseFiles` and `ParseFS` intersect with host filesystem access. Do not expose unrestricted file loading until xgoja host policy is clear.
- **Runtime callbacks:** JS functions inside templates require runtime-owner discipline and should receive a separate implementation review.

## Intern Implementation Checklist

1. Read this guide and the referenced files.
2. Start with service tests; do not write the module loader first.
3. Implement the builder/config validation path.
4. Implement text mode rendering.
5. Implement HTML mode rendering and escaping tests.
6. Add Glazed/Sprig function set selection.
7. Add `TemplateSet` metadata methods.
8. Add the goja `NativeModule` loader.
9. Add runtime integration tests that use `engine.New()` and `require("template")`.
10. Add TypeScript declarations.
11. Wire provider and xgoja YAML.
12. Add help docs and an example script.
13. Run all validation commands.
14. Only then consider jsverbs and phase-2 callback work.

## References

- `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/markdown/module.go` — canonical Go-backed AST module and callback traversal pattern.
- `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/sanitize/module.go` — namespace exports and Go-backed options builder examples.
- `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/sanitize/options.go` — builder validation and `Build()` pattern.
- `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/extract/module.go` — simple module loader and Go-backed result pattern.
- `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/xgoja/providers/text/text.go` — provider registration point to update.
- `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/cmd/goja-text/xgoja.yaml` — generated binary module selection to update.
- `/home/manuel/workspaces/2026-06-07/goja-text-templates/go-go-goja/modules/common.go` — `NativeModule` registry contract.
- `/home/manuel/workspaces/2026-06-07/goja-text-templates/go-go-goja/pkg/doc/02-creating-modules.md` — go-go-goja native module tutorial.
- `/home/manuel/workspaces/2026-06-07/goja-text-templates/go-go-goja/cmd/xgoja/doc/04-tutorial-providing-package-and-modules.md` — xgoja provider/module wiring tutorial.
- `/home/manuel/workspaces/2026-06-07/goja-text-templates/glazed/pkg/helpers/templating/templating.go` — reusable templating helpers, Sprig integration, text/html constructors, FS parsing.
- `/home/manuel/workspaces/2026-06-07/goja-text-templates/glazed/pkg/doc/topics/22-templating-helpers.md` — human-facing Glazed templating helper documentation.
