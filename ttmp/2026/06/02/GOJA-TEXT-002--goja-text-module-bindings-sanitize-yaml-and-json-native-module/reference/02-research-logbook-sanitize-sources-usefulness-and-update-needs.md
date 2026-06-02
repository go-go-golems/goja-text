---
Title: 'Research Logbook: Sanitize Sources, Usefulness, and Update Needs'
Ticket: GOJA-TEXT-002
Status: active
Topics:
    - goja
    - goja-bindings
    - sanitize
    - yaml
    - json
    - native-modules
    - text-algorithms
    - tree-sitter
DocType: reference
Intent: Track which resources shaped the sanitize goja binding work and what needs updating
Owners: []
RelatedFiles:
    - Path: ../../../../../../../go-go-goja/modules/exports.go
      Note: SetExport behavior that shaped namespace implementation
    - Path: ../../../../../../../sanitize/pkg/json/parse.go
      Note: Upstream JSON strict parse reference
    - Path: ../../../../../../../sanitize/pkg/yaml/options.go
      Note: Upstream YAML option semantics reference
    - Path: pkg/sanitize/module.go
      Note: Sanitize NativeModule exports
    - Path: pkg/sanitize/module_test.go
      Note: JavaScript runtime behavior evidence
    - Path: pkg/sanitize/options.go
      Note: Builder/config validation source
    - Path: xgoja.yaml
      Note: xgoja runtime composition
ExternalSources: []
Summary: Source-by-source research log for GOJA-TEXT-002 sanitize native module bindings.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Research Logbook: Sanitize Sources, Usefulness, and Update Needs

## Goal

This logbook records the source files, ticket documents, local documentation, and module-version evidence that shaped GOJA-TEXT-002. It is meant to help a future maintainer answer four questions quickly:

1. Which resources were useful for the sanitize goja binding design and implementation?
2. Which resources are authoritative versus only contextual?
3. Which assumptions were corrected during implementation?
4. What should be updated before future sanitize or structured-text helper work continues?

## Context

GOJA-TEXT-002 added a `require("sanitize")` native module to `goja-text`. The module exposes YAML and JSON linting, sanitizing, parse-tree inspection, rule catalogs, examples, and strict JSON validation from `github.com/go-go-golems/sanitize v0.0.2`. The API uses Go-backed builder/config objects so Go owns unknown-option policy and validation.

This logbook records both pre-implementation planning sources and implementation evidence. It includes local files read directly, generated ticket documents, validation commands, and the published module-version check for sanitize.

## Legend

- **Useful**: keep using this as implementation evidence.
- **Partly useful**: useful for context but not sufficient by itself.
- **Needs update**: our derived docs or code should be revised if this changes.
- **Out of date / wrong**: a source or earlier conclusion conflicted with final implementation.

---

## Quick Reference: Current Source Authority

| Area | Most authoritative source today | Why |
| --- | --- | --- |
| Public JS API | `goja-text/pkg/sanitize/module.go`, `module_test.go`, README | Code and runtime tests define behavior. |
| Builder/config validation | `goja-text/pkg/sanitize/options.go`, `options_test.go` | Implements unknown-option policy and rule validation. |
| YAML sanitize semantics | `github.com/go-go-golems/sanitize/pkg/yaml` at v0.0.2 | Upstream dependency behavior. |
| JSON sanitize semantics | `github.com/go-go-golems/sanitize/pkg/json` at v0.0.2 | Upstream dependency behavior. |
| xgoja composition | `goja-text/xgoja.yaml`, `pkg/xgoja/providers/text/text.go` | Defines generated binary modules. |
| Design rationale | `design-doc/01-sanitize-native-module-design-and-implementation-guide.md` | Updated with builder/config decision. |
| Mentoring critique | `design-doc/02-review-of-the-sanitize-module-plan-and-spec.md` | Explains earlier plan risks and corrections. |
| Work history | `reference/01-investigation-diary.md` | Chronological commands, failures, fixes. |

---

