---
Title: Investigation Diary
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
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: glazed/pkg/doc/topics/22-templating-helpers.md
      Note: Human-facing Glazed templating helper documentation read during investigation
    - Path: go-go-goja/cmd/xgoja/doc/04-tutorial-providing-package-and-modules.md
      Note: xgoja provider and modules selection tutorial read during investigation
    - Path: go-go-goja/modules/common.go
      Note: NativeModule registry contract that shaped the diary's module-loading notes
    - Path: go-go-goja/pkg/doc/02-creating-modules.md
      Note: Native module tutorial read during investigation
    - Path: goja-text/README.md
      Note: Top-level module and demo documentation
    - Path: goja-text/cmd/goja-text/xgoja.yaml
      Note: xgoja module selection for require template
    - Path: goja-text/examples/js/template-demo.js
      Note: Runnable template module demo
    - Path: goja-text/pkg/template/builder.go
      Note: Go-backed fluent template builder and validation
    - Path: goja-text/pkg/template/funcs.go
      Note: Glazed and Sprig function-set selection
    - Path: goja-text/pkg/template/module.go
      Note: Goja NativeModule adapter and JS exports
    - Path: goja-text/pkg/template/module_test.go
      Note: Runtime integration tests for require template
    - Path: goja-text/pkg/template/render.go
      Note: Text and HTML template parsing rendering and metadata wrapper
    - Path: goja-text/pkg/template/template_test.go
      Note: Phase-1 service tests
    - Path: goja-text/pkg/template/types.go
      Note: Phase-1 template service result config and metadata types
    - Path: goja-text/pkg/template/typescript.go
      Note: TypeScript declarations for the template module
    - Path: goja-text/pkg/xgoja/providers/text/doc/template-api-reference.md
      Note: Template module API help page
    - Path: goja-text/pkg/xgoja/providers/text/doc/template-user-guide.md
      Note: Template module user guide
    - Path: goja-text/pkg/xgoja/providers/text/text.go
      Note: Provider wiring for the template module
ExternalSources: []
Summary: Chronological diary for the GOJA-TEXT-004 template module design investigation.
LastUpdated: 2026-06-07T16:20:00-04:00
WhatFor: Use to resume or review the template module design-ticket investigation.
WhenToUse: Before continuing implementation work or checking why design choices were made.
---





# Diary

## Goal

This diary captures the creation of GOJA-TEXT-004 and the evidence-based design work for adding Go `text/template` and `html/template` support to `goja-text` as a goja JavaScript API.

## Step 1: Create the ticket and research the existing module patterns

I created a new docmgr ticket for the template-module work, then inspected the existing `goja-text`, `go-go-goja`, and `glazed` repositories to ground the design in real code. The investigation showed that `goja-text` already has a clear pattern for Go-backed JavaScript APIs: modules implement `modules.NativeModule`, use `modules.SetExport`, register in `init()`, and expose Go-backed builders or domain objects to scripts.

The main design conclusion from this step is that the template API should not be a loose JavaScript object wrapper. It should expose fluent Go-backed builders and parsed template sets, reusing Glazed's template helper package for Sprig and custom helper functions while keeping future JavaScript callback support as a separate phase.

### Prompt Context

**User prompt (verbatim):**

```text
Create a new docmgr ticket to add go text/template and html/template support to goja-text/ as goja JS API (elegant, builder pattern, fluid, using go for the object side to avoir JS maps and objects, see other goja-text APIs).

Read the relevant go-go-goja docs, and there's a afair amount of templating helpers in glazed/ as well. We also in a more advanced phase want to export JS functions to the template renderer too.

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.
```

**Assistant interpretation:** Create a new structured docmgr ticket for the goja-text template-module feature, research the adjacent code and docs, write a detailed intern-oriented implementation guide, store it in the ticket, and upload the resulting documentation bundle to reMarkable.

**Inferred user intent:** Provide an implementable, evidence-backed design package so a new intern can build the template module correctly without rediscovering the goja-text module conventions, xgoja provider wiring, or Glazed templating helper APIs.

**Commit (code):** N/A — documentation and ticket workspace only.

