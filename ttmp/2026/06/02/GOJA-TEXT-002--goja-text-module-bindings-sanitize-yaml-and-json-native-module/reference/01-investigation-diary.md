---
Title: "Investigation Diary"
Ticket: GOJA-TEXT-002
DocType: reference
Topics:
  - goja
  - goja-bindings
  - sanitize
  - yaml
  - json
---

# Investigation Diary

## Goal

Capture the investigation and implementation process for creating goja bindings for the sanitize library, exposing YAML and JSON sanitization, linting, parse-tree inspection, rule catalog enumeration, and example retrieval to JavaScript.

---

## Step 1: Close GOJA-TEXT-001 and Initialize GOJA-TEXT-002

Closed the markdown parser ticket and created a new ticket for the sanitize module bindings. The sanitize library is a mature Go project with tree-sitter-based YAML and JSON linting/fixing. The goal is to expose it through `require("sanitize")` in the same goja-text runtime that already hosts `require("markdown")`.

### Prompt Context

**User prompt (verbatim):** "ok, let's create a new docmgr ticket after closing this one: - adding support to sanitize yaml and json files, based on the ./sanitize library. Create a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet points and pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and then upload to remarkable."

**Assistant interpretation:** Close the current markdown ticket, create a new ticket for sanitize goja bindings, and produce a comprehensive intern-facing design document.

**Inferred user intent:** Continue the goja-text project by adding a second native module (sanitize) using the same patterns established for markdown.

### What I did

- Closed GOJA-TEXT-001 via `docmgr ticket close`
- Created GOJA-TEXT-002 via `docmgr ticket create-ticket`
- Read the entire sanitize library to understand its architecture:
  - `sanitize/pkg/yaml/types.go`, `sanitize.go`, `lint.go`, `parse.go`, `rules.go`, `options.go`, `analysis.go`, `fix.go`, `examples.go`
  - `sanitize/pkg/json/types.go`, `sanitize.go`, `lint.go`, `parse.go`, `rules.go`, `options.go`, `analysis.go`, `fix.go`, `examples.go`
  - `sanitize/README.md`, `go.mod`
- Read the existing goja-text infrastructure:
  - `pkg/markdown/module.go` — reference native module pattern
  - `pkg/xgoja/providers/text/text.go` — reference provider wrapping
  - `xgoja.yaml` — reference build spec
- Wrote the primary design document:
  - Executive summary
  - Problem statement
  - Sanitize library architecture overview
  - Goja native module pattern
  - xgoja build system
  - Decision records
  - Proposed module API with exact shapes
  - Phased implementation plan
  - File layout
  - Key source references
  - Testing strategy
  - Risks, alternatives, and open questions
  - Implementation checklist

### Why

The sanitize library is the next logical module for goja-text. It follows the same native module pattern as markdown but has different domain concerns: iterative fixing, rule catalogs, options configuration, and two parallel format packages. The design document must explain all of this to an intern who may not be familiar with tree-sitter, functional options, or the sanitize library's conservative fix philosophy.

### What worked

- The sanitize library has clean parallel packages for YAML and JSON. This makes the module design straightforward: mirror the Go package structure with `sanitize.yaml.*` and `sanitize.json.*` JavaScript namespaces.
- The existing markdown module in goja-text provides a complete reference for the native module pattern, xgoja provider wrapping, and test structure.
- The options decoding pattern (JS object → Go functional options) is a natural mapping.

### What didn't work

- N/A — this step produced documentation only.

### What I learned

- The sanitize library uses tree-sitter for both parsing and structural error detection. This is different from goldmark (used in markdown), which is a purpose-built Markdown parser.
- The sanitize library's iterative fix algorithm is conservative: it stops when no more fixes can be applied. This is an important behavioral detail to document for JavaScript callers.
- JSON has a `StrictParseClean` field that YAML does not, because JSON has a separate `encoding/json` validation layer. This asymmetry must be visible in the JavaScript API.

### What was tricky to build

- The main design challenge was deciding between a unified API (single `sanitize()` function with format dispatch) and a namespace-based API (`sanitize.yaml.sanitize()`, `sanitize.json.sanitize()`). The namespace approach was chosen because it keeps the API self-documenting and mirrors the Go package structure.
- Another challenge was mapping Go functional options to JavaScript. The solution is a camelCase options object where each key corresponds to one Go `Option` function.

### What warrants a second pair of eyes

- Whether the `parseTree` return shape (string + errors) is sufficient, or whether a structured tree would be more useful.
- Whether `StrictParse` should be exposed as a dedicated function for JSON.

### What should be done in the future