## 1. Workspace and Ticket Context

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/go.work`

- **What I was researching:** Whether the workspace contains `sanitize` alongside `goja-text` and `go-go-goja`.
- **What I was looking for in this document:** Local module topology and whether the sanitize checkout is part of the workspace.
- **Why I chose it:** Dependency-resolution questions came up when deciding whether to use a local `replace` for sanitize.
- **How I found the resource:** The workspace root had been used throughout GOJA-TEXT-001, and `rg` confirmed `../go.work` mentions `./sanitize`.
- **What I found useful:** Confirms local source discovery: `./sanitize` is present in the workspace.
- **What I didn't find useful:** It does not prove `goja-text` should depend on the local checkout. It only describes workspace membership.
- **What is out of date / wrong:** None.
- **What would need updating:** If the workspace stops including `./sanitize`, docs that refer to the local checkout as reference material should be updated.

### Resource: `goja-text/go.mod`

- **What I was researching:** The dependency boundary for the new sanitize module.
- **What I was looking for in this document:** Whether `github.com/go-go-golems/sanitize` was already required, whether local replaces existed, and whether `go-go-goja` was still locally replaced.
- **Why I chose it:** The implementation imports `github.com/go-go-golems/sanitize/pkg/yaml` and `/pkg/json`; `go.mod` determines whether standalone builds succeed.
- **How I found the resource:** It is the module manifest for the implementation repo.
- **What I found useful:** Initially, sanitize was absent. After implementation, `go.mod` requires `github.com/go-go-golems/sanitize v0.0.2` without a local replace.
- **What I didn't find useful:** It does not explain why a dependency was chosen; that context lives in the design doc and diary.
- **What is out of date / wrong:** Earlier design text said to use a local sanitize replace. That was corrected after verifying the published `v0.0.2` module resolves.
- **What would need updating:** If a future implementation requires sanitize APIs newer than `v0.0.2`, update the pin and record why.

### Resource: `goja-text/ttmp/.../GOJA-TEXT-002/design-doc/01-sanitize-native-module-design-and-implementation-guide.md`

- **What I was researching:** The planned API, architecture decisions, implementation phases, and testing strategy.
- **What I was looking for in this document:** Whether the plan matched implementation constraints discovered in code.
- **Why I chose it:** It is the primary design document for the ticket.
- **How I found the resource:** Created as the first GOJA-TEXT-002 deliverable.
- **What I found useful:** The document captures the final direction: one `sanitize` module, `yaml`/`json` namespaces, Go-backed builders/configs, pinned sanitize dependency, and xgoja integration.
- **What I didn't find useful:** The early version used raw options objects and local replace language, both later corrected.
- **What is out of date / wrong:** Any old references to raw options objects as the primary API or sanitize local replace were corrected in the current version.
- **What would need updating:** If Makefile targets are added or the API changes to lowerCamel aliases, update the implementation checklist and API examples.

### Resource: `goja-text/ttmp/.../GOJA-TEXT-002/design-doc/02-review-of-the-sanitize-module-plan-and-spec.md`

- **What I was researching:** Weaknesses in the initial sanitize design before implementation.
- **What I was looking for in this document:** Concrete corrections to fold into the primary plan.
- **Why I chose it:** It was written as a technical review of the initial plan.
- **How I found the resource:** Created as the second GOJA-TEXT-002 design document after the user requested an intern-facing review.
- **What I found useful:** It identified real risks: dotted `SetExport` names, dependency wiring, options decoding, unknown-option behavior, namespace-aware TypeScript declarations, and strict JSON parse scope.
- **What I didn't find useful:** It still assumed local sanitize replace was likely needed before the user clarified that the local checkout is only reference material.
- **What is out of date / wrong:** The local-replace recommendation is superseded by the final dependency decision: use published `v0.0.2` without local replace.
- **What would need updating:** Add a superseding note if this review becomes an onboarding artifact rather than a historical critique.

### Resource: `goja-text/ttmp/.../GOJA-TEXT-002/reference/01-investigation-diary.md`

- **What I was researching:** Chronological implementation context, commands, failures, and fixes.
- **What I was looking for in this document:** What had already been tried, which assumptions changed, and which validation commands passed.
- **Why I chose it:** The diary is the continuation source for non-trivial ticket work.
- **How I found the resource:** Created with the ticket and updated after each implementation step.
- **What I found useful:** Captures the local-replace correction, the builder/config switch, the `GOWORK=off` failure, xgoja validation, and reMarkable uploads.
- **What I didn't find useful:** It is chronological, not a compact API reference.
- **What is out of date / wrong:** Earlier entries mention local replace as reasonable; later entries correct that decision.
- **What would need updating:** Continue updating it if Makefile targets are added or the ticket is closed.

---

## 2. Sanitize Library Sources

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/README.md`

