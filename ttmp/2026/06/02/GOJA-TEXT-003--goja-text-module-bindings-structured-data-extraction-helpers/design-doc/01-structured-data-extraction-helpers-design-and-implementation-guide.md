---
Title: Structured Data Extraction Helpers Design and Implementation Guide
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
DocType: design-doc
Intent: Intern-ready design and implementation guide for structured-data extraction helpers
Owners: []
RelatedFiles:
    - Path: Makefile
      Note: Validation target pattern for future extract smoke target
    - Path: pkg/markdown/convert.go
      Note: Existing Markdown fenced code field extraction reference
    - Path: pkg/markdown/module.go
      Note: NativeModule pattern for text modules
    - Path: pkg/sanitize/module.go
      Note: Namespace-based module pattern and sanitize validation reference
    - Path: pkg/sanitize/module_test.go
      Note: Runtime test pattern for Go-backed objects
    - Path: pkg/sanitize/options.go
      Note: Go-backed builder/config pattern
    - Path: pkg/xgoja/providers/text/text.go
      Note: xgoja provider registration list
    - Path: xgoja.yaml
      Note: Generated binary module composition
ExternalSources: []
Summary: Design for a goja-text module that extracts structured data candidates from Markdown, XML-tag-wrapped text, raw JSON/YAML, frontmatter, and LLM-style wrappers.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Structured Data Extraction Helpers Design and Implementation Guide

This document designs the next `goja-text` capability: helpers for extracting structured data from messy text. The first two modules established the foundation. `require("markdown")` parses Markdown into Go-backed AST nodes. `require("sanitize")` repairs and validates YAML and JSON. GOJA-TEXT-003 should build on those modules by locating structured payloads inside larger text: fenced code blocks, XML-like tags, raw JSON/YAML documents, YAML frontmatter, and common LLM response wrappers.

The goal is not to guess intent with a large language model. The goal is to provide deterministic extraction primitives that JavaScript scripts can combine. A caller should be able to say: "Find candidate JSON payloads in this text, tell me where they came from, validate or sanitize them if possible, and preserve enough source location data that I can show the user what was extracted."

---

## Part 0: Executive Summary

GOJA-TEXT-003 should add a new native module, tentatively named `extract`, exposed as `require("extract")`. The module should scan text and return Go-backed extraction candidates. A candidate is a typed span of source text with metadata: kind, format, start/end offsets, line/column positions, wrapper details, confidence, and optional validation status.

The proposed Phase 1 helpers are:

- `extract.markdownCodeBlocks(input, options?)` — extract fenced Markdown code blocks, including language/info string and source spans.
- `extract.xmlTagged(input, options?)` — extract content wrapped in XML-like tags such as `<json>...</json>`, `<yaml>...</yaml>`, `<data>...</data>`, and caller-provided tag names.
- `extract.rawStructured(input, options?)` — recognize whole-input or substring candidates that look like raw JSON or YAML.
- `extract.frontmatter(input, options?)` — extract YAML/TOML/JSON frontmatter from Markdown-style documents.
- `extract.all(input, options?)` — run selected extractors and return a merged ordered candidate list.
- `extract.validate(candidate, options?)` — validate or sanitize a candidate using the existing `sanitize` module/library semantics.

The module should not parse every candidate into JavaScript values in Phase 1. Extraction and parsing are separate responsibilities. Extraction finds candidate spans. Validation checks whether candidates are valid or repairable. Parsing into application data can happen later through `sanitize`, core `yaml`, `JSON.parse`, or future helper functions.

---

## Part 1: Problem Statement

Structured data often appears inside unstructured text. This is especially common in LLM output, documentation, logs, chat transcripts, issue descriptions, and Markdown notes. The same payload can appear in many wrappers:

```md
Here is the JSON:

```json
{"name":"Alice","age":30}
```
```

```xml
<json>
{"name":"Alice"}
</json>
```

```text
---
title: Example
tags:
  - demo
---

# Document
```

