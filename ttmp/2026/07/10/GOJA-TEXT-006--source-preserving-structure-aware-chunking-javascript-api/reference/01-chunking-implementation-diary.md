---
Title: Chunking Implementation Diary
Ticket: GOJA-TEXT-006
Status: active
Topics:
    - goja
    - goja-bindings
    - markdown
    - native-modules
    - text-algorithms
    - xgoja
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://pkg/chunking/recursive.go
      Note: Recursive fallback and absolute range translation
    - Path: repo://pkg/chunking/validate.go
      Note: Lossless partition validator
    - Path: repo://pkg/markdown/source_ranges.go
      Note: Central exact Markdown source-range derivation
    - Path: repo://pkg/markdown/source_ranges_test.go
      Note: Structural syntax and Unicode coordinate contracts
    - Path: repo://ttmp/2026/07/10/GOJA-TEXT-006--source-preserving-structure-aware-chunking-javascript-api/scripts/01-inspect-goldmark-ranges.go
      Note: Reproducible Goldmark source-position probe
ExternalSources:
    - https://github.com/go-go-golems/goja-text/issues/9
Summary: Chronological implementation evidence, failures, decisions, commands, and review instructions for the goja-text chunking module.
LastUpdated: 2026-07-10T13:09:36.237896447-04:00
WhatFor: Reproduce the implementation, understand why contracts changed, and continue the work without rediscovering module and source-position constraints.
WhenToUse: Read before resuming GOJA-TEXT-006, reviewing a checkpoint, diagnosing a failed invariant, or preparing the final delivery.
---





# Chunking Implementation Diary

## Goal

Record the evidence and implementation sequence for GitHub issue #9: exact Markdown ranges, source-preserving segmenters, budgeted and recursive packers, the `require("chunking")` module, generated-host integration, documentation, and delivery.

## Step 1: Create the ticket and establish the implementation baseline

I created `GOJA-TEXT-006` before changing source so the design, experiments, failures, and code relations have a stable home. I also converted GitHub issue #9 into eight docmgr tasks that separate research, Markdown coordinates, segmenters, packing, JavaScript exposure, documentation, validation, and delivery.

The workspace checkout was clean on branch `task/goja-text-chunking`. A baseline `go test ./... -count=1` passed for every current package, so later failures can be attributed to this change rather than inherited repository state.

### Prompt Context

**User prompt (verbatim):** "Work in this workspace: /home/manuel/workspaces/2026-07-10/goja-text-chunking (docmgr is in goja-text/ttmp), and create a new ticket to work on the issue, and implement the new JS stuff. 

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Work only in the supplied WSM workspace, create a new goja-text docmgr ticket for GitHub issue #9, implement the complete native and JavaScript chunking API, document it for an intern, validate it, and deliver the ticket bundle to reMarkable.

**Inferred user intent:** Leave a working, reviewable implementation and enough durable technical context that a new engineer can understand, test, extend, and safely maintain it.

### What I did

- Confirmed the requested workspace and `goja-text` checkout.
- Read repository agent instructions, README, Makefile, current module/provider code, Markdown AST types, tests, help embedding, and generated xgoja spec.
- Created ticket `GOJA-TEXT-006` with design and reference documents.
- Added eight implementation tasks.
- Ran `go test ./... -count=1`.

### Why

- The issue changes a public JavaScript API, native Go domain logic, Markdown coordinates, generated TypeScript, provider registration, and help content. A ticket is required to keep those layers consistent.
- A clean baseline separates new regressions from existing failures.
- The task list provides checkpoints that can be implemented and reviewed independently.

### What worked

- Docmgr created the ticket at `ttmp/2026/07/10/GOJA-TEXT-006--source-preserving-structure-aware-chunking-javascript-api`.
- The branch was clean and correctly named for the work.
- All existing Go tests passed.
- Existing modules provide clear patterns for `modules.NativeModule`, `modules.TypeScriptDeclarer`, provider registration, embedded help, and runtime integration tests.

### What didn't work

- Two initial inspection commands referenced files that do not exist: `pkg/markdown/typescript.go` and `pkg/markdown/positions_test.go`. TypeScript declarations are implemented inside `pkg/markdown/module.go`, and source-position tests currently live in `pkg/markdown/module_test.go`. The missing-file messages were inspection mistakes, not repository failures.

### What I learned

- `goja-text` keeps each domain module in one package. Domain operations live in ordinary files, while `module.go` owns Goja glue and TypeScript declarations.
- Provider registration enumerates module names, resolves native modules from the shared registry, and forwards TypeScript descriptors.
- Go-backed values intentionally expose PascalCase fields to JavaScript, while module functions use lower camel case.
- The generated command module is committed under `cmd/goja-text` and must be regenerated after provider or jsverb changes.