- **What I was researching:** What sanitize does, its CLI behavior, and its public library usage.
- **What I was looking for in this document:** Supported formats, core operations, examples of `Sanitize`, `Lint`, `ParseTree`, and rule names.
- **Why I chose it:** README is the fastest orientation source for a library before reading package internals.
- **How I found the resource:** `find sanitize -type f` showed the repo root and README.
- **What I found useful:** Confirmed sanitize supports YAML lint/repair, JSON lint/recovery, tree-sitter parse-tree inspection, rule catalogs, examples, and library usage with `yamlsanitize` and `jsonsanitize` packages.
- **What I didn't find useful:** It does not describe goja binding concerns, builder/config patterns, or xgoja integration.
- **What is out of date / wrong:** Nothing observed for v0.0.2-level usage, but the local checkout may include commits after v0.0.2.
- **What would need updating:** If the goja binding depends on APIs added after `v0.0.2`, this README should be checked against the pinned module version before updating goja-text.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/go.mod`

- **What I was researching:** sanitize's module path, Go version, and dependency profile.
- **What I was looking for in this document:** Confirm module path `github.com/go-go-golems/sanitize` and tree-sitter grammar dependencies.
- **Why I chose it:** The goja-text dependency must import the correct module path.
- **How I found the resource:** Standard Go library manifest in the local sanitize checkout.
- **What I found useful:** Confirms module path and dependencies on `tree-sitter-yaml`, `go-tree-sitter`, and `tree-sitter-json`.
- **What I didn't find useful:** It does not indicate which version goja-text should pin.
- **What is out of date / wrong:** The local checkout is newer than the published `v0.0.2` tag (`v0.0.2-5-gc142cca`), so it should be treated as reference material, not dependency truth.
- **What would need updating:** If goja-text upgrades sanitize, compare the pinned module's go.mod with the local checkout.

### Resource: `go list -m -json github.com/go-go-golems/sanitize@v0.0.2`

- **What I was researching:** Whether sanitize has a published pinned version that resolves without a local replace.
- **What I was looking for in this output:** Module path, version, tag time, module cache directory, and VCS hash.
- **Why I chose it:** The user questioned the local replace requirement and noted the local checkout is only for reference.
- **How I found the resource:** Ran `go list -m -json` from `goja-text` after the dependency-design correction.
- **What I found useful:** Confirmed `github.com/go-go-golems/sanitize v0.0.2` resolves and points to tag hash `f1f965c450178d25978bc6ba317ed769f3fcc5b3`.
- **What I didn't find useful:** It does not show API details; it only proves dependency availability.
- **What is out of date / wrong:** Supersedes earlier plan text recommending `replace ../sanitize`.
- **What would need updating:** If the tag is retracted, deleted, or a newer version is required, update `go.mod`, docs, and diary.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/yaml/types.go`

- **What I was researching:** YAML result object shapes exposed by the sanitize library.
- **What I was looking for in this document:** `Result`, `LintIssue`, `ErrorNode`, `Fix`, `Example`, and field names.
- **Why I chose it:** The goja binding exposes these structs directly as Go-backed objects.
- **How I found the resource:** `find sanitize -type f` and package layout inspection.
- **What I found useful:** Defined the public YAML result fields that JavaScript sees as PascalCase properties: `Sanitized`, `ParseClean`, `LintIssues`, `Fixes`, etc.
- **What I didn't find useful:** It does not define configuration; that lives in `options.go`.
- **What is out of date / wrong:** Local checkout may include post-v0.0.2 changes; implementation should be validated against the pinned module through tests.
- **What would need updating:** If sanitize changes result fields, update TypeScript declarations and runtime tests.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/json/types.go`

- **What I was researching:** JSON result object shapes and how they differ from YAML.
- **What I was looking for in this document:** JSON-specific fields such as `StrictParseClean` and `OriginalStrictParseClean`.
- **Why I chose it:** The JavaScript API needs to expose JSON strict-parse status clearly.
- **How I found the resource:** Package layout inspection after reading YAML types.
- **What I found useful:** Confirmed JSON result extends the YAML-like shape with strict parse fields.
- **What I didn't find useful:** It does not explain strict parse implementation; that is in `parse.go`.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If strict parse fields change, update `JsonResult` RawDTS and tests.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/yaml/sanitize.go`

- **What I was researching:** YAML sanitize algorithm and return semantics.
- **What I was looking for in this document:** Iteration loop, original/final state capture, stop conditions, and `SanitizeWithOptions` behavior.
- **Why I chose it:** The binding calls `SanitizeWithOptions` directly.
- **How I found the resource:** It is the natural implementation file after `types.go`.
- **What I found useful:** Confirmed the conservative loop: analyze, lint, apply fixes, stop when clean/no progress/max iterations.
- **What I didn't find useful:** It does not expose lower-level rule behavior; that is split across lint/fix/analysis files.
- **What is out of date / wrong:** None for the current binding.
- **What would need updating:** If sanitize adds context-aware or streaming APIs, revisit whether the goja binding should expose them.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/json/sanitize.go`

- **What I was researching:** JSON sanitize algorithm and strict parse integration.
- **What I was looking for in this document:** Stop conditions including `StrictParseError == nil` and final result fields.
- **Why I chose it:** JSON repair has stricter success criteria than YAML because of `encoding/json` validation.
- **How I found the resource:** Parallel to YAML sanitize implementation.
- **What I found useful:** Confirmed `StrictParseClean` is part of the success criteria and final result.
- **What I didn't find useful:** It does not explain individual JSON heuristics; those are in fix/heuristics/lint files.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If JSON repair becomes more aggressive, update docs to clarify conservative versus best-effort semantics.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/yaml/options.go`