```text
The configuration is:
name: Alice
age: 30
```

The caller needs deterministic tools that answer:

1. Where are the structured payloads?
2. What format do they appear to be?
3. What wrapper produced the candidate?
4. What source span should be highlighted?
5. Is the candidate valid JSON/YAML/XML-like content?
6. If invalid, is it repairable by the sanitize module?

Without extraction helpers, every script has to reimplement brittle regexes. That creates inconsistent behavior across goja-text users. GOJA-TEXT-003 should provide the shared extraction layer.

---

## Part 2: Existing Foundation

### Markdown module

The current `markdown` module already parses Markdown with goldmark and exposes Go-backed nodes:

- `parse(input)`
- `walk(root, visitor)`
- `textContent(node)`
- `validate(value)`

Relevant files:

- `goja-text/pkg/markdown/types.go`
- `goja-text/pkg/markdown/convert.go`
- `goja-text/pkg/markdown/module.go`
- `goja-text/pkg/markdown/parser_test.go`
- `goja-text/pkg/markdown/module_test.go`

The Markdown converter already captures fenced code block metadata:

- `Type: "fencedCodeBlock"`
- `Language`
- `Info`
- `Text`
- `SourcePos`

However, the current `MarkdownNode` does not include byte offsets or exact fence spans. For extraction, line/column and byte span precision matter. Phase 1 can either:

1. Re-scan Markdown fences directly with a lightweight scanner, or
2. Extend Markdown conversion to expose byte spans.

The recommended Phase 1 path is a dedicated fence scanner in the new `extract` package. It avoids changing the public Markdown AST contract and can capture exact opening fence, content, closing fence, info string, and byte spans.

### Sanitize module

The current `sanitize` module wraps `github.com/go-go-golems/sanitize v0.0.2` and exposes YAML/JSON repair and validation with Go-backed builder/config objects.

Relevant files:

- `goja-text/pkg/sanitize/types.go`
- `goja-text/pkg/sanitize/options.go`
- `goja-text/pkg/sanitize/module.go`
- `goja-text/pkg/sanitize/module_test.go`
- `goja-text/examples/js/sanitize-demo.js`

GOJA-TEXT-003 should reuse sanitize semantics for validation. It should not duplicate JSON/YAML repair logic. If `extract.validate(candidate)` needs to check JSON, it should call the sanitize library directly from Go or reuse internal helper code that mirrors the `sanitize` module behavior.

### xgoja provider

The goja-text provider currently registers `markdown` and `sanitize`:

- `goja-text/pkg/xgoja/providers/text/text.go`
- `goja-text/xgoja.yaml`

GOJA-TEXT-003 should add `extract` to the same provider and runtime module list.

---

## Part 3: Proposed Module Shape

### Entry point

```js
const extract = require("extract");
```

### Core API

```js
const blocks = extract.markdownCodeBlocks(text, options);
const tagged = extract.xmlTagged(text, options);
const raw = extract.rawStructured(text, options);
const fm = extract.frontmatter(text, options);
const all = extract.all(text, options);

const validation = extract.validate(all[0], options);
```

### Candidate model

All extractors return `ExtractionCandidate` objects:

```ts
interface ExtractionCandidate {
  Kind: string;              // "markdownCodeBlock", "xmlTagged", "raw", "frontmatter"
  Format: string;            // "json", "yaml", "xml", "toml", "text", "unknown"
  Text: string;              // payload text without wrapper when possible
  Raw: string;               // full matched source including wrapper
  Wrapper: string;           // "markdownFence", "xmlTag", "frontmatter", "none"
  Label?: string;            // fence language, XML tag name, or frontmatter delimiter label
  Info?: string;             // full markdown fence info string
  StartByte: number;
  EndByte: number;
  StartRow: number;          // 0-indexed
  StartColumn: number;
  EndRow: number;
  EndColumn: number;
  PayloadStartByte: number;
  PayloadEndByte: number;
  PayloadStartRow: number;
  PayloadStartColumn: number;
  PayloadEndRow: number;
  PayloadEndColumn: number;
  Confidence: number;        // 0.0..1.0 deterministic heuristic score
  Diagnostics?: string[];
}
```

