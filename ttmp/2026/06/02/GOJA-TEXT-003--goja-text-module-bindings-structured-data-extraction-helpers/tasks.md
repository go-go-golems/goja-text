# Tasks

## Phase 0: Planning and scaffolding

- [x] 0.1 Create detailed phase/task checklist for GOJA-TEXT-003
- [x] 0.2 Create `pkg/extract` package skeleton
- [x] 0.3 Add package-level docs explaining extraction vs parsing vs sanitizing

## Phase 1: Source positions, candidate types, and options

- [x] 1.1 Implement byte-offset to row/column line index helpers
- [x] 1.2 Add tests for line index behavior, including EOF and multiline offsets
- [x] 1.3 Define `ExtractionCandidate`, `ExtractOptions`, `ExtractOptionsBuilder`, and validation result types
- [x] 1.4 Implement builder methods: `Formats`, `Tags`, `Extractors`, `IncludeDiagnostics`, `InferFormat`, `MinConfidence`, `MaxCandidates`, `Validate`, `Build`
- [x] 1.5 Add unit tests for options builder validation and defaults

## Phase 2: Wrapper extractors

- [x] 2.1 Implement Markdown fenced code block extraction for backtick and tilde fences
- [x] 2.2 Add Markdown fence tests for language/info strings, payload spans, unterminated blocks, and multiple blocks
- [x] 2.3 Implement XML-like tag wrapper extraction with default and caller-provided tags
- [x] 2.4 Add XML-like tag tests for multiline payloads, attributes, multiple tags, and missing close tags
- [x] 2.5 Implement YAML frontmatter extraction
- [x] 2.6 Add frontmatter tests for valid frontmatter, missing close delimiter, and non-frontmatter Markdown

## Phase 3: Raw structured recognition and validation

- [x] 3.1 Implement raw JSON recognition using strict JSON parse and sanitize-backed repair checks
- [x] 3.2 Implement conservative raw YAML recognition with false-positive avoidance
- [x] 3.3 Add raw structured tests for strict JSON, repairable JSON, simple YAML, prose false positives, and empty input
- [x] 3.4 Implement sanitize-backed `Validate(candidate, options)` for JSON and YAML candidates
- [x] 3.5 Add validation tests for valid, repaired, invalid, and unknown-format candidates
- [x] 3.6 Implement `All(input, options)` to merge selected extractors in source order
- [x] 3.7 Add combined extraction tests with overlapping Markdown/XML/raw payloads

## Phase 4: Native module and JavaScript runtime tests

- [x] 4.1 Implement `extract` NativeModule exports
- [x] 4.2 Implement namespace-aware TypeScript declarations with `RawDTS`
- [x] 4.3 Add runtime tests for `require("extract")`
- [x] 4.4 Add runtime tests for Markdown codeblocks, XML-like tags, raw JSON/YAML, frontmatter, `all`, and `validate`
- [x] 4.5 Confirm PascalCase Go-backed candidate fields from JavaScript and lowercase absence where relevant

## Phase 5: xgoja integration and examples

- [x] 5.1 Blank-import `pkg/extract` in the xgoja text provider
- [x] 5.2 Add `extract` to `textModuleNames`
- [x] 5.3 Add `extract` to `xgoja.yaml`
- [x] 5.4 Add `examples/text/structured-data-sample.md`
- [x] 5.5 Add `examples/js/extract-demo.js`
- [x] 5.6 Add `smoke-extract` to the Makefile and include it in `smoke`

## Phase 6: Documentation, validation, and delivery

- [x] 6.1 Update README with extract module examples
- [x] 6.2 Run `go test ./... -count=1`
- [x] 6.3 Run `GOWORK=off go test ./... -count=1`
- [x] 6.4 Run `make check`
- [x] 6.5 Update diary and changelog after each implementation phase
- [x] 6.6 Relate modified files to ticket docs
- [x] 6.7 Run `docmgr doctor --ticket GOJA-TEXT-003 --stale-after 30`
- [x] 6.8 Upload updated bundle to reMarkable