- **What I was researching:** YAML configuration and validation behavior.
- **What I was looking for in this document:** `WithMaxIterations`, `WithTabWidth`, `WithOnlyRules`, `WithDisabledRules`, and rule-overlap validation.
- **Why I chose it:** The builder/config layer converts JavaScript calls into these Go functional options.
- **How I found the resource:** Followed README library usage and sanitize package file list.
- **What I found useful:** Confirmed rule-name validation and overlap rejection are already part of the library config build process.
- **What I didn't find useful:** The internal `config` type is private, so goja-text cannot reuse it directly.
- **What is out of date / wrong:** The original plan treated raw JS options as the primary API; this source helped motivate a builder layer that centralizes validation.
- **What would need updating:** If sanitize exposes public config structs later, goja-text could simplify its builder implementation.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/json/options.go`

- **What I was researching:** JSON configuration and differences from YAML.
- **What I was looking for in this document:** Whether JSON has `tabWidth` and how rule validation works.
- **Why I chose it:** The JSON builder should mirror YAML but omit YAML-only options.
- **How I found the resource:** Parallel to YAML options after package inspection.
- **What I found useful:** Confirmed JSON supports max iterations, only rules, and disabled rules, but no tab width.
- **What I didn't find useful:** The private config type cannot be returned directly to JavaScript.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If JSON adds format-specific options, update `JsonOptionsBuilder`, `JsonConfig`, RawDTS, and tests.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/yaml/rules.go`

- **What I was researching:** YAML rule catalog and validation helpers.
- **What I was looking for in this document:** Rule names, `RuleSpec` fields, `RuleCatalog`, `ValidateRuleNames`.
- **Why I chose it:** JavaScript exposes `sanitize.yaml.rules()` and builder validation depends on known rule names.
- **How I found the resource:** Read after options because `options.go` calls rule validation.
- **What I found useful:** Defined the YAML rule names used in tests and examples: `tab_indent`, `missing_space_after_colon`, `list_dash_no_space`, etc.
- **What I didn't find useful:** It does not say which rules are likely to fix a particular input; tests discover that through actual sanitize calls.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If rule names change, update fixtures, demos, and builder validation tests.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/json/rules.go`

- **What I was researching:** JSON rule catalog and parse-aware metadata.
- **What I was looking for in this document:** Rule names, `ParseAware`, and catalog size.
- **Why I chose it:** JavaScript exposes `sanitize.json.rules()` and TypeScript declares JSON `RuleSpec` with `ParseAware`.
- **How I found the resource:** Parallel to YAML rule catalog.
- **What I found useful:** Confirmed JSON has 15 rules in the current validation run, including `markdown_fence_wrapper`, `single_quotes`, `python_literals`, `trailing_comma`, `strict_parse_error`, and parse/missing-node rules.
- **What I didn't find useful:** It does not encode repair order; that is in fix logic.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If JSON rules are added/removed, update smoke expectations that mention rule counts only if they become strict assertions.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/yaml/parse.go`

- **What I was researching:** YAML tree-sitter parse-tree API.
- **What I was looking for in this document:** `ParseTree(src)` signature and `ErrorNode` collection behavior.
- **Why I chose it:** The goja module exposes `sanitize.yaml.parseTree(input)`.
- **How I found the resource:** Package layout inspection.
- **What I found useful:** Confirmed `ParseTree` returns rendered tree text and parse errors, not a structured AST.
- **What I didn't find useful:** It does not support traversable tree nodes; therefore structured tree exposure is out of scope for the current binding.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If sanitize exposes structured parse trees later, add a separate design decision before exposing them to JavaScript.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/json/parse.go`