Keep this as a Go-backed struct. JavaScript callers access PascalCase fields:

```js
for (const c of extract.all(text)) {
  console.log(c.Kind, c.Format, c.StartRow, c.Text);
}
```

### Validation model

```ts
interface CandidateValidationResult {
  Candidate: ExtractionCandidate;
  Valid: boolean;
  Format: string;
  Sanitized?: string;
  Errors?: string[];
  LintIssues?: unknown[];
  Fixes?: unknown[];
}
```

Validation should be conservative:

- JSON: use `jsonsanitize.StrictParse` and/or `jsonsanitize.SanitizeWithOptions`.
- YAML: use `yamlsanitize.SanitizeWithOptions` and lint status.
- XML-like wrappers: Phase 1 should validate matching tags only, not full XML documents.
- Unknown format: return `Valid: false` with a diagnostic unless `options.inferFormat` is enabled.

---

## Part 4: Extractor Details

## 4.1 Markdown code block extraction

### What it should detect

Markdown fenced code blocks with backtick or tilde fences:

````md
```json
{"ok": true}
```

~~~yaml
name: Alice
~~~
````

### Why not only use goldmark?

Goldmark already identifies fenced code blocks, but the existing `MarkdownNode` does not expose exact byte spans or the raw wrapper. Extraction needs exact source spans and raw wrapper text. A direct scanner can produce better extraction metadata without changing the Markdown AST contract.

### Algorithm sketch

```text
scanMarkdownCodeBlocks(input):
    lineIndex = build line/byte index
    candidates = []
    i = 0

    while i < len(lines):
        line = lines[i]
        if line starts with optional spaces + fence marker (` or ~ repeated at least 3):
            markerChar = ` or ~
            markerLen = count markerChar
            info = rest of opening line trimmed
            language = first word of info lowercased
            startByte = line start byte + opening fence column
            payloadStart = byte after opening line newline

            j = i + 1
            while j < len(lines):
                if line j has closing fence with same marker char and len >= markerLen:
                    payloadEnd = line j start byte
                    endByte = end of closing line including newline if present
                    emit candidate
                    i = j + 1
                    continue outer
                j++

            emit unterminated candidate with diagnostic
            break

        i++
```

### Format inference

Fence info maps naturally to `Format`:

| Info/language | Format |
| --- | --- |
| `json`, `jsonc` | `json` |
| `yaml`, `yml` | `yaml` |
| `xml` | `xml` |
| `toml` | `toml` |
| empty | `unknown` |
| anything else | lowercased label or `text` depending options |

### Important edge cases

- Tilde fences should work because tests already used tilde fences to avoid Go raw-string backtick issues.
- Opening fence indentation up to three spaces should be accepted; deeper indentation may be indented code.
- Closing fence must use the same marker character.
- Closing fence length must be at least opening length.
- Info string should be preserved exactly.
- Payload should exclude opening and closing fence lines.

---

## 4.2 XML tag-wrapped extraction

### What it should detect

LLM and tool outputs often wrap payloads in simple XML-like tags:

```xml
<json>{"ok": true}</json>
<yaml>
name: Alice
</yaml>
<answer>
{"ok": true}
</answer>
```

The goal is not to implement a full XML parser. The goal is to extract balanced, same-name, non-overlapping tag wrappers from text.

### Proposed API options

```ts
interface XmlTaggedOptions {
  Tags?: string[];             // allowed tags; default common tags
  InferFormatFromTag?: boolean; // default true
  AllowAttributes?: boolean;    // default true
  MaxDepth?: number;            // default 1 for Phase 1
}
```

Default tags:

- `json`
- `yaml`
- `xml`
- `data`
- `result`
- `answer`
- `tool_call`
- `arguments`
- `payload`

