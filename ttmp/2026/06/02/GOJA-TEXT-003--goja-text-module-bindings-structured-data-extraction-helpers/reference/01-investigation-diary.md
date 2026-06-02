---
Title: Investigation Diary
Ticket: GOJA-TEXT-003
Status: active
Topics:
    - goja
    - goja-bindings
    - text-algorithms
    - native-modules
    - markdown
    - json
    - yaml
    - structured-data
    - xml
    - extraction
DocType: reference
Intent: Chronological diary for structured-data extraction helper design and implementation
Owners: []
RelatedFiles:
    - Path: Makefile
      Note: smoke-extract and check targets
    - Path: README.md
      Note: extract user documentation
    - Path: examples/js/extract-demo.js
      Note: extract smoke script
    - Path: examples/text/structured-data-sample.md
      Note: extract demo fixture
    - Path: pkg/extract/all.go
      Note: Combined extraction
    - Path: pkg/extract/doc.go
      Note: Package-level extraction documentation
    - Path: pkg/extract/format.go
      Note: Format inference helpers
    - Path: pkg/extract/frontmatter.go
      Note: YAML frontmatter extraction
    - Path: pkg/extract/markdown_fences.go
      Note: Markdown fenced code block extraction
    - Path: pkg/extract/module.go
      Note: NativeModule exports
    - Path: pkg/extract/module_test.go
      Note: JavaScript runtime tests
    - Path: pkg/extract/options_test.go
      Note: Options builder tests
    - Path: pkg/extract/positions.go
      Note: Source position infrastructure
    - Path: pkg/extract/positions_test.go
      Note: Line index tests
    - Path: pkg/extract/raw.go
      Note: Raw JSON/YAML recognition
    - Path: pkg/extract/raw_validate_test.go
      Note: Raw/validation/all tests
    - Path: pkg/extract/types.go
      Note: Candidate and options model
    - Path: pkg/extract/typescript.go
      Note: TypeScript declarations
    - Path: pkg/extract/validate.go
      Note: Sanitize-backed candidate validation
    - Path: pkg/extract/wrappers_test.go
      Note: Wrapper extractor tests
    - Path: pkg/extract/xml_tags.go
      Note: XML-like tag wrapper extraction
    - Path: pkg/xgoja/providers/text/text.go
      Note: xgoja provider wiring for extract
    - Path: xgoja.yaml
      Note: extract module runtime entry
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---






# Investigation Diary

## Goal

Capture the design and implementation process for GOJA-TEXT-003: structured-data extraction helpers for code blocks, XML-like wrappers, raw JSON/YAML recognition, frontmatter, and related deterministic extraction primitives.

---

## Step 1: Create Structured Data Extraction Ticket and Design Guide

Created GOJA-TEXT-003 after closing GOJA-TEXT-002. The new ticket designs an `extract` module that locates structured payload candidates inside messy text while preserving source spans and wrapper metadata. The design intentionally separates extraction from parsing and repair: extraction finds candidates, while validation can use the existing sanitize package for JSON/YAML.

### Prompt Context

**User prompt (verbatim):** "do it. then close the ticket.

Then open a new ticket where we are going to provide helpers for extracting structured data from text:

- codeblocks from markdown
- xml tag wrapped 
- recognizing raw json / yaml
- other suggestions you might have.

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Finish GOJA-TEXT-002 by adding validation targets and closing it, then create a new docmgr ticket and intern-ready design guide for a structured-data extraction module.

**Inferred user intent:** Continue evolving goja-text from format-specific parsing/sanitizing modules into higher-level helpers for extracting structured payloads from unstructured text.

### What I did

- Added Makefile validation targets to GOJA-TEXT-002 and validated with `make check`.
- Closed GOJA-TEXT-002.
- Created GOJA-TEXT-003 with topics covering extraction, structured data, Markdown, XML, JSON, and YAML.
- Added primary design doc: `design-doc/01-structured-data-extraction-helpers-design-and-implementation-guide.md`.
- Added this diary document.
- Designed a new `require("extract")` module with helpers for:
  - Markdown fenced code blocks
  - XML-like tag wrappers
  - raw JSON/YAML recognition
  - YAML frontmatter
  - combined extraction
  - sanitize-backed validation
- Included decision records, algorithms, pseudocode, file layout, testing strategy, risks, open questions, and implementation checklist.