### What was tricky to build

No behavior was implemented in this step. The main design constraint discovered is that Goldmark nodes expose source through different mechanisms: direct inline segments, block line segments, container children, and special HTML closure lines. One tested range helper must centralize those rules.

### What warrants a second pair of eyes

- The public coordinate semantics must not reinterpret existing `StartLine` and `StartColumn` behavior.
- The initial API needs to balance useful Go-backed domain types with JavaScript option ergonomics.
- The implementation phases are large enough that focused commits are preferable to one final monolithic change.

### What should be done in the future

- Implement and test exact Markdown ranges before building structural segmenters.
- Keep the design guide synchronized with actual API names as implementation proceeds.

### Code review instructions

- Start with `pkg/markdown/module.go`, `pkg/markdown/types.go`, and `pkg/markdown/convert.go`.
- Review `pkg/xgoja/providers/text/text.go` and `text_test.go` for generated-host packaging.
- Run `go test ./... -count=1` to reproduce the baseline.

### Technical details

Baseline command:

```bash
go test ./... -count=1
```

Result: all current packages passed.

## Step 2: Add exact Markdown byte, rune, and end positions

I extended `MarkdownNode` so every parsed node carries a half-open UTF-8 byte range, rune range, and end line/column. The implementation preserves existing start-line/start-column behavior and does not change `markdown.parse()` error semantics.

Goldmark does not expose all source ranges uniformly. The implementation uses exact segments for text and raw-HTML leaves, then uses the next sibling or ancestor sibling as the structural boundary for headings, paragraphs, lists, block quotes, code fences, HTML blocks, and thematic breaks. Trailing whitespace between blocks is excluded from the AST node range.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Implement the first accepted design phase and prove its source-coordinate semantics before building chunking algorithms.

**Inferred user intent:** Ensure all later chunks can cite exact original text instead of approximate JavaScript indices.

### What I did

- Added a Goldmark range inspection program under the ticket `scripts/` directory.
- Added `StartByte`, `EndByte`, `StartRune`, `EndRune`, `EndLine`, and `EndColumn` to `MarkdownNode`.
- Added a central `nodeSourceRange` helper.
- Updated Markdown TypeScript declarations.
- Added pure-Go tests for structural syntax, Unicode, inline emphasis/link syntax, rune counts, and end positions.
- Extended the runtime JavaScript test to read the new PascalCase fields.
- Ran `go test ./pkg/markdown -count=1` and `go test ./... -count=1`.

### Why

- Markdown `Lines()` often contains only body content and excludes syntax such as heading markers and fences.
- JavaScript UTF-16 indices cannot serve as byte or rune citations.
- Structural chunkers need exact block starts and verified source slices.

### What worked

- The range probe showed stable `Pos()` values for block markers and direct segments for text leaves.
- Tests confirm top-level ranges preserve heading, list, fenced-code, blockquote, HTML, and thematic-break syntax.
- Unicode rune counts remain consistent with UTF-8 byte slices.
- All existing module and provider tests still pass.

### What didn't work

- The first commit attempt was stopped by the pre-commit lint hook. `pkg/markdown/module.go` needed `gofmt`, and the inspection script called Goldmark's deprecated `Node.Text` method (`SA1019`). The production implementation and tests had passed; the failure was isolated to formatting and the ticket experiment. I removed the deprecated call, kept the probe focused on source coordinates, formatted both files, and reran the repository linter before retrying the commit.
- The probe also demonstrated that a naive `Lines()`-only approach would have omitted syntax; that approach was rejected before production code was written.

### What I learned

- A block node's next structural sibling is a more reliable outer boundary than its content lines.
- Inline nodes such as emphasis and links need surrounding Markdown markers, while leaf text nodes should keep their direct segment.
- Trimming only trailing Unicode whitespace gives structural nodes useful syntax-preserving envelopes without claiming inter-block separators.

### What was tricky to build

The helper must find an end boundary even when the node is the last child of several nested containers. It walks outward through ancestors until it finds a next sibling, otherwise it uses source end. Direct text and raw-HTML segments bypass this structural rule.

### What warrants a second pair of eyes

- Uncommon Goldmark extension nodes may have source behavior not covered by core-node fixtures.
- Trailing whitespace is intentionally excluded from AST node envelopes even though chunking segmenters will later partition separators losslessly.
- Invalid UTF-8 remains accepted by the existing Markdown parser; only chunking entry points will reject it.

### What should be done in the future