- **What I was researching:** JSON parse-tree API and strict parser API.
- **What I was looking for in this document:** `ParseTree`, `StrictParse`, and multiple-top-level-value behavior.
- **Why I chose it:** The review identified `json.strictParse(input)` as an API choice, and implementation exposed it.
- **How I found the resource:** Package layout inspection and review follow-up.
- **What I found useful:** Confirmed `StrictParse(src) error` exists and is simple to expose as `{ Valid, Error }`.
- **What I didn't find useful:** It does not explain how repair heuristics work.
- **What is out of date / wrong:** The initial plan left strict parse as an open question; implementation resolved it by exposing `sanitize.json.strictParse`.
- **What would need updating:** If strict parse starts returning richer diagnostics, update `StrictParseResult`.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/yaml/lint.go`

- **What I was researching:** YAML lint issue generation and rule-source semantics.
- **What I was looking for in this document:** How lint issues are assembled from parse errors, line heuristics, duplicate keys, and mixed indentation.
- **Why I chose it:** JavaScript exposes `sanitize.yaml.lint()` and result `LintIssues`.
- **How I found the resource:** Followed `SanitizeWithOptions` into linting functions.
- **What I found useful:** Confirmed lint issues include rule, source, description, byte/row/column spans, and display row.
- **What I didn't find useful:** It was not necessary to read every fixer detail to implement bindings.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If lint issue shape changes, update RawDTS and field projection tests.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/json/lint.go`

- **What I was researching:** JSON lint issue generation and strict parse error surfacing.
- **What I was looking for in this document:** How parse errors, strict parser errors, heuristics, and duplicate keys become `LintIssue` values.
- **Why I chose it:** JavaScript exposes `sanitize.json.lint()` and strict parse status.
- **How I found the resource:** Followed JSON sanitize and parse behavior.
- **What I found useful:** Confirmed JSON lint sources include `parse`, `strict-parser`, and heuristic-style sources.
- **What I didn't find useful:** The goja binding does not need to know each internal heuristic to expose the result.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If JSON lint source names become public API, document them explicitly in README or TypeScript comments.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/yaml/examples.go`

- **What I was researching:** Built-in YAML examples to expose through JavaScript.
- **What I was looking for in this document:** Shape and contents of `yamlsanitize.Examples`.
- **Why I chose it:** The design included `sanitize.yaml.examples()`.
- **How I found the resource:** Package file list and README mention examples.
- **What I found useful:** Confirms examples are simple Go structs and can be returned directly.
- **What I didn't find useful:** Examples are not exhaustive; they are demonstration inputs, not a regression corpus.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If examples become large or file-backed, consider lazy loading or metadata-only APIs.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/sanitize/pkg/json/examples.go`

- **What I was researching:** Built-in JSON examples to expose through JavaScript.
- **What I was looking for in this document:** Shape and contents of `jsonsanitize.Examples`.
- **Why I chose it:** The design included `sanitize.json.examples()`.
- **How I found the resource:** Parallel to YAML examples.
- **What I found useful:** Confirms examples include common malformed LLM JSON cases such as single quotes, Python literals, Markdown fences, and leading prose.
- **What I didn't find useful:** It does not guarantee all fix combinations; implementation uses separate demo fixtures.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If examples change names or fields, update RawDTS and demo expectations if needed.

---

## 3. go-go-goja and goja-text Binding Sources

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/go-go-goja/modules/common.go`

- **What I was researching:** Native module registration interface.
- **What I was looking for in this document:** `NativeModule`, registry behavior, `Register`, `GetModule`.
- **Why I chose it:** The sanitize module must register with `modules.Register(&module{})`.
- **How I found the resource:** It had already been central in GOJA-TEXT-001 and was listed in design references.
- **What I found useful:** Confirms the minimal module contract: `Name`, `Doc`, `Loader`.
- **What I didn't find useful:** It does not explain TypeScript declarations or xgoja provider wrapping.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If `NativeModule` changes, all goja-text modules need review.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/go-go-goja/modules/exports.go`

- **What I was researching:** How `modules.SetExport` actually sets properties.
- **What I was looking for in this document:** Whether dotted names such as `"yaml.sanitize"` create nested exports.
- **Why I chose it:** The review document flagged the initial pseudocode as likely wrong.
- **How I found the resource:** Followed the `modules.SetExport` helper used by existing modules.
- **What I found useful:** Confirmed `SetExport` just calls `Set(name, value)`; dotted strings are literal keys. This directly shaped the implementation: create `yamlObj`/`jsonObj`, set simple function names, then attach objects to top-level exports.
- **What I didn't find useful:** It only logs errors; it does not return them. For namespace object setting, the implementation uses direct `exports.Set` and panics on failure.
- **What is out of date / wrong:** Superseded initial design pseudocode that used `"yaml.sanitize"` as an export name.
- **What would need updating:** If `SetExport` later supports dotted paths, current explicit namespace code remains valid but docs can mention the helper behavior changed.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/go-go-goja/modules/yaml/yaml.go`

- **What I was researching:** Existing goja option decoding and unknown option behavior.
- **What I was looking for in this document:** How a native module handles optional config maps, numeric conversions, and unknown keys.
- **Why I chose it:** The sanitize review needed a concrete option-decoding reference.
- **How I found the resource:** Existing go-go-goja module list and review investigation.
- **What I found useful:** It rejects unknown options for `yaml.stringify` and handles numeric values as `int64`, `int`, or `float64`. This influenced sanitize builder validation.
- **What I didn't find useful:** The sanitize implementation ultimately uses Go-backed builders rather than direct map decoding in exported functions.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If go-go-goja adopts a shared options codec, sanitize builders could reuse it.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/markdown/module.go`