### What I did

- Ran `docmgr status --summary-only` to locate the docmgr root for this workspace.
- Added missing vocabulary topics for `templating` and `html`.
- Created ticket `GOJA-TEXT-004` titled `Goja Text Template and HTML Template JavaScript API`.
- Added a design document and an investigation diary under the ticket.
- Added five ticket tasks covering research, design writing, relations, and reMarkable upload.
- Read and cross-checked key files:
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/markdown/module.go`
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/sanitize/module.go`
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/sanitize/options.go`
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/extract/module.go`
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/xgoja/providers/text/text.go`
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/cmd/goja-text/xgoja.yaml`
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/go-go-goja/modules/common.go`
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/go-go-goja/pkg/doc/02-creating-modules.md`
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/go-go-goja/cmd/xgoja/doc/04-tutorial-providing-package-and-modules.md`
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/glazed/pkg/helpers/templating/templating.go`
  - `/home/manuel/workspaces/2026-06-07/goja-text-templates/glazed/pkg/doc/topics/22-templating-helpers.md`
- Wrote the design guide with current-state architecture, API proposals, diagrams, pseudocode, implementation plan, tests, risks, and decision records.

### Why

- The feature crosses three systems: `goja-text` native modules, `go-go-goja`/xgoja module composition, and Glazed template helper functions.
- The user requested an intern-ready guide, so the document needs enough system background to be useful before implementation begins.
- The request specifically asked for Go-backed objects and future JavaScript template functions, which are design-sensitive enough to document before coding.

### What worked

- Existing `goja-text` modules provide strong examples for the requested style:
  - Markdown exposes Go-backed AST objects and callback traversal.
  - Sanitize exposes Go-backed option builders.
  - Extract exposes Go-backed candidates and options.
- Glazed already has reusable `text/template` and `html/template` helper constructors with Sprig and Glazed-specific function maps.
- xgoja provider docs clearly show that implementation must update both provider registration and the buildspec `modules` list.

### What didn't work

- No implementation was attempted in this step, so there were no compiler or test failures.
- The initial generated diary document used the generic reference template, so I replaced it with the stricter diary format required for implementation continuation.

### What I learned

- `goja-text` intentionally exposes Go-backed objects with exported Go field and method names in JavaScript; the template module should document this instead of hiding it.
- The xgoja provider boundary matters: importing/registering the module package is not sufficient unless the generated buildspec also selects the module.
- JavaScript functions inside Go templates are feasible but should be deferred because Go template execution is synchronous and goja runtime ownership/error translation need careful review.

### What was tricky to build

- The main tricky part was choosing an API that is fluid in JavaScript without making JS maps the source of truth. The solution proposed in the design is to keep state in Go-backed builders, configs, parsed template sets, and render result objects, while allowing plain JS data as render input after explicit Go-side normalization.
- Another tricky point is naming: Go-backed methods appear as PascalCase methods in JavaScript. This is less idiomatic JS but matches the existing `sanitize.yaml.options().MaxIterations(5).Build()` pattern, so the guide calls it out as a deliberate tradeoff.

### What warrants a second pair of eyes

- Whether the module name should be `template`, `templates`, or separate `textTemplate`/`htmlTemplate` aliases.
- Whether the default function set should include Glazed and Sprig automatically or require explicit `.Funcs("glazed", "sprig")`.
- Whether the default missing-key policy should be strict `missingkey=error`.
- The exact phase-2 design for JavaScript callbacks inside templates.

### What should be done in the future

- Implement phase 1 from the checklist in the design guide.
- Add a focused design addendum before implementing JavaScript callback functions in template renderers.
- Consider a small spike for `goja.Value` export behavior with Go templates before finalizing data-conversion documentation.

### Code review instructions

- Start with the design guide at `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/ttmp/2026/06/07/GOJA-TEXT-004--goja-text-template-and-html-template-javascript-api/design-doc/01-template-module-analysis-design-and-implementation-guide.md`.
- Verify that every major recommendation points to concrete repository files.
- Validate docmgr hygiene with `docmgr doctor --ticket GOJA-TEXT-004 --stale-after 30`.