### Algorithm sketch

```text
extractXmlTagged(input, tags):
    candidates = []
    for each allowed tag:
        find opening tags <tag> or <tag attr="...">
        for each opening tag:
            find next matching </tag>
            if found and not overlapping already accepted candidate:
                payload = text between open end and close start
                raw = text from open start to close end
                format = infer from tag name and payload
                emit candidate
            else:
                emit diagnostic or skip depending options
```

### Important constraints

- Phase 1 should not support arbitrary nested same-name tags.
- Tag matching should be case-sensitive by default; add case-insensitive option only if needed.
- Attribute text can be preserved in `Info` or a future `AttributesRaw` field.
- Regex is acceptable for Phase 1 if tests cover attributes, newlines, and non-greedy matching. Do not claim full XML parsing.

---

## 4.3 Raw JSON/YAML recognition

### What it should detect

Sometimes the entire input is structured data, or a large substring is structured data without wrappers:

```json
{"name":"Alice","tags":["demo"]}
```

```yaml
name: Alice
tags:
  - demo
```

### Recommended strategy

Raw recognition should be ordered and conservative:

1. Trim leading/trailing whitespace.
2. If trimmed text starts with `{` or `[`:
   - attempt strict JSON parse
   - if strict parse fails, attempt JSON sanitize and mark confidence lower
3. If text contains YAML indicators and not obvious prose:
   - attempt YAML sanitize/lint
   - require at least one mapping/list indicator
4. Return candidates only if validation succeeds or repair is plausible.

### Heuristic examples

JSON indicators:

- starts with `{` or `[` and ends with matching `}` or `]`
- strict parse succeeds
- sanitize JSON repair reaches `StrictParseClean`

YAML indicators:

- multiple lines with `key: value`
- list markers `- item`
- frontmatter delimiter `---` handled separately
- parse/lint status from sanitize YAML

### Avoid false positives

Do not classify a normal paragraph containing one colon as YAML. Require structural evidence:

- at least two mapping-like lines, or
- one mapping line plus one nested/list line, or
- YAML parser/linter confidence with clean or repairable result.

---

## 4.4 Frontmatter extraction

### What it should detect

Markdown documents frequently start with frontmatter:

```md
---
title: Demo
tags:
  - goja
---

# Body
```

Also consider TOML and JSON variants:

```md
+++
title = "Demo"
+++
```

```md
;;;
{"title":"Demo"}
;;;
```

Phase 1 should support YAML frontmatter with `---`. TOML/JSON frontmatter can be planned but deferred.

### Algorithm sketch

```text
extractFrontmatter(input):
    if input does not start with --- followed by newline:
        return []
    find next line exactly ---
    if not found:
        emit diagnostic or return []
    payload = lines between delimiters
    emit candidate Kind="frontmatter", Format="yaml", Wrapper="frontmatter"
```

### Why separate from raw YAML?

Frontmatter is a wrapper. The payload may be YAML, but the source span includes delimiters. Keeping it separate lets callers preserve the document body and replace only the frontmatter payload later.

---

## 4.5 Other suggested extractors

### LLM response wrapper extraction

Many outputs include prose around a payload:

```text
Here is the JSON you asked for:
{"ok": true}
Let me know if you need anything else.
```

This overlaps with the sanitize JSON rule `leading_or_trailing_prose`. Phase 1 can delegate validation to sanitize rather than writing a separate extractor. A future helper could expose `extract.llmPayload(input)` that returns the candidate substring sanitize identified as the likely payload.

### Balanced delimiter extraction

A helper could locate balanced `{...}` or `[...]` regions inside text and validate them as JSON. This is useful for logs and chat transcripts but more error-prone than fenced/tagged extraction. Defer until fenced/tagged/raw/frontmatter are stable.

### YAML document separator extraction

YAML streams can contain multiple documents separated by `---` and `...`. This is related to frontmatter but belongs in a YAML-specific extractor if needed.