- **What I was researching:** Existing goja-text native module implementation style.
- **What I was looking for in this document:** Loader structure, TypeScript declaration pattern, Go-backed object projection, error wrapping.
- **Why I chose it:** Markdown is the first implemented goja-text module and the closest local pattern.
- **How I found the resource:** Existing implementation from GOJA-TEXT-001.
- **What I found useful:** Provided the module skeleton and RawDTS pattern.
- **What I didn't find useful:** It has no options builder and no nested namespaces, so it was not enough for sanitize by itself.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If markdown changes to lowerCamel adapters, revisit whether sanitize should follow.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/markdown/module_test.go`

- **What I was researching:** Runtime integration test shape for `require()` modules.
- **What I was looking for in this document:** How to create a go-go-goja runtime and execute JavaScript tests with `rt.Owner.Call`.
- **Why I chose it:** Sanitize needed runtime tests, not only Go unit tests.
- **How I found the resource:** Existing markdown module tests.
- **What I found useful:** Provided the pattern for `engine.NewBuilder().UseModuleMiddleware(engine.MiddlewareOnly(...))` and checking JavaScript-visible fields.
- **What I didn't find useful:** It did not cover builder methods, Go-backed slices, or nested namespaces.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If runtime setup changes, update sanitize and markdown tests together.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/xgoja/providers/text/text.go`

- **What I was researching:** How goja-text modules are registered with xgoja.
- **What I was looking for in this document:** Provider package ID, module list, blank imports, `nativeModuleEntry` wrapper.
- **Why I chose it:** Sanitize must be included in the generated `dist/goja-text` binary.
- **How I found the resource:** Existing xgoja provider file referenced in the GOJA-TEXT-001 design.
- **What I found useful:** The implementation only needed a blank import of `pkg/sanitize` and adding `"sanitize"` to `textModuleNames`.
- **What I didn't find useful:** It does not manage per-module configuration; sanitize configuration is runtime API-level, not provider-level.
- **What is out of date / wrong:** Earlier provider only had `markdown`; current implementation includes `sanitize`.
- **What would need updating:** If new text helper modules are added, they should be added to this provider list.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/xgoja.yaml`

- **What I was researching:** Generated binary composition.
- **What I was looking for in this document:** Whether `sanitize` is included in the `main` runtime and whether host `fs` is available for demo scripts.
- **Why I chose it:** xgoja is the primary exercise harness for goja-text.
- **How I found the resource:** Existing xgoja build spec.
- **What I found useful:** The implementation added `sanitize` next to `markdown` and reused existing `fs` access.
- **What I didn't find useful:** It cannot express sanitize builder behavior; that is runtime API code.
- **What is out of date / wrong:** None after implementation.
- **What would need updating:** If Makefile targets are added, keep their commands aligned with this spec.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/README.md`

- **What I was researching:** User-facing usage documentation.
- **What I was looking for in this document:** Existing markdown module docs and xgoja build/smoke command style.
- **Why I chose it:** Sanitize usage needed to be documented after implementation.
- **How I found the resource:** Repo root documentation.
- **What I found useful:** Provided a concise structure for module docs and smoke commands.
- **What I didn't find useful:** It did not previously mention sanitize; implementation updated it.
- **What is out of date / wrong:** Current README includes a hardcoded absolute `--xgoja-replace` path, inherited from earlier work.
- **What would need updating:** Add Makefile targets to avoid repeating the absolute path manually.

---

## 4. Implemented GOJA-TEXT-002 Files

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/sanitize/types.go`

- **What I was researching:** Final local wrapper types produced during implementation.
- **What I was looking for in this document:** Config objects, unknown-option policy, validation result, parse-tree wrappers, strict parse result.
- **Why I chose it:** It is now an authoritative source for goja-text-specific sanitize API types.
- **How I found the resource:** Created during implementation Step 5.
- **What I found useful:** It separates goja-text binding types from upstream sanitize result types.
- **What I didn't find useful:** It does not contain validation logic; that is in `options.go`.
- **What is out of date / wrong:** None.
- **What would need updating:** If lowerCamel adapters are added, decide whether to add separate plain-object types rather than changing these Go-backed structs.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/sanitize/options.go`