### Technical details

Relevant command examples from this step:

```bash
docmgr status --summary-only
docmgr vocab add --category topics --slug templating --description "Go text/template and html/template rendering, helpers, and template API design"
docmgr vocab add --category topics --slug html --description "HTML rendering and escaping semantics"
docmgr ticket create-ticket --ticket GOJA-TEXT-004 --title "Goja Text Template and HTML Template JavaScript API" --topics goja,goja-bindings,native-modules,text-algorithms,templating,html,xgoja,jsverbs,cli
docmgr doc add --ticket GOJA-TEXT-004 --doc-type design-doc --title "Template Module Analysis Design and Implementation Guide"
docmgr doc add --ticket GOJA-TEXT-004 --doc-type reference --title "Investigation Diary"
```

## Step 2: Implement the phase-1 Go template service layer

I added the first implementation slice: a Go-backed template service package under `pkg/template`. This package now has the domain types, fluent builders, validation, function-set selection, text/html parsing, rendering, named-template metadata, and service-level tests. The code is intentionally independent of xgoja provider wiring and mostly independent of goja runtime concerns so that the core renderer can be reviewed and tested before JavaScript module plumbing is added.

This step implements the design-guide recommendation that the JavaScript API should be backed by stable Go objects. `TemplateBuilder`, `TemplateConfig`, `TemplateSet`, `TemplateInfo`, and `RenderResult` now exist as real Go structs and methods, while Glazed and Sprig helpers are merged through named function sets.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Start implementing the ticket in phases, record detailed progress, and commit coherent checkpoints.

**Inferred user intent:** Move from design documentation to working code while preserving a continuation-friendly record of decisions, failures, and validation.

**Commit (code):** d14129b21f530d90637fc9bb9ae0a32c29bf5299 — "Add template service design and phase one implementation"

### What I did

- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/types.go` with `Mode`, `TemplateConfig`, `ValidationResult`, `RenderResult`, and `TemplateInfo`.
- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/builder.go` with `TemplateBuilder`, `Name`, `Funcs`, `MissingKey`, `Delims`, `Validate`, `BuildConfig`, `Parse`, and `ParseNamed`.
- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/funcs.go` with `sprig`, `glazed`, and `none` function-set handling.
- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/render.go` with `TemplateSet`, text/html parsing, `Render`, `RenderTemplate`, `RenderString`, `Templates`, `Lookup`, `RenderText`, and `RenderHTML`.
- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/template_test.go` covering text rendering, HTML escaping, named templates, missing-key errors, builder validation, and convenience rendering.
- Ran `gofmt -w goja-text/pkg/template`.
- Ran `cd goja-text && go test ./pkg/template -count=1`.

### Why

- The service layer is the safest first slice because it proves the template behavior before goja reflection, xgoja provider registration, generated command docs, or jsverbs are involved.
- Keeping the renderer in Go-backed objects matches existing `goja-text` builder/result patterns and the user's explicit request to avoid making JS maps the object model.

### What worked

- `go test ./pkg/template -count=1` passed.
- Glazed helper integration works through `glazedtemplating.TemplateFuncs` and Sprig integration works through `sprig.TxtFuncMap()` / `sprig.HtmlFuncMap()`.
- `html/template` contextual escaping is visible in tests: untrusted names are escaped and `javascript:` URLs do not pass through as raw links.
- Strict `missingkey=error` behavior is implemented as the default and tested.

### What didn't work

- My first `RenderText`/`RenderHTML` convenience implementation attempted to chain through `NewTextBuilder().Parse(source).Render(data)`, but `Parse` returns `(*TemplateSet, error)`, so that shape cannot compile. I fixed it by splitting parse and render into explicit steps:

```go
set, err := NewTextBuilder().Parse(source)
if err != nil { return nil, err }
return set.Render(data)
```

### What I learned

- The phase-1 abstraction is small enough that text and HTML engines can share one `TemplateSet` wrapper while keeping exactly one underlying template pointer initialized.
- Function-set merge order is now deterministic because helper names are sorted before merging each map. Later maps override earlier maps within the selected preset order.
- The `none` preset needs explicit validation because combining `none` with `sprig` or `glazed` is ambiguous.

### What was tricky to build

- The tricky part was creating a unified wrapper over `text/template` and `html/template` without losing their distinct escaping behavior. The implementation stores either `text *texttemplate.Template` or `html *htmltemplate.Template` and switches on `Mode` during parsing, rendering, and metadata lookup.
- Another subtle point was the missing-key policy. Go templates expect the option string form `missingkey=<policy>`, so the builder stores a user-facing policy string but `parseTemplateSet` constructs the exact Go option when creating the template.

### What warrants a second pair of eyes

- Confirm that defaulting to `missingkey=error` is acceptable for all intended `goja-text` template use cases.
- Review function-set merge order and whether Glazed helpers should override Sprig helpers or vice versa.
- Review package path/name `pkg/template`; it is clear for users but requires aliasing standard-library imports inside the package.

### What should be done in the future

- Add the goja `NativeModule` adapter and runtime integration tests next.
- Exercise JS object data conversion once the module adapter exists.
- Consider additional service tests for delimiter customization and `Funcs("none")`.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/builder.go` to understand the fluent API and validation.
- Then review `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/render.go` for text vs HTML execution behavior.
- Validate with:

```bash
cd goja-text
go test ./pkg/template -count=1
```

### Technical details

The current phase-1 API supports Go usage like:

```go
set, err := template.NewTextBuilder().Funcs("sprig", "glazed").Parse("Hello {{ .Name | upper }}")
result, err := set.Render(map[string]any{"Name": "intern"})
fmt.Println(result.Text) // Hello INTERN
```

## Step 3: Add the goja native module adapter and runtime tests

I added the JavaScript-facing module adapter for `require("template")`. The module now exports `text()`, `html()`, `renderText()`, and `renderHTML()` through the go-go-goja `NativeModule` interface, and it declares TypeScript metadata matching the Go-backed builder and template-set surface.

This step proves that the phase-1 service objects are usable from a real goja runtime. The runtime tests exercise the intended fluent builder API, HTML contextual escaping, named template rendering, convenience rendering, helper functions, and validation errors as JavaScript sees them.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue implementing the task list one phase at a time, preserving test evidence and diary context.

**Inferred user intent:** Turn the design into an actual native module that follows existing goja-text conventions and can be reviewed incrementally.

**Commit (code):** 9da89ead6c79d3ac2b4547924abe478ab85c686a — "Expose template as a goja native module"

### What I did

- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/module.go` implementing `modules.NativeModule`, `Doc()`, `Loader`, and `init()` registration.
- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/typescript.go` implementing `modules.TypeScriptDeclarer` with RawDTS for builders, config, template sets, render results, and exported functions.
- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/module_test.go` with goja runtime integration tests using `engine.NewRuntimeFactoryBuilder().UseModuleMiddleware(engine.MiddlewareOnly("template"))`.
- Ran `gofmt -w goja-text/pkg/template`.
- Ran `cd goja-text && go test ./pkg/template -count=1`.

### Why

- The module adapter is the point where the Go-backed service becomes a goja JavaScript API.
- Runtime tests are necessary because method reflection, object export, JavaScript object data, and error translation can behave differently from pure Go service tests.

### What worked

- JavaScript can now call `require("template")` in a go-go-goja runtime.
- The fluent chain works from JS:

```js
template.text().Name("greeting").Funcs("sprig", "glazed").Parse("Hello {{ .Name | upper }}").Render({ Name: "intern" })
```

- JavaScript object data passed into Go-backed `TemplateSet.Render(...)` rendered correctly.
- HTML mode escaped untrusted text and unsafe JavaScript URLs in the runtime test.
- Validation errors from builders propagate as JavaScript execution errors.

### What didn't work

- The first runtime test compared `got["mode"] != "text"`, but the exported value retained the Go `Mode` type rather than a plain string. The failure looked like this:

```text
--- FAIL: TestRequireTemplateTextBuilder (0.00s)
    module_test.go:36: unexpected result: map[string]interface {}{"bytes":12, "mode":"text", "name":"greeting", "text":"Hello INTERN"}