### Key-value block extraction

Some text contains simple blocks like:

```text
name: Alice
age: 30
city: Paris
```

This should be covered by raw YAML recognition if enough lines match. Avoid a separate extractor until false positives are understood.

---

## Part 5: Proposed Go Package Layout

```text
goja-text/pkg/extract/
  types.go             # Candidate, options, validation result
  positions.go         # line index and byte/row/col helpers
  markdown_fences.go   # fenced code block scanner
  xml_tags.go          # XML-like tag wrapper extractor
  raw.go               # raw JSON/YAML recognition
  frontmatter.go       # frontmatter extraction
  validate.go          # candidate validation using sanitize packages
  module.go            # NativeModule exports
  typescript.go        # RawDTS declarations
  *_test.go            # unit tests
  module_test.go       # goja runtime integration tests
```

Keep domain extraction code separate from `module.go`. The module adapter should only wire exports and convert arguments. Extraction logic should be testable without goja.

---

## Part 6: Native Module Design

### NativeModule

```go
type module struct{}

func (module) Name() string { return "extract" }
func (module) Doc() string  { return "..." }
func (module) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)

    modules.SetExport(exports, "extract", "markdownCodeBlocks", func(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) { ... })
    modules.SetExport(exports, "extract", "xmlTagged", func(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) { ... })
    modules.SetExport(exports, "extract", "rawStructured", func(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) { ... })
    modules.SetExport(exports, "extract", "frontmatter", func(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) { ... })
    modules.SetExport(exports, "extract", "all", func(input string, options *ExtractOptions) ([]*ExtractionCandidate, error) { ... })
    modules.SetExport(exports, "extract", "validate", func(candidate *ExtractionCandidate, options *ValidationOptions) (*CandidateValidationResult, error) { ... })
}
```

### Options builder

Follow the GOJA-TEXT-002 pattern. Use Go-backed builders for options rather than raw JavaScript option maps:

```js
const options = extract.options()
  .Formats("json", "yaml")
  .Tags("json", "yaml", "data")
  .IncludeDiagnostics(true)
  .Build();

const candidates = extract.all(text, options);
```

Suggested builder methods:

- `Formats(...formats)`
- `Tags(...tags)`
- `IncludeDiagnostics(enabled)`
- `InferFormat(enabled)`
- `MinConfidence(n)`
- `MaxCandidates(n)`
- `Build()`
- `Validate()`

---

## Part 7: Decision Records

### Decision 1: Create a new `extract` module instead of adding helpers to `markdown` or `sanitize`

- **Context:** Extraction overlaps with Markdown and sanitize behavior but is not identical to either. Markdown parses documents; sanitize repairs structured formats; extraction locates candidate payloads in messy text.
- **Options considered:** Add codeblock extraction to `markdown`, add JSON/YAML recognition to `sanitize`, or create a new `extract` module.
- **Decision:** Create a new `extract` module.
- **Rationale:** Extraction is a separate layer. It composes Markdown-style scanning and sanitize validation without bloating either existing module.
- **Consequences:** xgoja provider and `xgoja.yaml` need a third goja-text module entry.
- **Status:** proposed

### Decision 2: Return candidates, not parsed values

- **Context:** Callers need source spans and wrapper metadata, not only parsed data.
- **Options considered:** Return parsed JS values, return raw strings, or return Go-backed candidate objects.
- **Decision:** Return Go-backed `ExtractionCandidate` objects.
- **Rationale:** Candidates preserve provenance and can be validated or parsed later. This avoids hiding extraction uncertainty.
- **Consequences:** JavaScript callers use PascalCase fields and call validation/parsing explicitly.
- **Status:** proposed

### Decision 3: Implement Markdown fence extraction with a dedicated scanner

