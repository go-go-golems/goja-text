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
    - Path: goja-text/pkg/template/builder.go
      Note: Go-backed fluent template builder and validation
    - Path: goja-text/pkg/template/funcs.go
      Note: Glazed and Sprig function-set selection
    - Path: goja-text/pkg/template/render.go
      Note: Text and HTML template parsing rendering and metadata wrapper
    - Path: goja-text/pkg/template/template_test.go
      Note: Phase-1 service tests
    - Path: goja-text/pkg/template/types.go
      Note: Phase-1 template service result config and metadata types
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

**Commit (code):** pending — phase-1 service implementation checkpoint.

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
