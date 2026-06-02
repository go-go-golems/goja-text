---
Title: "goja-text extract JavaScript API reference"
Slug: goja-text-extract-api-reference
Short: "Reference for require(\"extract\") structured-data candidate helpers."
Topics:
- goja-text
- extract
- structured-data
- javascript
Commands:
- goja-text
- goja-text eval
- goja-text run
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

Use `require("extract")` to find JSON, YAML, and similar structured-data candidates embedded in larger text.

The module returns candidates, not fully trusted parsed values. A candidate preserves source spans, wrapper metadata, format guesses, confidence, and diagnostics so downstream code can decide what to parse or repair.

## Loading

```js
const extract = require("extract");
```

## Functions

### options()

Returns a Go-backed extraction options builder. Use this to control accepted formats, tags, confidence thresholds, and validation behavior as the API grows.

```js
const options = extract.options().Build();
```

### markdownCodeBlocks(input, options?)

Finds Markdown fenced code blocks that look like structured data.

```js
const candidates = extract.markdownCodeBlocks("~~~json\n{\"ok\": true}\n~~~");
```

### xmlTagged(input, options?)

Finds simple XML-like same-name wrappers such as `<json>...</json>` and `<yaml>...</yaml>`.

This is intentionally XML-like extraction, not a full XML parser.

### frontmatter(input, options?)

Finds YAML frontmatter at the start of a document.

### rawStructured(input, options?)

Finds whole-string raw JSON or YAML candidates.

### all(input, options?)

Runs all extractors and returns candidates sorted by source order.

Overlapping candidates are preserved in the current API so scripts can inspect the evidence and choose a policy.

### validate(candidate, options?)

Validates and sanitizes one candidate by delegating JSON/YAML repair to the sanitize semantics.

```js
const candidate = extract.all(input)[0];
const result = extract.validate(candidate);
console.log(result.Valid, result.Sanitized);
```

## Candidate fields

Common `ExtractionCandidate` fields include:

- `Kind` — wrapper kind such as `markdownCodeBlock`, `xmlTagged`, `frontmatter`, or `raw`.
- `Format` — guessed data format, commonly `json` or `yaml`.
- `Text` — payload text intended for parsing.
- `Raw` — full raw wrapper text.
- `StartByte`, `EndByte`, `StartRow`, `StartColumn`, `EndRow`, `EndColumn` — source span for the raw candidate.
- `PayloadStartByte`, `PayloadEndByte`, `PayloadStartRow`, `PayloadStartColumn`, `PayloadEndRow`, `PayloadEndColumn` — source span for the payload.
- `Fence`, `TagName`, `Info`, `Language` — wrapper metadata when available.
- `Confidence` — heuristic confidence score.
- `Diagnostics` — extraction-time notes.

## Validation result fields

Validation results expose PascalCase fields such as `Valid`, `Format`, `Sanitized`, `Issues`, and `Fixes`.