FAIL
```

The printed value looked like a string, but the dynamic type was not exactly `string`. I fixed the assertion with `fmt.Sprint(got["mode"]) == "text"`.

### What I learned

- goja can pass a JavaScript object into a reflected Go method accepting `any` well enough for Go templates to read map keys with `{{ .Name }}`.
- Type aliases like `Mode string` may remain distinct Go types after `Export()`, so tests should avoid overly strict dynamic-type assumptions when checking exported values.
- The existing `MiddlewareOnly("template")` pattern is enough to test module loading when the package is blank-imported in the test.

### What was tricky to build

- The main sharp edge was data conversion across reflected Go-backed methods. The top-level convenience exports accept `goja.Value` and explicitly call `Export()`, but methods on `TemplateSet` are invoked by goja reflection. The runtime tests confirmed that the reflected path is currently sufficient for normal object data.
- Another tricky point was ensuring the module adapter stayed small. It only creates builders and convenience render functions; all parsing and rendering remains in the service layer.

### What warrants a second pair of eyes

- Review whether `TemplateSet.Render(data any)` should proactively normalize `goja.Value` in case future code calls it from Go with raw goja values.
- Confirm the TypeScript declaration naming and whether `RenderString` should be part of the public JS API.
- Review whether top-level convenience helpers should return `RenderResult` or bare strings.

### What should be done in the future

- Wire the module into the xgoja provider and generated command buildspec.
- Add user-facing help pages and examples.
- Add tests covering delimiter customization and `Funcs("none")` if time permits.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/module.go` to inspect module exports.
- Then read `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/template/module_test.go` for the JavaScript contract.
- Validate with:

```bash
cd goja-text
go test ./pkg/template -count=1
```

### Technical details

The goja module now exposes this minimal API:

```js
const template = require("template");
const result = template.renderText("Hello {{ .Name }}", { Name: "Ada" });
console.log(result.Text);
```

Additional validation note for Step 3: the first commit attempt failed in the pre-commit lint hook because `normalizeTemplateData` in `pkg/template/module.go` was unused. The exact lint finding was:

```text
pkg/template/module.go:55:6: func normalizeTemplateData is unused (unused)
func normalizeTemplateData(value any) any {
     ^
```

I removed the unused helper because the current adapter only needs explicit `exportTemplateData` for top-level convenience functions; reflected Go-backed methods are covered by runtime tests.

## Step 4: Wire xgoja, add docs/examples, and validate the generated binary

I wired the template module into the generated `goja-text` binary path and added user-facing documentation. The xgoja provider now imports and exposes the `template` module, the buildspec selects it as `require("template")`, and the provider help bundle includes an API reference and user guide. I also added a runnable `examples/js/template-demo.js` script and updated the root README module list and help examples.

This step completes the product-facing path: a user can build `dist/goja-text`, run a JavaScript file that imports `template`, and read help pages for the new API inside the generated binary.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue the implementation by wiring the module into xgoja and making it discoverable through docs and examples.

**Inferred user intent:** Ensure the feature is not only implemented as a package, but also available and teachable in the shipped goja-text command.

**Commit (code):** 59959280ebf746604cf29b0ebc3309e76352949e — "Wire template module into goja-text command"

### What I did

- Updated `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/xgoja/providers/text/text.go` to blank-import `pkg/template`, add `template` to `textModuleNames`, and update the help-source description.
- Updated `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/cmd/goja-text/xgoja.yaml` to select the new `template` module as `require("template")`.
- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/xgoja/providers/text/doc/template-api-reference.md`.
- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/xgoja/providers/text/doc/template-user-guide.md`.
- Added `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/examples/js/template-demo.js`.
- Updated `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/README.md` with the new module, help pages, and demo script.
- Ran xgoja generation and built the generated binary.
- Ran the new demo through `dist/goja-text`.
- Ran repository tests and lint in normal and `GOWORK=off` modes.

### Why