- Implement the module following the phased plan in the design document.
- Add edge-field regression tests for all result types.
- Validate the xgoja binary with demo scripts.

### Code review instructions

- Start with `design-doc/01-sanitize-native-module-design-and-implementation-guide.md`
- Verify the API shapes match the sanitize library's Go types.
- Check that the decision records capture the important architectural choices.

### Technical details

- Ticket path: `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/ttmp/2026/06/02/GOJA-TEXT-002--goja-text-module-bindings-sanitize-yaml-and-json-native-module/`
- Design doc: `design-doc/01-sanitize-native-module-design-and-implementation-guide.md`
- Key source references:
  - `sanitize/pkg/yaml/*.go`
  - `sanitize/pkg/json/*.go`
  - `goja-text/pkg/markdown/module.go`
  - `goja-text/pkg/xgoja/providers/text/text.go`

---

## Step 2: Review the Sanitize Module Plan

Created a second intern-facing review document for the GOJA-TEXT-002 sanitize module plan. The review praises the plan's correct architecture direction, then calls out concrete implementation risks that would likely slow or break the first coding pass: dependency wiring, dotted export names, options decoding, unknown option handling, TypeScript namespace declarations, and strict JSON validation scope.

### Prompt Context

**User prompt (verbatim):** "Our little brother just did a detailed plan a spec and analysis of the feature. I want you to create a second document where you review their work: what's good, what's bad, what could be better, what     
they should have known, what they should haev looked at, what they should keep in mind next time they do this kind of work.                                                                                 
                                                                                                                                                                                                            
Create  a detailed analysis / design / review  that is for our little brother (intern). It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Write a detailed technical review of the existing sanitize binding plan, aimed at mentoring an intern and making the plan more implementation-ready.

**Inferred user intent:** Preserve the intern plan as useful work, but add a second document that teaches how to critique and harden such a plan before implementation starts.

### What I did

- Re-read the primary sanitize design document.
- Checked the exact `modules.SetExport` implementation in `go-go-goja/modules/exports.go`.
- Checked the existing `go-go-goja/modules/yaml/yaml.go` option-decoding pattern.
- Checked the current `goja-text/go.mod` dependency state.
- Checked the existing xgoja provider wrapper at `pkg/xgoja/providers/text/text.go`.
- Wrote `design-doc/02-review-of-the-sanitize-module-plan-and-spec.md`.
- Related the review document to the primary plan and key implementation evidence files.

### Why

The primary plan was directionally good, but it contained several implementation-level risks. A mentoring review is valuable here because it teaches which details must be checked before converting a design document into code.

### What worked

- The review found concrete, file-backed issues rather than vague criticism.
- The strongest correction is the `SetExport` issue: dotted names are literal properties, not nested paths.
- The review also captures a missing Phase 0 for `go.mod` dependency wiring and local `replace` setup.

### What didn't work

- N/A — this step produced documentation only.

### What I learned

- The sanitize plan's high-level API shape is solid, but the implementation sketch needs tightening around goja conversion behavior.
- `go-go-goja/modules/yaml/yaml.go` is a better option-decoding reference than the Markdown module because Markdown does not decode an options object.

### What was tricky to build

- The tricky part was keeping the review constructive while still being specific about code-level mistakes. The review therefore separates strengths, risky details, recommended corrections, and final advice.

### What warrants a second pair of eyes

- Whether Phase 1 should expose `sanitize.json.strictParse(input)` directly or keep strict parse status only in JSON sanitize/lint results.
- Whether the review's recommendation to reject unknown options should become an explicit decision record in the primary design doc.

### What should be done in the future

- Update the primary design document with the review's concrete corrections before implementation starts.
- Add Phase 0 dependency wiring to the task list.

### Code review instructions

- Start with `design-doc/02-review-of-the-sanitize-module-plan-and-spec.md`.
- Compare its correction list against `design-doc/01-sanitize-native-module-design-and-implementation-guide.md`.
- Verify the `SetExport` critique against `/home/manuel/workspaces/2026-06-02/goja-text/go-go-goja/modules/exports.go`.

### Technical details

- Review doc path: `ttmp/2026/06/02/GOJA-TEXT-002--goja-text-module-bindings-sanitize-yaml-and-json-native-module/design-doc/02-review-of-the-sanitize-module-plan-and-spec.md`
- Important evidence files:
  - `/home/manuel/workspaces/2026-06-02/goja-text/go-go-goja/modules/exports.go`
  - `/home/manuel/workspaces/2026-06-02/goja-text/go-go-goja/modules/yaml/yaml.go`
  - `/home/manuel/workspaces/2026-06-02/goja-text/goja-text/go.mod`
