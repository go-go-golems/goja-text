---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: go-go-goja/cmd/xgoja/internal/generate/gomod.go
      Note: xgoja generated go.mod and replace behavior
    - Path: go-go-goja/modules/uidsl/node.go
      Note: Reference Go-backed domain object shape
    - Path: go-go-goja/modules/yaml/yaml.go
      Note: Reference native module pattern
    - Path: go-go-goja/pkg/xgoja/providers/core/core.go
      Note: xgoja NativeModule provider wrapper pattern
    - Path: goja-text/go.mod
      Note: Placeholder module path and dependency update need
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---






# Research Logbook: Sources, Usefulness, and Update Needs

## Goal

This logbook tracks which documents, source files, examples, and workflow references shaped the `GOJA-TEXT-001` goja-text design work. It is meant to help future contributors quickly answer: "Why did we read this? Was it useful? Is it still reliable? What should be updated before implementation?"

## Context

The research goal was to design goja bindings for text algorithms, starting with a Markdown parser module that exposes Go-backed AST objects to JavaScript. The design evolved from a generic native module plan into an xgoja-based testing/build plan with a `walk()` traversal primitive and Go-backed `MarkdownNode` objects.

This logbook records resources read during the research/design/review passes. It includes repository source files, generated-binary infrastructure, existing module patterns, ticket-management workflow docs, and goja runtime documentation discovered locally in the Go module cache.

## Legend

- **Useful**: Keep using this as evidence or implementation guidance.
- **Partly useful**: Useful for context, but not sufficient by itself.
- **Needs update**: The source or our derived docs need correction before implementation.
- **Out of date / wrong**: The source conflicted with the final project decision or was misleading for this ticket.

---

## 1. Workspace and Project Topology

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/go.work`

- **What I was researching:** How the local workspace resolves `go-go-goja`, `goja-text`, `glazed`, and `sanitize` together.
- **What I was looking for:** Whether `goja-text` could import local `go-go-goja` packages without publishing a module.
- **Why I chose it / what led me to it:** The user said everything is included in this repository/workspace; `go.work` is the authoritative Go workspace file.
- **Useful findings:** Confirms the workspace contains `./glazed`, `./go-go-goja`, `./goja-text`, and `./sanitize`.
- **Not useful:** Does not help generated xgoja temporary modules resolve local provider imports.
- **Out of date / wrong:** None in the file itself.
- **Needs updating:** Nothing in `go.work`; the design docs needed to clarify that xgoja generated builds still need package-level `replace: .` for the local `goja-text` provider.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/go.mod`