### Why

Structured data appears in many wrappers inside text. The Markdown and sanitize modules already parse and repair formats, but callers still need deterministic span-preserving extraction helpers. The new module should locate candidates and preserve provenance before validation or parsing.

### What worked

- The existing Markdown module provides a useful reference for codeblock semantics.
- The sanitize module provides validation and repair semantics for JSON/YAML candidates.
- The builder/config pattern from GOJA-TEXT-002 gives a clear options approach for `extract.options()`.

### What didn't work

- N/A — this step produced documentation only.

### What I learned

- Extraction should be treated as a separate responsibility from parsing and repair. Returning parsed values too early would discard source-span and wrapper metadata.

### What was tricky to build

- The main design challenge was avoiding overclaiming. XML-like tags should not be documented as full XML parsing, and raw YAML recognition must be conservative to avoid false positives.

### What warrants a second pair of eyes

- Whether `extract.all` should keep overlapping candidates by default or apply an overlap policy.
- Whether XML-like tag extraction should support nested same-name tags in Phase 1.
- Whether TOML/JSON frontmatter should be Phase 1 or deferred.

### What should be done in the future

- Implement the extraction package in phases, starting with source-position helpers and Markdown fenced code blocks.
- Add runtime tests before xgoja integration.

### Code review instructions

- Start with the decision records in the design doc.
- Then review extractor-specific algorithms for Markdown fences, XML-like tags, raw structured recognition, and frontmatter.
- Check that the implementation plan keeps domain extraction code separate from goja module wiring.

### Technical details

- Ticket path: `ttmp/2026/06/02/GOJA-TEXT-003--goja-text-module-bindings-structured-data-extraction-helpers/`
- Design doc: `design-doc/01-structured-data-extraction-helpers-design-and-implementation-guide.md`

---

## Step 2: Expand GOJA-TEXT-003 into Detailed Implementation Phases

Converted the initial high-level task list into detailed implementation phases. The ticket now has explicit tasks for planning, source-position infrastructure, candidate/options types, wrapper extractors, raw JSON/YAML recognition, sanitize-backed validation, native module wiring, xgoja integration, examples, README updates, validation commands, and reMarkable delivery.

### Prompt Context

**User prompt (verbatim):** "Now add a detailed set of phases + tasks for each phase to the ticket, then work on them one by one, committing at appropriate intervals, and keeping a detailed log."

**Assistant interpretation:** Replace the coarse GOJA-TEXT-003 task list with phase-level and subtask-level checkboxes, then proceed through implementation with commits and diary updates.

**Inferred user intent:** Make the implementation process reviewable and resumable, not a single large untracked coding pass.

### What I did

- Rewrote `tasks.md` with six phases:
  - planning and scaffolding
  - source positions, candidate types, and options
  - wrapper extractors
  - raw structured recognition and validation
  - native module and JavaScript runtime tests
  - xgoja integration, docs, validation, and delivery
- Marked task 0.1 complete.

### Why

The initial task list was useful but too coarse for step-by-step implementation. Detailed phases make commit boundaries and diary entries clearer.

### What worked

- The task list now maps directly to implementable files and validation checkpoints.

### What didn't work

- N/A — planning step only.

### What I learned

- The extraction work has enough surface area that separate commits for positions/types, wrapper extractors, raw/validation, and xgoja integration will make review easier.

### What was tricky to build

- The task split needed to keep extraction concerns separate from goja module wiring so domain logic can be tested independently.

### What warrants a second pair of eyes

- Whether raw YAML recognition should be implemented before or after wrapper extraction; the current plan puts it after wrapper extraction to reduce false-positive risk.

### What should be done in the future

- Implement Phase 1 next: line index helpers, candidate types, and options builder.

### Code review instructions

- Review `tasks.md` for implementation sequencing.

### Technical details

- Updated file: `tasks.md`

---

## Step 3: Implement Extract Package Skeleton, Source Positions, Candidate Types, and Options

Implemented the first code slice for `pkg/extract`: package documentation, source-position helpers, candidate/config types, a Go-backed options builder, and unit tests. This creates the foundation all extractors will use to report byte spans, row/column spans, allowed formats, enabled extractors, confidence thresholds, and candidate limits.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue the detailed task plan by implementing the first phase and committing after validation.