- **Context:** Goldmark finds fenced code blocks, but the current Markdown AST does not expose exact raw wrapper spans.
- **Options considered:** Extend `MarkdownNode`, use goldmark internals, or implement a small scanner.
- **Decision:** Implement a dedicated scanner for fenced code blocks.
- **Rationale:** Extraction needs exact byte spans, raw wrapper text, and unterminated-block diagnostics. A scanner is simpler and keeps Markdown AST stable.
- **Consequences:** Tests must cover CommonMark fence rules enough for intended use.
- **Status:** proposed

### Decision 4: XML-tag extraction is XML-like wrapper extraction, not full XML parsing

- **Context:** LLM outputs often use simple tags, but full XML parsing has different rules and failure modes.
- **Options considered:** Use an XML parser, use regex-like wrapper extraction, or defer XML tags.
- **Decision:** Implement XML-like same-name tag wrapper extraction in Phase 1.
- **Rationale:** The goal is payload recovery, not XML document validation.
- **Consequences:** Document limitations clearly. Do not claim full XML support.
- **Status:** proposed

### Decision 5: Use sanitize for validation, not extraction

- **Context:** The sanitize module can repair JSON/YAML, including LLM wrappers, but extraction needs source spans and candidate provenance.
- **Options considered:** Let sanitize do all recognition, duplicate sanitize rules, or use sanitize only after candidate extraction.
- **Decision:** Use sanitize to validate/repair candidates after extraction.
- **Rationale:** This keeps responsibilities separate and avoids duplicating repair logic.
- **Consequences:** `extract.validate` depends on `github.com/go-go-golems/sanitize v0.0.2` through the existing dependency.
- **Status:** proposed

---

## Part 8: Implementation Plan

### Phase 0: Source-position infrastructure

- Add `pkg/extract/positions.go`.
- Build a line index that maps byte offsets to row/column.
- Add tests for LF, CRLF, beginning/end offsets, and multibyte UTF-8 behavior.

### Phase 1: Candidate types and options

- Add `ExtractionCandidate`, `ExtractOptions`, `ExtractOptionsBuilder`, `CandidateValidationResult`.
- Follow the sanitize builder/config pattern.
- Add TypeScript declarations.

### Phase 2: Markdown codeblock extractor

- Implement `MarkdownCodeBlocks(input, options)`.
- Support backtick and tilde fences.
- Capture language, info string, raw wrapper, payload, byte spans, row/column spans.
- Add unit tests and runtime tests.

### Phase 3: XML-like tag extractor

- Implement `XMLTagged(input, options)`.
- Support caller-provided tag list and default tags.
- Capture tag name, attributes raw text, payload spans, diagnostics for missing close tags if diagnostics are enabled.
- Add tests for multiline payloads, attributes, multiple tags, and non-overlap.

### Phase 4: Raw structured and frontmatter extraction

- Implement raw JSON/YAML recognition with conservative heuristics.
- Implement YAML frontmatter extraction.
- Use sanitize validation to assign format/confidence.
- Add tests for false-positive avoidance.

### Phase 5: `all` and validation

- Implement `All(input, options)` to run selected extractors, merge candidates by source order, and optionally remove overlaps.
- Implement `Validate(candidate, options)` using sanitize packages.
- Add tests for combined documents.

### Phase 6: xgoja integration and demos

- Add `pkg/extract/module.go` and `typescript.go`.
- Update `pkg/xgoja/providers/text/text.go` to blank-import `pkg/extract` and include `extract` in `textModuleNames`.
- Update `xgoja.yaml` to include `extract`.
- Add `examples/js/extract-demo.js`.
- Update README.
- Validate with `make check`.

---

## Part 9: Testing Strategy

### Unit tests

- `positions_test.go` — byte/row/column mapping.
- `markdown_fences_test.go` — fenced code extraction.
- `xml_tags_test.go` — XML-like wrappers.
- `frontmatter_test.go` — frontmatter delimiters.
- `raw_test.go` — JSON/YAML recognition and false positives.
- `validate_test.go` — sanitize-backed validation.

### Runtime tests