- **What I was researching:** The starting state of the new `goja-text` module.
- **What I was looking for:** Module path, Go version, and dependencies.
- **Why I chose it / what led me to it:** New implementation will live in `goja-text`, so `go.mod` determines import paths.
- **Useful findings:** The module currently exists as a scaffold.
- **Not useful:** Current dependencies are mostly template leftovers and do not describe the final module.
- **Out of date / wrong:** Module path is `github.com/go-go-golems/XXX`, which is wrong for the intended design.
- **Needs updating:** Change module path to `github.com/go-go-golems/goja-text`; add dependencies for `github.com/go-go-golems/go-go-goja` and `github.com/yuin/goldmark`; ensure local xgoja builds can resolve it with `replace: .`.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/README.md`

- **What I was researching:** Whether `goja-text` already contained project-specific documentation.
- **What I was looking for:** Existing purpose, CLI shape, or constraints.
- **Why I chose it / what led me to it:** README is normally the entrypoint for a new module.
- **Useful findings:** Confirmed `goja-text` is currently a template/scaffold, not a fleshed-out project.
- **Not useful:** The ASCII-template content does not describe goja text bindings.
- **Out of date / wrong:** It is effectively placeholder content for this project.
- **Needs updating:** Replace with a real README after Phase 1: explain xgoja build, markdown module, examples, and smoke tests.

---

## 2. Native Module System

### Resource: `go-go-goja/modules/common.go`

- **What I was researching:** The core native module interface and registry behavior.
- **What I was looking for:** How Go modules become `require()` modules in goja.
- **Why I chose it / what led me to it:** The design begins with a Go-backed `markdown` module; this file defines `NativeModule` and `modules.Register()`.
- **Useful findings:** `NativeModule` requires `Name()`, `Doc()`, and `Loader(*goja.Runtime, *goja.Object)`. `DefaultRegistry` stores registered modules and `Enable()` registers each loader with `goja_nodejs/require.Registry`.
- **Not useful:** Does not explain xgoja provider wrapping; that lives elsewhere.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update needed. The design should keep referencing this as the core contract.

### Resource: `go-go-goja/modules/exports.go`

- **What I was researching:** How existing modules attach exports safely.
- **What I was looking for:** Utility function for module export registration.
- **Why I chose it / what led me to it:** The YAML module uses `modules.SetExport`; new markdown module should follow local style.
- **Useful findings:** `SetExport()` wraps `exports.Set()` and logs export-registration errors.
- **Not useful:** It does not validate function signatures or provide runtime argument checking.
- **Out of date / wrong:** None observed.
- **Needs updating:** No update needed.

### Resource: `go-go-goja/modules/typing.go`

- **What I was researching:** TypeScript declaration hooks for native modules.
- **What I was looking for:** How a module can describe its JS API.
- **Why I chose it / what led me to it:** The design originally used `spec.Any()` heavily; later review identified stronger typings as useful.
- **Useful findings:** Modules can implement `TypeScriptDeclarer` with `TypeScriptModule() *spec.Module`.
- **Not useful:** Does not itself document best practices for Go-backed struct field names.
- **Out of date / wrong:** None observed.
- **Needs updating:** No code update needed; markdown module should implement this interface with `MarkdownNode`, `WalkContext`, and `ValidationResult` declarations.

### Resource: `go-go-goja/modules/yaml/yaml.go`

- **What I was researching:** A simple, production-like native module pattern.
- **What I was looking for:** Loader structure, error handling, docs, TypeScript declaration example, tests to emulate.
- **Why I chose it / what led me to it:** YAML is the closest existing data parser module.
- **Useful findings:** Great reference for stateless module struct, compile-time interface checks, `Doc()`, `TypeScriptModule()`, `Loader()`, `init()` registration, and function export style.
- **Not useful:** YAML returns generic maps/slices; it is not a good model for Go-backed domain objects with validation behavior.
- **Out of date / wrong:** Not wrong; just a different output-shape pattern.
- **Needs updating:** No update needed. Future docs should explicitly contrast YAML's generic-data pattern with Markdown's Go-backed object pattern.

### Resource: `go-go-goja/modules/uidsl/module.go` and `go-go-goja/modules/uidsl/node.go`

- **What I was researching:** How existing modules expose Go structs or domain objects into JavaScript.
- **What I was looking for:** Evidence that Go-backed objects are a valid project pattern.
- **Why I chose it / what led me to it:** The Markdown AST should remain a Go object graph; `uidsl` is an existing module that returns Go domain objects.
- **Useful findings:** `ui.dsl` creates `Element`, `Text`, `Fragment`, and other Go structs and passes them through goja. This supports the final decision to keep `MarkdownNode` as a Go-backed object.
- **Not useful:** `uidsl` is primarily an HTML builder, not a parser; it does not answer goldmark AST conversion details.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update needed. The design doc should keep this as the strongest local precedent for Go-backed domain objects.

### Resource: `go-go-goja/modules/fs/fs.go` and `go-go-goja/modules/fs/backend.go`

- **What I was researching:** How filesystem access works inside goja modules.
- **What I was looking for:** Whether xgoja can expose `require("fs")` to load Markdown files from disk.
- **Why I chose it / what led me to it:** The user specifically wanted xgoja because it can load `fs` for reading files.
- **Useful findings:** `fs` supports configurable backends; host provider can instantiate it with `OSBackend{}` when explicitly allowed.
- **Not useful:** Direct `modules/fs` source does not show xgoja safety gating; host provider does.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update needed. Design docs should mention that `fs` access comes from the xgoja host provider, not from the markdown module.

---

## 3. Engine and Runtime Composition

### Resource: `go-go-goja/engine/factory.go`

- **What I was researching:** Runtime creation and module composition.
- **What I was looking for:** How an `engine.Factory` creates a `goja.Runtime`, `require` registry, event loop, owner, and registered modules.
- **Why I chose it / what led me to it:** Native module testing and xgoja runtimes ultimately use the engine factory.
- **Useful findings:** Confirms the builder/factory/runtime lifecycle and where modules are registered. Also shows runtime services and lifecycle context setup.
- **Not useful:** xgoja uses its own `RuntimeFactory` wrapper, so this file alone is not enough for generated binaries.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update needed.

### Resource: `go-go-goja/engine/runtime.go`

- **What I was researching:** Runtime lifecycle and default module imports.
- **What I was looking for:** Which modules are globally blank-imported and how runtime close/owner cleanup works.
- **Why I chose it / what led me to it:** Initial design considered `MiddlewareAdd("markdown")`; this file shows default-registry module behavior.
- **Useful findings:** Default builder imports many built-in modules via blank imports and closes runtime resources explicitly.
- **Not useful:** For goja-text, we should not edit `engine/runtime.go`; xgoja provider composition is better.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update needed. The design doc was updated away from editing runtime.go.

### Resource: `go-go-goja/engine/module_middleware.go`

- **What I was researching:** Default registry filtering.
- **What I was looking for:** Whether `MiddlewareAdd`, `MiddlewareOnly`, and `MiddlewareSafe` were the right mechanism for goja-text.
- **Why I chose it / what led me to it:** Initial approach considered a custom CLI using middleware.
- **Useful findings:** Good context for goja-repl/custom runtime behavior.
- **Not useful:** Less central after switching to xgoja; xgoja disables implicit default modules and loads explicit provider modules.
- **Out of date / wrong:** Not wrong, but not the final primary path.
- **Needs updating:** No code update. Design should treat this as background only.

### Resource: `go-go-goja/engine/module_specs.go` and `go-go-goja/engine/runtime_modules.go`

- **What I was researching:** Runtime-aware module registration.
- **What I was looking for:** How module specs receive runtime context and register loaders.
- **Why I chose it / what led me to it:** Needed to understand how xgoja provider modules become engine modules.
- **Useful findings:** `RuntimeModuleSpec` is the lower-level mechanism xgoja uses to register provider modules.
- **Not useful:** Does not expose xgoja YAML/config behavior directly.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update needed.

---

## 4. goja Runtime Reflection and Field Names

### Resource: local goja module cache — `README.md`, `object_goreflect.go`, `object_goreflect_test.go`

- **What I was researching:** Whether JS should access returned Go struct fields as `node.Type` or `node.type`.
- **What I was looking for:** goja field-name mapping behavior and how `json` tags interact with reflected Go structs.
- **Why I chose it / what led me to it:** The review identified that the original plan assumed lowercase JS property names from `json` tags. The user then asked whether `node.Type` would work.
- **Useful findings:** goja exposes exported struct fields by default and supports `SetFieldNameMapper`, `TagFieldNameMapper("json", true)`, and `UncapFieldNameMapper()` when different property naming is desired.
- **Not useful:** The goja docs are generic; they do not decide our project-level API shape.
- **Out of date / wrong:** The earlier design assumption that lowercase `node.type` was the primary API was wrong for the final project direction.
- **Needs updating:** Primary design doc was updated to specify Go-backed objects and exported Go field names such as `node.Type` and `node.Children`. Implementation should add a small test proving this behavior.

---

## 5. xgoja Build System and Provider Model

### Resource: `go-go-goja/cmd/xgoja/root.go` and `go-go-goja/cmd/xgoja/cmd_build.go`

- **What I was researching:** How xgoja builds custom binaries.
- **What I was looking for:** CLI flags, build workflow, `--xgoja-replace`, generated workspace behavior.
- **Why I chose it / what led me to it:** User requested using xgoja to test the setup.
- **Useful findings:** `xgoja build` validates spec, generates files, runs `go mod tidy`, and builds output. `--xgoja-replace` points generated builds at a local `go-go-goja` checkout.
- **Not useful:** Does not explain provider runtime behavior; app/provider files do.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update. Design docs should continue to include the build command with `--xgoja-replace ../go-go-goja`.

### Resource: `go-go-goja/cmd/xgoja/internal/buildspec/spec.go`, `load.go`, and `validate.go`

- **What I was researching:** The `xgoja.yaml` schema and validation rules.
- **What I was looking for:** Required fields, defaults, package `replace`, command runtime requirements, jsverbs sources.
- **Why I chose it / what led me to it:** Needed to produce an implementation-ready `xgoja.yaml` for `goja-text`.
- **Useful findings:** Package specs support `replace`; enabled commands require a runtime; defaults fill `go.version`, `go.module`, target kind/output, and package `Register` function.
- **Not useful:** Validation checks structural spec correctness, not whether provider modules exist at runtime.
- **Out of date / wrong:** The earlier design omitted `replace: .` for the local `goja-text` provider; that omission was wrong for local generated builds.
- **Needs updating:** Primary design doc has been updated to include `replace: .`. Implementation should run `xgoja build --dry-run` before full build.

### Resource: `go-go-goja/cmd/xgoja/internal/generate/gomod.go`

- **What I was researching:** How generated xgoja builds resolve dependencies.
- **What I was looking for:** Whether local provider packages need package-level `replace` directives.
- **Why I chose it / what led me to it:** The review identified generated-module dependency resolution as a likely gap.
- **Useful findings:** Generated `go.mod` always requires `github.com/go-go-golems/go-go-goja`; package versions/replaces are emitted from package specs. `providerModulePath()` derives module roots from provider import paths.
- **Not useful:** Does not validate that the replacement points to the right local module; build will reveal that.
- **Out of date / wrong:** None in source; our earlier spec needed updating.
- **Needs updating:** No source update. The ticket design now includes `replace: .`.

### Resource: `go-go-goja/cmd/xgoja/internal/generate/templates/main.go.tmpl`

- **What I was researching:** Generated binary bootstrap behavior.
- **What I was looking for:** How provider packages are imported and registered.
- **Why I chose it / what led me to it:** Needed to understand how `goja-text/pkg/xgoja/providers/text.Register` will be called.
- **Useful findings:** Generated `main.go` imports provider packages, constructs a `providerapi.Registry`, calls each provider's `Register`, then builds the app root.
- **Not useful:** Does not describe individual module behavior.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update needed.

### Resource: `go-go-goja/pkg/xgoja/app/root.go`, `host.go`, `factory.go`, `run.go`, and `tui.go`

- **What I was researching:** Runtime behavior of generated xgoja binaries.
- **What I was looking for:** How generated commands create runtimes, run scripts, start the TUI, and use provider modules.
- **Why I chose it / what led me to it:** Needed to replace the custom CLI plan with xgoja-based `eval`, `run`, and `repl`.
- **Useful findings:** `RuntimeFactory` creates engine runtimes with implicit default modules disabled. `run` adds module roots from the script path. `repl` starts the Bubble Tea JS REPL backed by an xgoja runtime.
- **Not useful:** Does not tell us how to write the markdown parser itself.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update. The design now treats xgoja as the primary harness.

### Resource: `go-go-goja/pkg/xgoja/providerapi/module.go` and `registry.go`

- **What I was researching:** Provider API contracts.
- **What I was looking for:** How to expose a new `goja-text` provider package.
- **Why I chose it / what led me to it:** xgoja provider model is the integration point for generated binaries.
- **Useful findings:** Provider packages register `providerapi.Module` entries with `Name`, `DefaultAs`, `Description`, optional `ConfigSchema`, and a `New(ModuleContext)` factory returning a `require.ModuleLoader`.
- **Not useful:** Does not provide a convenience wrapper for `modules.NativeModule`; core provider shows that pattern.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update. Implement `goja-text/pkg/xgoja/providers/text` using the same wrapper pattern as core provider.

### Resource: `go-go-goja/pkg/xgoja/providers/core/core.go`

- **What I was researching:** How to wrap existing `modules.NativeModule` values as xgoja provider modules.
- **What I was looking for:** The exact `nativeModuleEntry()` pattern.
- **Why I chose it / what led me to it:** Markdown will be a `NativeModule`, so this is the closest xgoja provider pattern.
- **Useful findings:** Core provider blank-imports built-in modules, fetches them with `modules.GetModule`, and wraps `mod.Loader` in a `providerapi.Module`.
- **Not useful:** Core provider is data-safe; it does not show host capability config.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update. New text provider should copy this pattern.

### Resource: `go-go-goja/pkg/xgoja/providers/host/host.go`

- **What I was researching:** Host capability modules, especially filesystem access.
- **What I was looking for:** How `fs` is guarded and configured.
- **Why I chose it / what led me to it:** User wanted `fs` support to load Markdown files from disk.
- **Useful findings:** Host `fs` requires either `config.allow=true` for OS filesystem or `config.embedded.allow=true` for embedded assets. It disallows combining both in one module instance.
- **Not useful:** Does not handle markdown directly.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update. `xgoja.yaml` should use explicit `config.allow: true` for development/testing.

### Resource: `go-go-goja/pkg/xgoja/testprovider/provider.go`

- **What I was researching:** A complete provider package example.
- **What I was looking for:** Modules, capabilities, command providers, embedded verb sources.
- **Why I chose it / what led me to it:** Needed a richer example than core/host for future extensions.
- **Useful findings:** Shows provider packages can bundle modules, capabilities, command providers, and verb sources. Useful for future pre-shipped jsverbs or markdown option sections.
- **Not useful:** Test fixture semantics are not production module semantics.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update. Future text provider may use this as a model if adding capabilities or verb sources.

### Resource: `go-go-goja/examples/xgoja/*/xgoja.yaml`

- **What I was researching:** Concrete xgoja specs.
- **What I was looking for:** Core provider, host provider, filesystem runtime, jsverbs examples.
- **Why I chose it / what led me to it:** Examples are the fastest way to avoid inventing spec shape from scratch.
- **Useful findings:** `01-core-provider` shows core modules; `02-host-provider` shows `fs` with `allow: true`; `06-runtime-filesystem` shows `jsverbs` and runtime configuration.
- **Not useful:** Examples do not include a local provider package requiring `replace: .`.
- **Out of date / wrong:** None observed.
- **Needs updating:** No example update required; our own spec needs local provider replace.

---

## 6. jsverbs and REPL Surfaces

### Resource: `go-go-goja/pkg/jsverbs/model.go`, `scan.go`, `binding.go`, `command.go`, and `runtime.go`

- **What I was researching:** How JavaScript functions become CLI commands.
- **What I was looking for:** Whether text queries can be written in JS and exposed as commands.
- **Why I chose it / what led me to it:** User said jsverbs can exercise the bindings.
- **Useful findings:** jsverbs scans JS with tree-sitter, builds Glazed command descriptions, and invokes functions in a goja runtime. It can reuse xgoja runtimes via `InvokeInRuntime`.
- **Not useful:** jsverbs is not required for the Phase 1 parser smoke test.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update. Design should treat jsverbs as Phase 3, not Phase 1.

### Resource: `go-go-goja/cmd/jsverbs-example/main.go`

- **What I was researching:** How jsverbs are exposed through a CLI.
- **What I was looking for:** A concrete command wiring example.
- **Why I chose it / what led me to it:** Needed a runnable reference for jsverbs outside xgoja.
- **Useful findings:** Shows scanning, shared sections, building commands, and adding them to Cobra via Glazed.
- **Not useful:** xgoja provides its own jsverbs attachment path, so this is background context.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update.

### Resource: `go-go-goja/cmd/goja-repl/root.go`, `cmd_eval.go`, and `cmd_bindings.go`

- **What I was researching:** Alternative REPL/testing workflow.
- **What I was looking for:** How existing goja-repl enables modules and evaluates JS.
- **Why I chose it / what led me to it:** Initial plan mentioned goja-repl; later xgoja became primary.
- **Useful findings:** Good background on runtime app construction, session APIs, module flags, and evaluation.
- **Not useful:** Less suitable than xgoja for a separate `goja-text` module because goja-repl would need module imports available in its binary.
- **Out of date / wrong:** The primary design document temporarily duplicated goja-repl material after the xgoja update; that should be cleaned up.
- **Needs updating:** Original design doc still has some duplicated goja-repl content and should be tidied in a doc cleanup pass.

### Resource: `go-go-goja/pkg/replapi/app.go`

- **What I was researching:** High-level REPL application facade.
- **What I was looking for:** How sessions evaluate source and expose runtime access.
- **Why I chose it / what led me to it:** Relevant to goja-repl and TUI understanding.
- **Useful findings:** Shows app/session boundaries and how runtime can be accessed via `WithRuntime`.
- **Not useful:** xgoja generated binary does not primarily use `replapi.App` for `run`/`eval`.
- **Out of date / wrong:** None observed.
- **Needs updating:** No source update.

---

## 7. Goldmark and Markdown Parsing

### Resource: `go-go-goja/go.mod` and `go-go-goja/go.sum`

- **What I was researching:** Whether goldmark was already present in the workspace.
- **What I was looking for:** Existing `github.com/yuin/goldmark` dependency version.
- **Why I chose it / what led me to it:** Wanted to avoid adding unnecessary or conflicting dependencies.
- **Useful findings:** `github.com/yuin/goldmark v1.8.2` is already present indirectly.
- **Not useful:** Indirect presence does not mean `goja-text` can rely on it without adding its own requirement.
- **Out of date / wrong:** None observed.
- **Needs updating:** `goja-text/go.mod` should add `github.com/yuin/goldmark v1.8.2` directly.

### Resource: goldmark API knowledge from module and planned imports (`github.com/yuin/goldmark`, `ast`, `text`)

- **What I was researching:** How to parse Markdown into an AST and convert it.
- **What I was looking for:** Parser flow: `goldmark.New()`, `Parser().Parse(text.NewReader(source))`, `ast.Walk`, node kinds.
- **Why I chose it / what led me to it:** Markdown parser implementation needs goldmark AST conversion.
- **Useful findings:** The planned conversion path is sound: source bytes → goldmark AST → internal `MarkdownNode` graph.
- **Not useful:** We have not yet verified all node-kind details in code; some fields such as image alt text need careful goldmark-specific handling.
- **Out of date / wrong:** The design includes pseudocode and should be verified against the exact goldmark v1.8.2 API before coding.
- **Needs updating:** Add a small prototype or tests in Phase 1 to verify node kinds, link/image fields, fenced code metadata, and source positions.

---

## 8. TypeScript Declaration Generation

### Resource: `go-go-goja/pkg/tsgen/spec/types.go`

- **What I was researching:** How to describe TypeScript declarations for the markdown module.
- **What I was looking for:** `spec.Module`, `spec.Function`, `spec.Param`, `spec.TypeRef`, and `RawDTS` support.
- **Why I chose it / what led me to it:** The initial plan used `spec.Any()` too broadly; later design needs explicit `MarkdownNode` / `WalkContext` types.
- **Useful findings:** `RawDTS` can define richer interfaces not expressible through the simple `spec` helpers.
- **Not useful:** Does not provide runtime validation; typings are documentation/static support only.
- **Out of date / wrong:** None observed.
- **Needs updating:** Markdown `TypeScriptModule()` should define Go-style field names (`Type`, `Children`, `Level`) to match final Go-backed object API.

---

## 9. Ticket, Writing, and Review Workflow Skills

### Resource: `/home/manuel/.pi/agent/skills/ticket-research-docmgr-remarkable/SKILL.md`

- **What I was researching:** How to create a complete ticket, design doc, diary, and reMarkable upload.
- **What I was looking for:** Required workflow and deliverable checklist.
- **Why I chose it / what led me to it:** User explicitly requested docmgr ticket docs, diary, and reMarkable upload.
- **Useful findings:** Gave the overall research pipeline: create ticket, gather evidence, write design doc, maintain diary, validate, upload.
- **Not useful:** At the time of use, it did not force explicit decision records strongly enough.
- **Out of date / wrong:** Not wrong, but incomplete for preventing buried design decisions.
- **Needs updating:** Add explicit decision-record requirements for major architecture/API choices.

### Resource: `/home/manuel/.pi/agent/skills/ticket-research-docmgr-remarkable/references/writing-style.md`

- **What I was researching:** Writing standard for primary design docs and diaries.
- **What I was looking for:** Structure, evidence rules, clarity patterns.
- **Why I chose it / what led me to it:** The design guide needed to be intern-ready.
- **Useful findings:** Strong guidance on evidence, tradeoffs, pseudocode, and stable structure.
- **Not useful:** Did not explicitly require decision records.
- **Out of date / wrong:** Not wrong; missing a useful pattern.
- **Needs updating:** Add a `Decision Records` section with context/options/decision/rationale/consequences/status.

### Resource: `/home/manuel/.pi/agent/skills/docmgr/SKILL.md`

- **What I was researching:** docmgr commands and conventions.
- **What I was looking for:** How to create docs, relate files, update changelog/tasks, and validate.
- **Why I chose it / what led me to it:** Ticket docs are managed by docmgr.
- **Useful findings:** Correct `docmgr doc relate` syntax, absolute file-note guidance, diary expectations, doctor workflow.
- **Not useful:** Does not define design-doc content patterns such as decision records.
- **Out of date / wrong:** None observed.
- **Needs updating:** Optional: mention that design docs may include decision records, but the better update is in writing/research skills.

### Resource: `/home/manuel/.pi/agent/skills/diary/SKILL.md`

- **What I was researching:** How to keep the investigation diary.
- **What I was looking for:** Step format, prompt context, what worked/failed/tricky sections.
- **Why I chose it / what led me to it:** User asked to keep a diary.
- **Useful findings:** Provided strict diary schema and continuity guidance.
- **Not useful:** Does not require linking to decision records when a step resolves a design choice.
- **Out of date / wrong:** None observed.
- **Needs updating:** Add a reminder to include or link decision records when a diary step resolves major architecture/API choices.

### Resource: `/home/manuel/.pi/agent/skills/full-blown-tech-research-design/SKILL.md`

- **What I was researching:** Which skills should mention decision records.
- **What I was looking for:** Whether broad research/design deliverables already require decision records.
- **Why I chose it / what led me to it:** User asked which skills could be updated so future interns do not miss decision records.
- **Useful findings:** Skill already requires intern-ready design docs but does not explicitly require decision records.
- **Not useful:** Not directly used for the original ticket workflow because the pinned ticket-research skill covered it.
- **Out of date / wrong:** Not wrong; incomplete for decision capture.
- **Needs updating:** Add decision records to output standard and design-doc checklist.

### Resource: `/home/manuel/.pi/agent/skills/go-go-goja-module-authoring/SKILL.md`

- **What I was researching:** Native-module authoring guidance and whether it conflicts with the Go-backed AST decision.
- **What I was looking for:** JS API conventions and module implementation guardrails.
- **Why I chose it / what led me to it:** This ticket is specifically about a go-go-goja native module.
- **Useful findings:** Strong guidance on separating domain logic from goja glue, runtime integration tests, and documentation.
- **Not useful:** It says to use lowerCamelCase option/result keys, which is not always correct for Go-backed domain objects.
- **Out of date / wrong:** The lowerCamel guidance is too broad for this ticket's final decision.
- **Needs updating:** Add an exception/decision-record requirement for intentionally exposing Go-backed domain objects with exported Go field names.

### Resource: `/home/manuel/.pi/agent/skills/code-quality-review-cleanup/SKILL.md` and `/home/manuel/.pi/agent/skills/frontend-review-docmgr-remarkable/SKILL.md`

- **What I was researching:** Whether review skills should also require decision records.
- **What I was looking for:** Existing review report structure and tradeoff guidance.
- **Why I chose it / what led me to it:** User asked broadly which skill documents could be updated.
- **Useful findings:** Both skills ask for evidence, tradeoffs, proposals, and file-backed recommendations.
- **Not useful:** They do not primarily govern this ticket's design workflow.
- **Out of date / wrong:** Not wrong.
- **Needs updating:** Add lightweight decision-record reminders when recommending major architectural cleanup or subsystem redesign.

---

## 10. Ticket Documents Produced During This Work

### Resource: `design-doc/01-goja-text-bindings-architecture-design-and-implementation-guide.md`

- **What I was researching:** This is the primary synthesized design document, not an external source.
- **What I was looking for:** A coherent implementation guide for the intern.
- **Why I chose it / what led me to it:** Created as the main deliverable.
- **Useful findings:** Captures architecture, xgoja, module provider, Go-backed AST, `walk`, implementation phases, testing, and risks.
- **Not useful:** It accumulated some outdated material across revisions and still deserves a cleanup pass for duplicated goja-repl content.
- **Out of date / wrong:** Earlier revisions assumed lowercase/map-like AST properties and one-off extractor exports. Current v4 corrects the final design direction.
- **Needs updating:** Clean duplicate goja-repl material; add explicit decision records inline; ensure all examples compile once implementation begins.

### Resource: `design-doc/02-review-of-the-goja-text-bindings-plan-and-spec.md`

- **What I was researching:** Review/mentorship artifact for the intern.
- **What I was looking for:** Strengths, weaknesses, missing evidence, and next-time planning advice.
- **Why I chose it / what led me to it:** User asked for a second document reviewing the intern's plan/spec.
- **Useful findings:** Good critique of xgoja replace, field-name assumptions, `walk`, TypeScript declarations, and validation semantics.
- **Not useful:** Its original map-based AST recommendation is superseded by the project decision to keep Go-backed objects.
- **Out of date / wrong:** Map-based AST as final recommendation is no longer correct.
- **Needs updating:** Superseding note added. If this becomes a canonical review artifact, rewrite relevant sections rather than relying on the note.

### Resource: `reference/01-investigation-diary.md`

- **What I was researching:** Chronological record of the investigation process.
- **What I was looking for:** What happened, why, what changed, what needs follow-up.
- **Why I chose it / what led me to it:** User asked to keep a diary.
- **Useful findings:** Tracks the evolution from initial design → xgoja update → review document → Go-backed AST decision.
- **Not useful:** Not a compact implementation guide; it is a chronological log.
- **Out of date / wrong:** Earlier steps record decisions that were later superseded; this is expected diary behavior.
- **Needs updating:** Keep adding entries during implementation, especially failed probes and final decisions.

---

## Quick Reference: Most Useful Sources for Implementation

1. `go-go-goja/modules/yaml/yaml.go` — simple module shape.
2. `go-go-goja/modules/uidsl/module.go` + `node.go` — Go-backed domain object precedent.
3. `go-go-goja/pkg/xgoja/providers/core/core.go` — wrapping `NativeModule` as xgoja provider module.
4. `go-go-goja/pkg/xgoja/providers/host/host.go` — enabling guarded `fs` access.
5. `go-go-goja/cmd/xgoja/internal/generate/gomod.go` — package `replace` behavior.
6. `go-go-goja/pkg/xgoja/app/run.go` — how generated binary runs JS files and sets module roots.
7. goja `FieldNameMapper` docs/tests in module cache — confirms field-name behavior and mapper options.
8. `go-go-goja/pkg/tsgen/spec/types.go` — TypeScript declarations for Go-backed API.

## Quick Reference: Known Update Needs

### Resolved in code / ticket docs

- ✅ `goja-text/go.mod` module path was updated from placeholder `github.com/go-go-golems/XXX` to `github.com/go-go-golems/goja-text`, with direct `goldmark` dependency and local `go-go-goja` replace.
- ✅ `replace: .` was added to the local `goja-text` provider package in `xgoja.yaml`.
- ✅ Runtime tests now prove JS can access `node.Type`, `node.Children`, and `node.Level` on returned `*MarkdownNode` values, and that lowercase `ast.type` is not the primary API.
- ✅ Parser and JS-runtime tests now cover goldmark edge fields for image destination/title/alt text, fenced code language/info/text, indented code block text/source positions, HTML block raw text, and inline raw HTML.
- ✅ TypeScript declarations now use Go-style field names via `RawDTS` in the markdown module.
- ✅ xgoja build and demo script have been validated with an absolute `--xgoja-replace` path.
- ✅ The primary design doc now has explicit decision records for xgoja, Go-backed AST objects, `walk()`, and validation semantics.
- ✅ The duplicated `goja-repl` section was cleaned from the primary design doc.

### Still worth addressing

- 🟠 **Decide whether to enable goldmark extensions and then test their node shapes.** Core goldmark edge fields now have executable coverage, but extension nodes remain intentionally unverified because `Parse` currently uses `goldmark.New()` without extensions.
- 🟡 **Clean or rewrite the review document sections that were superseded by the Go-backed AST decision** if the review document becomes a canonical teaching artifact rather than a historical critique.
- 🟡 **Propagate decision-record guidance to any remaining review/research skills opportunistically.** The most relevant docs now have explicit guidance (`ticket-research-docmgr-remarkable` main skill, its writing-style reference, and `go-go-goja-module-authoring`), but broader review skills can still be improved later.