- **What I was researching:** Final builder/config implementation.
- **What I was looking for in this document:** Unknown-option policies, validation flow, rule-name checks, numeric/string-array import behavior, config-to-functional-options conversion.
- **Why I chose it:** This is the core implementation of the user's requested builder pattern.
- **How I found the resource:** Created during implementation Step 5.
- **What I found useful:** Centralizes validation and prevents exported functions from decoding raw objects independently.
- **What I didn't find useful:** It does not expose JavaScript docs; README and RawDTS cover that.
- **What is out of date / wrong:** None.
- **What would need updating:** If sanitize adds options, add builder methods, config fields, RawDTS entries, and tests.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/sanitize/module.go`

- **What I was researching:** Final JavaScript module exports.
- **What I was looking for in this document:** `sanitize.yaml` and `sanitize.json` namespace wiring and function behavior.
- **Why I chose it:** This is the authoritative implementation of `require("sanitize")`.
- **How I found the resource:** Created during implementation Step 5.
- **What I found useful:** Confirms nested namespace objects use simple property names and direct `exports.Set("yaml", yamlObj)`/`exports.Set("json", jsonObj)`.
- **What I didn't find useful:** It intentionally does not include TypeScript declaration details.
- **What is out of date / wrong:** None.
- **What would need updating:** If more structured-data helpers are added under another module, decide whether they belong in `sanitize` or a new module.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/sanitize/typescript.go`

- **What I was researching:** Final TypeScript declaration surface.
- **What I was looking for in this document:** Namespace-aware `RawDTS` entries for builders, configs, results, rules, examples.
- **Why I chose it:** The review warned that nested namespaces are not well represented by flat `spec.Function` declarations.
- **How I found the resource:** Created during implementation Step 5.
- **What I found useful:** RawDTS cleanly describes nested `yaml` and `json` exports.
- **What I didn't find useful:** TypeScript declarations are not runtime validation; tests still pin actual behavior.
- **What is out of date / wrong:** None observed.
- **What would need updating:** If runtime API changes, update RawDTS at the same time.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/sanitize/options_test.go`

- **What I was researching:** Whether builder/config validation works at the Go layer.
- **What I was looking for in this document:** Tests for normal config build, unknown-option rejection/collection, unknown rules, overlap.
- **Why I chose it:** Builder behavior should be tested independently from goja runtime projection.
- **How I found the resource:** Created during implementation Step 5.
- **What I found useful:** Confirms Go-side validation semantics before JavaScript tests.
- **What I didn't find useful:** It does not prove JavaScript method calls work; `module_test.go` does.
- **What is out of date / wrong:** None.
- **What would need updating:** Add tests when builder methods are added.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/sanitize/module_test.go`

- **What I was researching:** Whether the public JavaScript API works through goja.
- **What I was looking for in this document:** Runtime tests for `require("sanitize")`, builder methods, sanitize/lint/parse/rules/examples/strictParse, field projection.
- **Why I chose it:** Public behavior is defined by JavaScript runtime visibility, not only Go compile success.
- **How I found the resource:** Created during implementation Step 5 using markdown runtime tests as a model.
- **What I found useful:** Captures tricky behavior: Go-backed slice indexing from JavaScript and PascalCase result fields.
- **What I didn't find useful:** It does not test the generated xgoja binary; smoke scripts do.
- **What is out of date / wrong:** None.
- **What would need updating:** If JavaScript adapters add lowerCamel aliases, add tests without removing PascalCase tests unless the API is intentionally migrated.

### Resource: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/examples/js/sanitize-demo.js`

- **What I was researching:** End-to-end generated binary behavior.
- **What I was looking for in this document:** File-backed YAML/JSON sanitizing with host `fs` and builder configs.
- **Why I chose it:** xgoja smoke tests need a script that exercises real file access and module behavior.
- **How I found the resource:** Created during implementation Step 6.
- **What I found useful:** Validated YAML repair, JSON repair, rule counts, and strict parse in one generated binary run.
- **What I didn't find useful:** It is a smoke demo, not a comprehensive test suite.
- **What is out of date / wrong:** None.
- **What would need updating:** If fixture contents change, update expected human interpretation in docs.

---

## 5. Validation Commands and Results

### Resource: `go test ./... -count=1`