- `module_test.go` should verify:
  - `require("extract")` loads.
  - `extract.markdownCodeBlocks` returns Go-backed candidates.
  - `candidate.Kind`, `candidate.Format`, `candidate.Text`, `candidate.StartByte` are visible from JavaScript.
  - `extract.xmlTagged` works for `<json>...</json>`.
  - `extract.rawStructured` recognizes strict JSON and simple YAML.
  - `extract.frontmatter` extracts YAML frontmatter.
  - `extract.validate` marks JSON/YAML candidates valid or repairable.

### xgoja smoke tests

Add demo script:

```js
const extract = require("extract");
const fs = require("fs");

const text = fs.readFileSync("examples/text/structured-data-sample.md", "utf-8");
const candidates = extract.all(text, extract.options().Formats("json", "yaml").Build());
console.log(JSON.stringify(candidates.map(c => ({
  Kind: c.Kind,
  Format: c.Format,
  Label: c.Label,
  Text: c.Text,
})), null, 2));
```

Then add a Makefile smoke target later:

```makefile
smoke-extract: build-xgoja
	$(XGOJA_BINARY) run examples/js/extract-demo.js
```

---

## Part 10: Risks and Open Questions

### Risk: False positives in raw YAML recognition

YAML is permissive. A normal paragraph with colons can look like YAML. Use conservative heuristics and sanitize validation, and prefer wrapper-based candidates when available.

### Risk: XML-like tags are not XML

The module should not claim full XML parsing. It extracts simple same-name wrappers. If full XML parsing becomes necessary, add a separate parser-backed implementation.

### Risk: Overlapping candidates

A Markdown code block may contain XML tags or raw JSON. `extract.all` needs an overlap policy:

- keep all candidates,
- prefer outer wrappers,
- prefer inner structured payloads,
- or return both with parent/child metadata.

Recommended Phase 1 default: keep all candidates sorted by source order, and add `ParentIndex` later if needed.

### Open question: Should extracted code blocks include indentation normalization?

Do not normalize in Phase 1. Return exact payload text. Add normalization options later if needed.

### Open question: Should frontmatter support TOML and JSON in Phase 1?

YAML frontmatter should be Phase 1. TOML and JSON frontmatter can be designed but deferred unless there is a concrete caller.

---

## Part 11: Key Source References

- `goja-text/pkg/markdown/convert.go` — current fenced code block field extraction.
- `goja-text/pkg/markdown/module.go` — native module and TypeScript pattern.
- `goja-text/pkg/sanitize/module.go` — namespace-based native module pattern.
- `goja-text/pkg/sanitize/options.go` — Go-backed builder/config pattern.
- `goja-text/pkg/sanitize/module_test.go` — runtime tests for Go-backed builders/results.
- `goja-text/pkg/xgoja/providers/text/text.go` — provider registration list.
- `goja-text/xgoja.yaml` — generated binary module composition.
- `goja-text/Makefile` — validation target pattern.
- `sanitize/pkg/json/parse.go` — strict JSON validation.
- `sanitize/pkg/yaml/sanitize.go` — YAML validation/repair loop.

---

## Part 12: Implementation Checklist

- [ ] Create `pkg/extract` package skeleton.
- [ ] Add line-index/source-position helpers and tests.
- [ ] Add `ExtractionCandidate`, validation result, and options builder types.
- [ ] Implement Markdown fenced code block extraction.
- [ ] Implement XML-like tag wrapper extraction.
- [ ] Implement YAML frontmatter extraction.
- [ ] Implement raw JSON/YAML recognition.
- [ ] Implement sanitize-backed candidate validation.
- [ ] Implement `extract` NativeModule and TypeScript declarations.
- [ ] Add runtime tests for JavaScript API.
- [ ] Wire `extract` into xgoja provider and `xgoja.yaml`.
- [ ] Add demo fixtures/scripts and Makefile smoke target.
- [ ] Update README.
- [ ] Run `make check`.
- [ ] Upload updated docs to reMarkable.