- Add any newly introduced Goldmark extension nodes to the range fixture when extensions are enabled.
- Keep range semantics documented in the Markdown API reference.

### Code review instructions

- Review `pkg/markdown/source_ranges.go` first.
- Confirm `pkg/markdown/source_ranges_test.go` slices the original source rather than comparing rendered text.
- Run:

```bash
go test ./pkg/markdown -count=1
go test ./... -count=1
```

### Technical details

Coordinate intervals are zero-based and half-open. Line/column positions remain one-based. For a heading `# H`, the node reports byte range `[0,3)` and end position `1:4`.

## Step 3: Implement lossless segmenters

I created `pkg/chunking` as a pure-Go domain package and implemented line, paragraph, Markdown-block, and Markdown-section segmenters. Every segmenter rejects invalid UTF-8, returns both byte and rune coordinates, and validates its own output before returning it.

### What I did

- Added the shared `Span`, `Diagnostic`, `StrategySpec`, and result types.
- Added one source index responsible for byte, rune, line, and column coordinates.
- Implemented LF and CRLF-aware line segmentation. When callers do not keep terminators on line spans, terminators remain present as explicit `lineTerminator` spans.
- Implemented paragraph separator ownership modes: `trailing`, `separate`, and `leading`.
- Implemented Markdown blocks by partitioning at consecutive top-level AST starts.
- Implemented Markdown sections with preamble handling and heading-path metadata.
- Added `ValidatePartition` and tests that slice the original source.

### Why

Separators cannot be discarded without invalidating source citations. A segmenter that reconstructs text from rendered Markdown also loses markers and exact whitespace. The implementation therefore assigns every input byte to exactly one span and treats derived heading paths as metadata.

### What worked

- LF, CRLF, Unicode, leading source whitespace, Markdown markers, fences, preambles, and nested heading paths pass focused tests.
- A fuzz target checks the line segmenter's lossless partition invariant over arbitrary valid UTF-8 strings.
- `go test ./pkg/chunking -count=1` passes.

### What didn't work

The first paragraph boundary draft compared the start of a blank run with the previous line end using a strict greater-than test. Those offsets are equal in the normal `paragraph + blank line` case, so the boundary would not have been emitted. I caught this during pre-test review, changed the boundary condition to distinguish leading whitespace from post-content blank runs, and then added an explicit `a\n\nb` ownership assertion.

### What I learned

- `keepTerminators: false` can remain source-preserving only if terminators become explicit spans.
- Top-level Goldmark node starts are sufficient to define a deterministic block partition; node ends are intentionally not used because inter-block separators still need an owner.
- A flat section partition and a hierarchical section tree are different APIs. This release returns a flat partition and records hierarchy in `HeadingPath`.

### What warrants a second pair of eyes

- Section boundaries start at every heading accepted by `maxHeadingLevel`; nested sections are represented as consecutive spans, not overlapping parent envelopes.
- Whitespace-only input is represented as one ordinary span rather than an empty result.
- Future sentence segmentation must preserve the same separator-ownership contract.

### Code review instructions

- Start with `pkg/chunking/validate.go` and `positions.go` to understand the invariant.
- Review each `segment_*.go` file together with `segment_test.go`.
- Run `go test ./pkg/chunking -count=1`.

## Step 4: Implement greedy, caller-weighted, and recursive packing

I implemented complete-span packing with byte, rune, or Unicode-whitespace word budgets. The regular packer supports a trailing whole-span overlap; the weighted packer accepts caller-computed nonnegative weights; and the recursive operation refines only oversized spans through an ordered sequence of segmenters.

### What I did

- Implemented budget validation and deterministic text measurement.
- Implemented greedy packing, whole-span overlap, and explicit `allow`/`error` oversized policy.
- Added stable oversized diagnostics at result and chunk scope.
- Implemented caller-weighted packing without adding a tokenizer dependency.
- Implemented recursive fallback with absolute coordinate translation and a final rune-window level.
- Added focused tests for exact budgets, overlap, oversized behavior, caller weights, fallback progress, and source slices.

### Why

Chunking and tokenization have separate release cycles. Accepting weights lets a downstream application use its model's tokenizer while this library remains deterministic and model-independent. Recursive fallback preserves large structures when possible and only uses fixed windows after stronger boundaries fail.

### What worked

- Whole-span overlap produces the expected repeated span without splitting it.
- An allowed oversized span is visibly marked and emits `span_exceeds_budget`; the error policy rejects it.
- Recursive fallback retains absolute source offsets even though each nested segmenter operates on a substring.
- `go test ./pkg/chunking -count=1` passes.