- **What I was researching:** Whether the implementation passes normal workspace tests.
- **What I was looking for in this command:** Compile success and test success for markdown, sanitize, and provider packages.
- **Why I chose it:** Standard Go validation checkpoint.
- **How I found the resource:** Existing GOJA-TEXT-001 workflow and diary practice.
- **What I found useful:** Passes after core sanitize implementation and after xgoja integration.
- **What I didn't find useful:** It did not catch the missing direct sanitize requirement while `go.work` was active.
- **What is out of date / wrong:** Treating this as sufficient alone was wrong.
- **What would need updating:** Make `GOWORK=off` part of standard validation.

### Resource: `GOWORK=off go test ./... -count=1`

- **What I was researching:** Whether goja-text can build as a standalone module outside workspace assistance.
- **What I was looking for in this command:** Missing direct dependencies and go.sum issues.
- **Why I chose it:** GOJA-TEXT-001 established this as an important generated-build/CI-style check.
- **How I found the resource:** Existing diary and goja module authoring skill guidance.
- **What I found useful:** Caught the missing `github.com/go-go-golems/sanitize` requirement after initial code compilation under workspace mode.
- **What I didn't find useful:** It does not build the xgoja binary.
- **What is out of date / wrong:** None; this should remain a required checkpoint.
- **What would need updating:** Add a Makefile target such as `test-standalone`.

### Resource: `go run ../go-go-goja/cmd/xgoja build -f xgoja.yaml --xgoja-replace /home/manuel/workspaces/2026-06-02/goja-text/go-go-goja`

- **What I was researching:** Whether the generated xgoja binary builds with sanitize included.
- **What I was looking for in this command:** xgoja spec validation, generated temporary module build success, provider registration success.
- **Why I chose it:** xgoja is the intended exercise harness.
- **How I found the resource:** Established during GOJA-TEXT-001 and repeated for sanitize integration.
- **What I found useful:** Build succeeded and produced `dist/goja-text`.
- **What I didn't find useful:** The command is verbose and contains a machine-specific absolute path.
- **What is out of date / wrong:** Manual command repetition is error-prone.
- **What would need updating:** Add Makefile targets that derive the absolute path.

### Resource: `./dist/goja-text run examples/js/sanitize-demo.js`

- **What I was researching:** End-to-end runtime behavior in the generated binary.
- **What I was looking for in this command:** `fs` access, `require("sanitize")`, builder/config use, YAML repair, JSON repair, strict parse result.
- **Why I chose it:** It validates the actual user-facing binary path.
- **How I found the resource:** Created the demo script as part of xgoja integration.
- **What I found useful:** Produced clean sanitized YAML and strict JSON from broken fixtures.
- **What I didn't find useful:** It is not a formal assertion-based test.
- **What is out of date / wrong:** None.
- **What would need updating:** Convert to an assertion-based smoke test if this becomes CI-critical.

---

## Quick Reference: Known Update Needs

### Resolved during GOJA-TEXT-002

- ✅ The sanitize dependency is pinned to `github.com/go-go-golems/sanitize v0.0.2` without a local replace.
- ✅ The primary design now uses Go-backed builder/config objects instead of raw options objects as the primary API.
- ✅ Unknown-option behavior is controllable through `RejectUnknownOptions`, `AllowUnknownOptions`, and `CollectUnknownOptions`.
- ✅ The initial dotted `SetExport` risk was avoided; implementation creates explicit nested namespace objects.
- ✅ Namespace-aware TypeScript declarations are implemented with `RawDTS`.
- ✅ `sanitize.json.strictParse(input)` is exposed.
- ✅ Normal tests, standalone tests, xgoja build, eval smoke, and demo smoke all passed.

### Still worth addressing

- ✅ **Makefile validation targets were added** for `test`, `test-standalone`, `build-xgoja`, `smoke-markdown`, `smoke-sanitize`, `smoke`, and `check`; `make check` passed.
- 🟡 **Update the review document with a superseding note** that local sanitize replace is no longer recommended.
- 🟡 **Decide whether `AllowUnknownOptions` should be documented as advanced/rare** to avoid callers using it casually.
- 🟡 **Consider assertion-based xgoja smoke tests** if generated binary behavior should be part of CI.
- 🟡 **Keep an eye on sanitize versions after `v0.0.2`**; the local checkout is newer than the pinned dependency, so do not assume local source behavior is identical to the pinned module without tests.

## Usage Examples

Use this logbook when:

- preparing to close GOJA-TEXT-002,
- adding Makefile validation targets,
- implementing future structured-data extraction helpers,
- upgrading the sanitize dependency,
- reviewing why the API uses Go-backed builders instead of raw options objects.