- xgoja has two boundaries: provider registration and buildspec module selection. Both must be updated for JavaScript users to `require("template")` in the generated command.
- Help docs and examples are part of the goja-text pattern for making each native module self-teaching.

### What worked

- `GOTOOLCHAIN=go1.26.4 GOWORK=off go generate` in `cmd/goja-text` succeeded after running `go mod tidy`.
- `GOTOOLCHAIN=go1.26.4 GOWORK=off go build -o ../../dist/goja-text .` succeeded.
- `../../dist/goja-text run ../../examples/js/template-demo.js` succeeded and printed text, named-template, HTML-escaped, and convenience-render output.
- Validation passed:

```bash
cd goja-text
go test ./... -count=1
GOTOOLCHAIN=go1.26.4 GOWORK=off go test ./... -count=1
GOTOOLCHAIN=go1.26.4 GOWORK=off make lint
```

### What didn't work

- `cd goja-text/cmd/goja-text && GOWORK=off go generate` failed because the local `go` command reported Go 1.26.1 while the root module requires Go 1.26.4:

```text
go: module ../.. requires go >= 1.26.4 (running go 1.26.1)
```

- Running `go generate` without `GOWORK=off` from the nested generated module failed because the workspace treated the root module as the main module:

```text
main module (github.com/go-go-golems/goja-text) does not contain package github.com/go-go-golems/goja-text/cmd/goja-text
```

- `GOTOOLCHAIN=go1.26.4 GOWORK=off go generate` then reported that the nested module needed tidying:

```text
go: updates to go.mod needed; to update it:
	go mod tidy
```

I fixed this with:

```bash
cd goja-text/cmd/goja-text
GOTOOLCHAIN=go1.26.4 GOWORK=off go mod tidy
GOTOOLCHAIN=go1.26.4 GOWORK=off go generate
```

### What I learned

- The generated command module now tracks `go 1.26.4`, which matches the root module requirement and avoids the local toolchain mismatch.
- The dry-run xgoja generation reported `modules=7`, which reflects the added template module alongside the existing goja-text/core/host modules.
- The generated binary can exercise the template module without adding a jsverb yet.

### What was tricky to build

- The tricky part was the nested module/toolchain interaction. The generated command lives under `cmd/goja-text` with its own `go.mod`, but the workspace root also has a `go.work`. For generation, the reliable command shape was `GOTOOLCHAIN=go1.26.4 GOWORK=off ...` from the nested module.
- Another subtle point is documentation discoverability. The provider help docs are embedded by a glob in `pkg/xgoja/providers/text/doc/doc.go`, so adding Markdown files there is enough, but the provider description should still mention the new module.

### What warrants a second pair of eyes

- Review the generated `cmd/goja-text/go.mod` change from `go 1.26.1` to `go 1.26.4`; it appears correct because the root module already requires 1.26.4.
- Review the help docs for exact terminology around `html/template` escaping and Go-backed PascalCase method names.
- Decide whether to add a first-class `template` jsverb command or keep the example script only for this phase.

### What should be done in the future

- Run `docmgr doctor` and update changelog/relations for the final state.
- Commit the xgoja/docs/examples checkpoint.
- Optionally upload the updated ticket bundle to reMarkable again after final validation.

### Code review instructions

- Start with `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/xgoja/providers/text/text.go` and `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/cmd/goja-text/xgoja.yaml` to verify module exposure.
- Review `/home/manuel/workspaces/2026-06-07/goja-text-templates/goja-text/pkg/xgoja/providers/text/doc/template-user-guide.md` and `template-api-reference.md` for user-facing API clarity.
- Validate the generated binary with:

```bash
cd goja-text/cmd/goja-text
GOTOOLCHAIN=go1.26.4 GOWORK=off go build -o ../../dist/goja-text .
../../dist/goja-text run ../../examples/js/template-demo.js
```

### Technical details

The generated binary demo printed HTML output containing Go's `html/template` unsafe-URL sentinel:

```json
"HTML": "<p>&lt;Ada&gt;</p><a href=\"#ZgotmplZ\">open</a>"
```

This confirms the demo is using `html/template` contextual escaping rather than plain string substitution.