**Inferred user intent:** Build the extraction module in reviewable layers, starting with reusable infrastructure rather than jumping directly into parser logic.

### What I did

- Added `pkg/extract/doc.go` to document extraction vs parsing/sanitizing.
- Added `pkg/extract/positions.go` with:
  - `lineIndex`
  - byte offset to row/column conversion
  - source line splitting with byte offsets
  - span population helper
- Added `pkg/extract/types.go` with:
  - `ExtractionCandidate`
  - `ExtractOptions`
  - `ExtractOptionsBuilder`
  - `CandidateValidationResult`
  - option filtering helpers
- Added tests:
  - `positions_test.go`
  - `options_test.go`
- Ran `gofmt -w pkg/extract`.
- Ran `go test ./... -count=1` successfully.

### Why

Every extractor needs reliable source spans and a shared candidate representation. Implementing these first prevents each extractor from inventing its own position and filtering logic.

### What worked

- The line index and option builder tests pass.
- The builder defaults are explicit and include default XML-like tags and default extractors.

### What didn't work

- The first option-builder test failed because extractor names were normalized to lowercase while `knownExtractor` expected camelCase. I fixed this by using lowercase canonical extractor keys internally.
- One line-index test expected offset `11` to still be on the previous line; for `alpha\nbeta\ngamma`, offset `11` is the start of `gamma`, so the correct position is row 2 column 0.

### What I learned

- Canonical internal extractor IDs should be lowercase (`markdowncodeblocks`, `xmltagged`, `rawstructured`) even if public function names use camelCase.
- Byte-offset row/column behavior at newline boundaries should be pinned early because all span tests depend on it.

### What was tricky to build

- The line-index helper uses byte columns, not rune columns. This matches source byte spans and tree-sitter-style positions better than visual columns, but it should be documented if UTF-8 display columns ever matter.

### What warrants a second pair of eyes

- Whether public `Extractors(...)` should accept camelCase names and normalize them more intentionally instead of simply lowercasing all input.

### What should be done in the future

- Implement Markdown fenced code block extraction next.

### Code review instructions

- Start with `pkg/extract/types.go` to understand the public candidate/config model.
- Then review `pkg/extract/positions.go` for span semantics.
- Validate with `go test ./... -count=1`.

### Technical details

- Validation command: `go test ./... -count=1`

---

## Step 4: Implement Wrapper Extractors

Implemented the first payload extractors: Markdown fenced code blocks, XML-like tag wrappers, and YAML frontmatter. These extractors return Go-backed `ExtractionCandidate` values with raw wrapper text, payload text, byte spans, row/column spans, inferred format, wrapper kind, label, and optional diagnostics.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue implementing the GOJA-TEXT-003 phases in order and commit after validation.

**Inferred user intent:** Build deterministic structured-data location helpers before adding parsing or validation behavior.

### What I did

- Added `pkg/extract/format.go` for label/payload format inference.
- Added `pkg/extract/markdown_fences.go`:
  - supports backtick and tilde fences
  - captures info string, language label, payload text, raw wrapper text, and spans
  - reports unterminated fence diagnostics when requested
- Added `pkg/extract/xml_tags.go`:
  - extracts simple same-name XML-like wrappers
  - supports default/caller-provided tags and attributes in opening tags
  - reports missing close tag diagnostics when requested
- Added `pkg/extract/frontmatter.go`:
  - extracts leading YAML frontmatter delimited by `---`
  - reports missing close delimiter diagnostics when requested
- Added `pkg/extract/wrappers_test.go`.
- Ran `gofmt -w pkg/extract`.
- Ran `go test ./... -count=1` successfully.

### Why

These wrapper-based extractors are the lowest-false-positive extraction layer. They identify explicit source wrappers before raw JSON/YAML heuristics attempt to infer structure from unwrapped text.

### What worked

- Markdown code block extraction handles both backtick and tilde fences.
- XML-like tag extraction handles attributes and multiline payloads.
- Frontmatter extraction preserves wrapper delimiters in `Raw` and payload-only text in `Text`.

### What didn't work

- N/A; tests passed after adding the missing payload-format inference helper.

### What I learned

- The extractors need a clear distinction between wrapper span and payload span. This is why `Raw` and `Text`, plus `StartByte`/`PayloadStartByte`, are both necessary.