### What didn't work

The first core commit attempt passed all tests but the pre-commit linter reported `ineffassign` for `currentWeight = weight` in the oversized-span branch. `emit()` derives the emitted chunk weight from the immutable weight array, and the local accumulator is reset immediately afterward, so I removed the redundant assignment and reran lint before retrying the commit. During review, I also kept overlap removal in a loop so a large retained suffix can never prevent the next new span from being consumed.

### What was tricky to build

Nested segmenters naturally return coordinates relative to the substring they receive. `translateSpan` shifts byte, rune, line, and first-line column coordinates back into the original document before the recursive result is packed.

### What warrants a second pair of eyes

- Weighted overlap selects only complete trailing spans that fit both the overlap allowance and the next chunk budget.
- Atomic Markdown kinds remain oversized instead of being recursively split.
- Word measurement uses Go's Unicode whitespace fields and is not a model tokenizer.

### Code review instructions

- Review `pkg/chunking/pack.go` for the progress invariant and `recursive.go` for coordinate translation.
- Run `go test ./pkg/chunking -count=1` and inspect `pack_test.go`.

## Step 5: Expose the chunking API to JavaScript and xgoja

I added a native `require("chunking")` module, TypeScript declarations, strict JavaScript option decoders, runtime integration tests, and provider registration. The domain algorithms remain independent of Goja; `module.go` only validates JavaScript inputs and forwards them.

### What I did

- Implemented `modules.NativeModule` and `modules.TypeScriptDeclarer`.
- Exported `lines`, `paragraphs`, `markdownBlocks`, `markdownSections`, `pack`, `packWeighted`, and `recursive`.
- Added lower-camel plain-object decoders that reject unknown keys and incorrect primitive types.
- Preserved PascalCase Go-backed result fields.
- Registered `chunking` in the `goja-text` xgoja provider.
- Added runtime tests for segmentation, regular packing, weighted packing, recursive fallback, missing defaults, and option errors.
- Added provider-level validation for every TypeScript descriptor.
- Opened [go-go-goja issue #92](https://github.com/go-go-golems/go-go-goja/issues/92) proposing additive declaration builders and richer structured TypeScript nodes. This ticket continues with the current API and does not depend on that enhancement.

### Why

The JavaScript layer needs ergonomic plain objects while the returned values must remain consistent with existing goja-text modules. Strict decoding catches misspelled strategy settings before they silently change chunk boundaries. Provider registration makes the module available to generated xgoja applications and TypeScript tooling.

### What worked

- JavaScript can pass `blocks.Spans` directly into `pack` without copying fields.
- JavaScript can create `{ span, weight }` records for `packWeighted`.
- Omitted option objects and nested overlap objects receive documented defaults.
- The provider descriptor passes `tsgen/validate.Module`.
- `go test ./pkg/chunking ./pkg/xgoja/providers/text -count=1` passes.

### What didn't work

- The first adapter compile attempted to forward a two-value Go return directly into a generic helper after another argument. Go requires explicit `result, err` assignments in that call shape, so each wrapper now names both values.
- The initial TypeScript declaration used a nonexistent `spec.Optional(type)` helper. The pinned `go-go-goja` API represents declaration optionality as `spec.Param.Optional`; the declarations now use `Type: spec.Named(...)` with `Optional: true`.
- The first runtime test exposed that absent Goja object properties may arrive as `nil`, not only the JavaScript `undefined` singleton. `missingValue` now handles `nil`, `undefined`, and `null` uniformly.
- One runtime expectation asserted the parent heading path `Title`; the section under `## Detail` correctly returns the complete path `Title/Detail`. The test now verifies the documented hierarchical metadata.

### What I learned

- Optional declaration syntax and a union with `undefined` are different TypeScript contracts. The spec API correctly stores optionality on `Param` and `Field`.
- Goja object decoders must treat a nil property value as absent before calling `ToObject` or `Export`.
- Runtime integration tests are necessary because pure-Go option structs do not exercise JavaScript property projection and export conversion.

### What warrants a second pair of eyes

- The TypeScript interfaces are currently emitted through `RawDTS` while function signatures use structured descriptors; issue #92 proposes richer structured declaration support.
- `packWeighted` manually decodes its records so JavaScript can use lower-camel `span` and `weight` keys.
- Integer decoding intentionally rejects numeric strings and non-integral numbers.

### Code review instructions

- Review `pkg/chunking/module.go` beside `module_test.go`.
- Confirm provider registration in `pkg/xgoja/providers/text/text.go`.
- Run `go test ./pkg/chunking ./pkg/xgoja/providers/text -count=1`.