### What was tricky to build

- XML-like extraction needed to avoid claiming full XML behavior. The implementation is intentionally simple same-name tag matching with attributes preserved as raw opening-tag info.

### What warrants a second pair of eyes

- Whether missing-close XML tags should emit partial candidates only when diagnostics are enabled, as currently implemented.
- Whether Markdown fence indentation should exactly match CommonMark or remain a pragmatic subset.

### What should be done in the future

- Implement raw JSON/YAML recognition and sanitize-backed validation next.

### Code review instructions

- Start with `pkg/extract/markdown_fences.go`, then `xml_tags.go`, then `frontmatter.go`.
- Check tests in `pkg/extract/wrappers_test.go`.
- Validate with `go test ./... -count=1`.

### Technical details

- Validation command: `go test ./... -count=1`

---

## Step 5: Implement Raw Structured Recognition, Validation, and Combined Extraction

Implemented raw JSON/YAML recognition, sanitize-backed candidate validation, and the combined `All` extraction path. This adds the first layer that can infer structure without explicit Markdown fences, XML-like tags, or frontmatter wrappers.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue the phased implementation after wrapper extraction passed tests.

**Inferred user intent:** Add validation-aware extraction behavior while keeping raw detection conservative.

### What I did

- Added `pkg/extract/raw.go`:
  - strict JSON recognition with `encoding/json`
  - repairable JSON recognition through `jsonsanitize.Sanitize`
  - conservative YAML recognition requiring multiple mapping/list indicators
- Added `pkg/extract/validate.go`:
  - JSON candidate validation/repair
  - YAML candidate validation/repair
  - minimal XML-like wrapper validation
  - unknown-format errors
- Added `pkg/extract/all.go`:
  - runs enabled extractors
  - merges candidates in source order
  - applies option filtering
- Added `pkg/extract/raw_validate_test.go` covering strict JSON, repairable JSON, YAML, prose false positives, validation, and combined extraction.
- Ran `gofmt -w pkg/extract`.
- Ran `go test ./... -count=1` successfully.

### Why

Wrappers are precise but not always present. Raw recognition lets callers handle whole-input JSON/YAML payloads, while sanitize-backed validation gives candidates a deterministic validity/repair signal.

### What worked

- Strict JSON candidates receive high confidence.
- Repairable JSON candidates are recognized with lower confidence.
- Simple YAML blocks are recognized, while one-colon prose is rejected.
- `Validate` repairs malformed JSON/YAML candidates through the sanitize library.

### What didn't work

- The first `All` test assumed the Markdown code block would be the second candidate. Raw YAML recognition also emitted a whole-document YAML-like candidate before the Markdown block because it starts earlier. I changed the test to assert required candidates are present while only requiring frontmatter to be first.

### What I learned

- Combined extraction can legitimately return overlapping candidates. The design's choice to keep overlaps in Phase 1 was correct; tests should not assume a single non-overlapping sequence.

### What was tricky to build

- Raw YAML recognition needed false-positive avoidance. The implementation requires at least two mapping-like lines or a mapping plus list-like line before attempting sanitize-backed acceptance.

### What warrants a second pair of eyes

- Whether whole-document raw YAML candidates should be suppressed when stronger wrapper candidates exist.
- Whether candidate confidence values should be documented as stable or treated as heuristic.

### What should be done in the future

- Implement the `extract` NativeModule and runtime tests next.

### Code review instructions

- Review `pkg/extract/raw.go`, `validate.go`, and `all.go` together.
- Check `raw_validate_test.go` for intended raw detection and overlap behavior.
- Validate with `go test ./... -count=1`.

### Technical details

- Validation command: `go test ./... -count=1`

---

## Step 6: Implement Extract NativeModule and Runtime Tests

Exposed the extractor package through `require("extract")` and added JavaScript runtime tests. The native module now exports options, wrapper extractors, raw structured recognition, frontmatter extraction, combined extraction, and candidate validation.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue implementation by wiring the domain extraction package into go-go-goja's NativeModule system.

**Inferred user intent:** Make the extraction helpers available to JavaScript with the same Go-backed object pattern used by Markdown and sanitize.

### What I did

- Added `pkg/extract/module.go` implementing `modules.NativeModule`.
- Added `pkg/extract/typescript.go` with namespace-aware `RawDTS` declarations.
- Added `pkg/extract/module_test.go` runtime tests for:
  - `require("extract")`
  - `extract.options().Build()`
  - `markdownCodeBlocks`
  - `xmlTagged`
  - `frontmatter`
  - `rawStructured`
  - `all`
  - `validate`
  - PascalCase candidate field access and lowercase absence
- Ran `gofmt -w pkg/extract`.
- Ran `go test ./... -count=1` successfully.

### Why

The domain functions need to be validated through the actual goja runtime because public behavior depends on Go-backed field and method projection.

### What worked

- The runtime tests confirm `ExtractionCandidate` fields are visible in JavaScript as `Kind`, `Format`, `Text`, etc.
- `extract.validate(candidate)` works with candidates returned by `rawStructured`.

### What didn't work

- N/A; runtime tests passed on the first run.

### What I learned

- The flat `extract` module is simpler than the nested `sanitize` module because all operations share one candidate model and one options builder.

### What was tricky to build

- The module adapter must keep optional config handling simple. It accepts nil options for defaults, matching the underlying Go functions.

### What warrants a second pair of eyes

- Whether `CandidateValidationResult.Fixes`/`Issues` should stay `any` or become separate typed JSON/YAML validation result variants.

### What should be done in the future

- Wire `extract` into xgoja and add a file-backed demo script.

### Code review instructions

- Review `pkg/extract/module.go` and `pkg/extract/module_test.go` together.
- Validate with `go test ./... -count=1`.

### Technical details

- Validation command: `go test ./... -count=1`

---

## Step 7: Wire Extract into xgoja, Add Demo, README, and Full Validation

Integrated `extract` into the generated goja-text binary and added a file-backed demo. The `make check` target now runs normal tests, standalone tests, xgoja build, Markdown smoke, sanitize smoke, and extract smoke.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Finish the implementation path with xgoja integration, examples, README updates, and full validation.

**Inferred user intent:** Make the structured-data extraction helpers usable through the same generated binary as Markdown and sanitize.

### What I did

- Updated `pkg/xgoja/providers/text/text.go`:
  - blank-imported `pkg/extract`
  - added `extract` to `textModuleNames`
- Updated `xgoja.yaml` to include `extract` in the `main` runtime.
- Added `examples/text/structured-data-sample.md`.
- Added `examples/js/extract-demo.js`.
- Updated `Makefile`:
  - added `smoke-extract`
  - included it in `smoke`
- Updated `README.md` with extract usage and xgoja smoke command.
- Ran validation:
  - `go test ./... -count=1`
  - `GOWORK=off go test ./... -count=1`
  - `make check`

### Why

The generated xgoja binary is the user-facing exercise harness. The extract module is not complete until it can be loaded through `require("extract")` in that binary and run against a real text file from disk.

### What worked

- All tests passed in normal and standalone modes.
- xgoja build passed.
- Markdown, sanitize, and extract smoke scripts all passed.
- `extract-demo.js` returned frontmatter, raw YAML-like whole-document candidate, Markdown JSON code block, and XML-tagged YAML candidates.

### What didn't work

- N/A. The demo reveals one expected Phase 1 behavior: `extract.all` keeps overlapping/whole-document raw candidates rather than suppressing them when more specific wrapper candidates exist.

### What I learned

- The overlap policy is visible in real demo output. Keeping all candidates is useful for transparency, but future work may want a `PreferWrapped` or `SuppressOverlaps` option.

### What was tricky to build

- The demo validates a realistic mixed document and shows why provenance fields matter: `Kind`, `Wrapper`, `Label`, and `StartRow` make it clear where each payload came from.

### What warrants a second pair of eyes

- Whether raw whole-document YAML candidates should be emitted by default when a Markdown document contains frontmatter and other wrappers.
- Whether `extract.validate` should return typed `JsonValidationResult` / `YamlValidationResult` instead of `Fixes` and `Issues` as `any`.

### What should be done in the future

- Add an overlap policy option if callers find the current keep-all behavior too noisy.
- Consider TOML/JSON frontmatter support after YAML frontmatter stabilizes.

### Code review instructions

- Run `make check`.
- Inspect `examples/js/extract-demo.js` and its output.
- Review provider and `xgoja.yaml` wiring.

### Technical details

- Validation command: `make check`
