---
Title: Source-Preserving Chunking Architecture and Implementation Guide
Ticket: GOJA-TEXT-006
Status: active
Topics:
    - goja
    - goja-bindings
    - markdown
    - native-modules
    - text-algorithms
    - xgoja
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://cmd/goja-text/jsverbs/chunking.js
      Note: Generated app commands
    - Path: repo://cmd/goja-text/xgoja.yaml
      Note: Generated application module, jsverb, asset, and help composition
    - Path: repo://examples/js/chunking-demo.js
      Note: Runnable exploration entry point
    - Path: repo://pkg/chunking/module.go
      Note: JavaScript adapter and TypeScript contract
    - Path: repo://pkg/chunking/module_test.go
      Note: Goja runtime integration coverage
    - Path: repo://pkg/markdown/convert.go
      Note: Goldmark conversion and source-range extension point
    - Path: repo://pkg/markdown/module.go
      Note: Native module, TypeScript declaration, and runtime integration pattern
    - Path: repo://pkg/xgoja/providers/text/doc/chunking-api-reference.md
      Note: Runtime API reference
    - Path: repo://pkg/xgoja/providers/text/text.go
      Note: xgoja provider registration and help packaging
ExternalSources:
    - https://github.com/go-go-golems/goja-text/issues/9
Summary: Intern-facing architecture, API, algorithm, implementation, and validation guide for source-preserving text chunking in goja-text.
LastUpdated: 2026-07-10T13:09:36.156617008-04:00
WhatFor: Implement and maintain exact text spans, structure-aware segmenters, budgeted packing, recursive fallback, and the require("chunking") JavaScript module.
WhenToUse: Read before changing Markdown source coordinates, chunking algorithms, JavaScript contracts, TypeScript declarations, xgoja provider registration, or chunking tests.
---







# Source-Preserving Chunking Architecture and Implementation Guide

## Executive Summary

This ticket adds a `chunking` native module to `goja-text`. The module turns UTF-8 source text into exact, inspectable spans and then packs those spans into bounded chunks. It supports plain-text boundaries and Markdown structure without adding transcript, embedding, vector-search, or retrieval-specific behavior.

The central correctness property is **source preservation**. A built-in segmenter must partition the original source byte for byte. A packer may repeat complete spans only through an explicit overlap policy; it must never silently drop source text, invent citation ranges, or reconstruct Markdown from plain text.

The implementation has four layers:

1. Extend Markdown AST nodes with exact byte/rune/end coordinates.
2. Add pure-Go segmenting and packing algorithms in `pkg/chunking`.
3. Expose those algorithms through `require("chunking")` with TypeScript declarations.
4. Package the module, examples, jsverbs, and help pages into the generated xgoja binary.

The work implements [GitHub issue #9](https://github.com/go-go-golems/goja-text/issues/9). This document is self-contained: a new engineer should be able to implement, review, or extend the feature without reconstructing its design from the issue discussion.

## Problem Statement

### Why fixed windows are insufficient

A fixed rune window is deterministic and Unicode-safe, but it ignores document structure. It may split:

- a heading from the paragraph it introduces;
- a list halfway through an item;
- a fenced code block between its opening fence and body;
- a block quote in the middle of a statement;
- a Markdown section immediately before a decisive sentence.

Downstream systems can still embed such chunks, but citations and retrieval quality become harder to reason about. The text library already parses Markdown through Goldmark, so it should expose reusable structural boundaries rather than forcing each application to reverse-engineer the AST.

### Why source coordinates are part of the API

Chunking is not only a text-generation operation. Consumers need to relate a chunk back to the original document. That requires two coordinate systems:

- byte offsets for exact UTF-8 slices, storage, and transport;
- rune offsets for Unicode-aware character budgets and user-facing lengths.

JavaScript string indices are UTF-16 code units, so they cannot substitute for either coordinate system. Offsets must be computed in Go against the original UTF-8 source and returned as explicit fields.

### Scope

In scope:

- exact Markdown AST source ranges;
- line, paragraph, Markdown-block, and Markdown-section segmenters;
- byte, rune, word, and caller-provided weight budgets;
- whole-span overlap;
- explicit oversized-span policy;
- recursive fallback splitting;
- strategy descriptions and diagnostics;
- JavaScript, TypeScript, provider, generated-host, examples, and help integration.

Out of scope:

- transcript turn grouping;
- content hashes and application chunk IDs;
- embedding and tokenizer providers;
- semantic-breakpoint selection;
- vector indexes and retrieval fusion;
- generation manifests and evaluation metrics.

Those remain downstream responsibilities.

## Baseline Architecture Before This Ticket

### Native modules

`goja-text` currently registers four Go-backed modules: `markdown`, `sanitize`, `extract`, and `template`. Each package contains domain logic and a small `module.go` adapter implementing:

```go
type NativeModule interface {
    Name() string
    Doc() string
    Loader(*goja.Runtime, *goja.Object)
}
```

Module packages register themselves through `modules.Register` during `init()`. The xgoja provider blank-imports every package, looks up the registered module by name, forwards its loader, and forwards its TypeScript descriptor.

```mermaid
flowchart LR
    Domain["Pure Go operations"] --> Module["module.go<br/>exports + codecs + DTS"]
    Module --> Registry["modules.Register"]
    Registry --> Provider["goja-text xgoja provider"]
    Provider --> Host["generated goja-text binary"]
    Host --> JS["require(moduleName)"]

    style Domain fill:#dcfce7,stroke:#16a34a
    style Module fill:#dbeafe,stroke:#2563eb
    style Host fill:#fef3c7,stroke:#d97706
```

Evidence:

- `pkg/markdown/module.go` implements both `NativeModule` and `TypeScriptDeclarer`.
- `pkg/xgoja/providers/text/text.go` enumerates provider module names and imports all native packages.
- `pkg/xgoja/providers/text/text_test.go` verifies that every provider module has a TypeScript descriptor.
- `cmd/goja-text/xgoja.yaml` selects the provider modules, embedded jsverbs, assets, and help source.

### Markdown representation

`markdown.parse(input)` returns a Go-backed `MarkdownNode`. JavaScript intentionally reads PascalCase Go fields:

```javascript
const markdown = require("markdown");
const ast = markdown.parse("# Title\n\nBody");

console.log(ast.Type);
console.log(ast.Children[0].Level);
```

Before this ticket, the node included `StartLine`, `StartColumn`, and `SourcePos`, but not exact byte/rune end ranges. `ConvertAST` derived content fields through a type switch and recursively converted child nodes. The completed implementation adds the exact fields without changing the existing start-position contract.

Goldmark source storage differs by node:

- inline text exposes a direct segment;
- block nodes expose one or more line segments;
- container nodes may need child aggregation;
- fenced code separates fence syntax from body-line values;
- HTML blocks may include a closure line;
- structural nodes such as thematic breaks may have no child text but still occupy source bytes.

The completed implementation centralizes these cases in `pkg/markdown/source_ranges.go`.

### Generated application

The committed generated application under `cmd/goja-text` is part of the product:

```text
cmd/goja-text/
├── xgoja.yaml
├── generate.go
├── main.go
├── jsverbs/
├── xgoja_embed/
├── go.mod
└── go.sum
```

Changes to provider modules or jsverbs require regeneration and a build. Runtime help is embedded from `pkg/xgoja/providers/text/doc/*.md`.

## Baseline Gap Analysis

| Required capability | Current support | Gap |
| --- | --- | --- |
| Exact Markdown start coordinate | Start line/column | No start byte/rune fields |
| Exact Markdown end coordinate | None | No end byte/rune/line/column |
| Line segmentation | Application code only | No native lossless primitive |
| Paragraph segmentation | None | Separator ownership undefined |
| Markdown block segmentation | AST can be walked | Callers must reconstruct ranges |
| Markdown section segmentation | Headings are exposed | No heading-path or section ranges |
| Budgeted packing | None | No shared budget/overlap invariants |
| Token-aware packing | None | Must avoid hard tokenizer dependency |
| Recursive fallback | None | Oversized structures have no policy |
| Strategy identity | None | Downstream plans cannot describe behavior |
| Diagnostics | Module errors only | No per-result warning/error evidence |
| Generated host | Four modules | `chunking` not registered or documented |

## Implemented Solution

### Coordinate semantics

All byte and rune offsets are zero-based and half-open:

```text
[StartByte, EndByte)
[StartRune, EndRune)
```

Line and column coordinates are one-based to remain consistent with the current Markdown API. End coordinates point immediately after the range.

Chunking entry points reject invalid UTF-8. Existing `markdown.parse` behavior is not changed.

### Public domain types

```go
type Span struct {
    Ordinal int
    Kind    string
    Text    string

    StartByte int
    EndByte   int
    StartRune int
    EndRune   int

    StartLine   int
    StartColumn int
    EndLine     int
    EndColumn   int

    Atomic       bool
    HeadingLevel int
    HeadingPath  []string
    Language     string
}

type Diagnostic struct {
    Code      string
    Severity  string
    Message   string
    StartByte int
    EndByte   int
}

type StrategySpec struct {
    Name    string
    Version string
    Options map[string]any
}

type SegmentResult struct {
    Spec        StrategySpec
    SourceBytes int
    SourceRunes int
    Spans       []Span
    Diagnostics []Diagnostic
}

type PackedChunk struct {
    Ordinal     int
    Text        string
    StartByte   int
    EndByte     int
    StartRune   int
    EndRune     int
    SpanOrdinals []int
    HeadingPath []string
    Weight      int
    Oversized   bool
    Diagnostics []Diagnostic
}

type PackResult struct {
    Spec        StrategySpec
    Chunks      []PackedChunk
    Diagnostics []Diagnostic
}
```

Results remain Go-backed values with PascalCase fields. This matches existing Markdown nodes, extraction candidates, sanitize results, and builders.

### JavaScript functions

```javascript
const chunking = require("chunking");

const lines = chunking.lines(source, {
  keepTerminators: true,
});

const paragraphs = chunking.paragraphs(source, {
  blankLines: "trailing",
});

const blocks = chunking.markdownBlocks(source, {
  atomic: ["fencedCodeBlock", "codeBlock", "htmlBlock"],
});

const sections = chunking.markdownSections(source, {
  maxHeadingLevel: 4,
});

const packed = chunking.pack(blocks.Spans, {
  maxUnits: 2400,
  measure: "runes",
  overlap: { unit: "spans", value: 1 },
  oversized: "allow",
});

const weighted = chunking.packWeighted(
  blocks.Spans.map(span => ({
    span,
    weight: tokenizer.count(span.Text),
  })),
  {
    maxWeight: 512,
    overlapWeight: 64,
    oversized: "allow",
  },
);

const recursive = chunking.recursive(source, {
  maxUnits: 2400,
  measure: "runes",
  levels: [
    "markdownSections",
    "markdownBlocks",
    "paragraphs",
    "lines",
    "runes",
  ],
  overlap: { unit: "spans", value: 1 },
});
```

Module functions use lower camel case. Option objects use lower camel case. Unknown option keys are rejected rather than silently ignored.

### Source-preserving segmenters

Each built-in segmenter must satisfy:

```text
join(result.Spans[i].Text for i in ordinal order) == original source
```

Line spans own their terminator. Paragraph spans own trailing blank-line separators. Markdown block spans use consecutive top-level node starts as boundaries: the first begins at byte zero, each ends at the next block start, and the final ends at source length. This assigns leading and inter-block whitespace deterministically without rendering or reconstructing Markdown.

Markdown sections form a flat, non-overlapping partition. Every heading at or above the configured `maxHeadingLevel` starts a new section, and that section ends at the next accepted heading or at end of input. Content before the first accepted heading becomes a preamble span. Hierarchy is represented by `HeadingPath` metadata rather than overlapping parent ranges, so concatenating section text still reproduces the source exactly.

### Packing algorithm

The non-recursive packer greedily groups complete spans:

```text
current = []

for span in source order:
    weight = measure(span.Text)

    if current is empty and weight > budget:
        handle according to oversized policy
        continue

    if weight(current) + weight <= budget:
        append span
        continue

    emit current
    current = declared trailing overlap spans

    while overlap prevents adding span:
        remove oldest overlap span

    append span

emit current if non-empty
```

The overlap is expressed as complete spans. It may repeat source content across chunks, but it cannot split a span or prevent forward progress.

### Weighted packing

`packWeighted` accepts caller-supplied nonnegative integer weights. The module does not interpret or recompute them. This supports model-specific token counts without importing a tokenizer or Geppetto into `goja-text`.

### Recursive fallback

Recursive splitting applies an ordered set of segmenters only to oversized ranges:

```text
split range at level N
pack spans that fit

for each oversized span:
    if another level exists:
        split that exact absolute source range at level N+1
    else:
        apply fixed rune windows
```

Nested results retain absolute coordinates. The implementation must guarantee progress and record which fallback level produced each chunk.

### Diagnostics

Stable initial codes:

- `invalid_utf8`
- `invalid_range`
- `source_range_mismatch`
- `empty_chunk`
- `atomic_span_exceeds_budget`
- `span_exceeds_budget`
- `invalid_weight`
- `unknown_measure`
- `unknown_recursive_level`
- `overlap_prevents_progress`

Oversized content allowed by policy is a warning. Invalid ranges, unknown modes, and non-progress states are errors.

## Design Decisions

### Decision: create a separate `chunking` module

- **Context:** Markdown parsing and chunking share source structure but are different operations.
- **Options considered:** add chunk functions to `markdown`; create `chunking`; implement only downstream JavaScript helpers.
- **Decision:** create `require("chunking")`.
- **Rationale:** plain-text segmenters and packers do not require Markdown, while the module can reuse the Markdown package internally.
- **Consequences:** one provider module and help pair are added; existing Markdown callers remain unaffected.
- **Status:** accepted

### Decision: return Go-backed domain values

- **Context:** Existing modules expose validated Go-backed nodes, results, candidates, and builders.
- **Options considered:** plain lower-camel JavaScript objects; Go-backed structs with PascalCase fields.
- **Decision:** expose Go-backed result structs and lower-camel function/options APIs.
- **Rationale:** Go validates offsets and invariants, TypeScript can describe the fields, and the shape matches repository conventions.
- **Consequences:** examples must explicitly teach PascalCase result fields.
- **Status:** accepted

### Decision: segmentation is lossless by default

- **Context:** A structure-only AST traversal commonly omits separators and fence syntax.
- **Options considered:** return semantic text only; return source blocks but allow gaps; require an exact partition.
- **Decision:** every built-in segmenter partitions the source exactly.
- **Rationale:** exact citations and downstream validation require a reversible mapping.
- **Consequences:** whitespace ownership and empty/whitespace-only input require explicit tests.
- **Status:** accepted

### Decision: no tokenizer dependency

- **Context:** Token counts are model-specific and tokenizer dependencies can be large.
- **Options considered:** ship one tokenizer; callbacks from Go into JavaScript; caller-provided weights.
- **Decision:** support bytes/runes/words natively and preweighted spans for external tokenizers.
- **Rationale:** the packing invariant is independent of how weight was computed.
- **Consequences:** callers must retain the same weights when reproducing a token-budget experiment.
- **Status:** accepted

### Decision: derived context is metadata

- **Context:** Heading breadcrumbs can improve embeddings but are not contiguous source text.
- **Options considered:** prepend headings to `Text`; duplicate heading spans; return heading paths separately.
- **Decision:** return `HeadingPath` metadata and leave embedding-text composition to consumers.
- **Rationale:** source coordinates remain truthful.
- **Consequences:** downstream code needs an explicit embedding-text policy in its own plan.
- **Status:** accepted

## Alternatives Considered

- **JavaScript-only splitting:** fast to prototype, but UTF-16 indices do not provide trustworthy UTF-8/rune coordinates and each consumer would duplicate invariants.
- **Markdown-only API:** insufficient for plain text, tool output, and callers that already have structural spans.
- **Returning `markdown.textContent`:** removes syntax and whitespace and cannot support exact source slices.
- **Arbitrary regular-expression separator lists:** flexible but hard to type, validate, document, and guarantee progress in the first stable API.
- **Semantic chunking in this module:** requires embeddings and belongs in a higher-level experimental system.
- **Backwards-compatibility aliases:** unnecessary because `chunking` is a new module.

## End-to-End Runtime Flow

The completed runtime keeps structural decisions in pure Go and uses JavaScript as the composition layer. A JavaScript caller can choose a segmenter, inspect exact spans, obtain deterministic or model-specific weights, and choose a packing policy without changing the native module.

```mermaid
sequenceDiagram
    participant JS as JavaScript experiment
    participant Codec as module.go codecs
    participant Segment as Go segmenter
    participant Pack as Go packer
    participant App as embedding/index application

    JS->>Codec: markdownBlocks(source, options)
    Codec->>Codec: reject unknown/mistyped options
    Codec->>Segment: MarkdownBlocks(source, typed options)
    Segment->>Segment: parse AST and partition source
    Segment->>Segment: ValidatePartition(source, spans)
    Segment-->>JS: Go-backed SegmentResult
    JS->>JS: inspect spans or compute tokenizer weights
    JS->>Codec: pack(spans, options) or packWeighted(items, options)
    Codec->>Pack: validated spans and budget
    Pack->>Pack: greedy complete-span packing
    Pack-->>JS: Go-backed PackResult
    JS->>App: text, exact ranges, strategy, diagnostics
```

The source itself is never rendered or normalized during this flow. Markdown parsing supplies boundaries and metadata only. `Span.Text` always comes from a byte slice of the original source.

## Implemented File-by-File Guide

This section is the recommended reading order for a new intern. Each layer has one responsibility, and the tests beside that layer state its externally visible contract.

### Exact Markdown coordinates

Files:

- `pkg/markdown/types.go` defines the JavaScript-visible `MarkdownNode` fields.
- `pkg/markdown/convert.go` converts Goldmark nodes and computes byte, rune, line, and column coordinates.
- `pkg/markdown/source_ranges.go` derives syntax-preserving node envelopes from direct segments, block starts, siblings, and ancestors.
- `pkg/markdown/source_ranges_test.go` verifies headings, emphasis, links, lists, fences, block quotes, HTML, thematic breaks, and Unicode.

Goldmark exposes different source information for leaf and structural nodes. Text and raw HTML leaves have direct segments. Structural nodes use their own start and the next structural sibling boundary, walking through ancestors when the node has no direct next sibling. Trailing inter-block whitespace is removed from AST envelopes; the chunking block segmenter later assigns separators losslessly.

### Domain types and position calculation

Files:

- `pkg/chunking/types.go` defines spans, diagnostics, strategy descriptions, options, and results.
- `pkg/chunking/positions.go` creates spans and translates substring-relative recursive ranges to document-absolute ranges.
- `pkg/chunking/validate.go` validates UTF-8, exact partitions, and packable span sequences.

The coordinate code is deliberately centralized. A new segmenter should identify only byte boundaries, then call the source index to populate all coordinate systems. It should not independently count lines or runes.

Pseudocode for a new segmenter:

```text
function segment(source, options):
    index = validateUTF8AndCreateIndex(source)
    boundaries = detectBoundaries(source, options)
    spans = []

    for each consecutive [start, end) in boundaries:
        spans.append(index.span(start, end, ordinal, kind))

    validatePartition(source, spans)
    return SegmentResult(strategy, source sizes, spans)
```

### Segmenters

Files:

- `pkg/chunking/segment_lines.go` handles LF and CRLF and can expose terminators separately.
- `pkg/chunking/segment_paragraphs.go` detects blank-line runs and implements explicit separator ownership.
- `pkg/chunking/segment_markdown.go` implements top-level Markdown blocks, flat sections, heading paths, language metadata, and atomic kinds.
- `pkg/chunking/segment_test.go` verifies preservation, Unicode, CRLF, structural metadata, invalid UTF-8, and fuzz seeds.

Line terminators and paragraph separators are never dropped. When `keepTerminators` is false, a CRLF terminator becomes one `lineTerminator` span containing both bytes. Paragraph mode changes ownership, not content.

Markdown blocks partition at top-level node starts. The first span starts at byte zero even when the first node begins after leading whitespace. The final span ends at source length. This rule assigns every leading, inter-block, and trailing byte deterministically.

### Packing and recursive fallback

Files:

- `pkg/chunking/pack.go` implements measurement, span-sequence validation, greedy packing, overlap, weighted packing, and oversized diagnostics.
- `pkg/chunking/recursive.go` refines oversized ranges through ordered segmenters and translates absolute coordinates.
- `pkg/chunking/pack_test.go` covers budgets, overlap, oversized policy, forged ranges, caller weights, recursion, and fuzz seeds.

The packer validates that each supplied span has a contiguous ordinal, that byte and rune lengths agree with `Text`, and that ranges are consecutive. This prevents a JavaScript caller from forging a plausible Go-shaped object with inconsistent citations.

Oversized handling occurs before ordinary fit logic on every iteration. This ordering matters when an oversized span follows a full chunk: the previous chunk is emitted, then the oversized span is emitted alone with `Oversized: true` and `span_exceeds_budget`. A generated-host smoke test found this edge case, and `TestPackMarksOversizedSpanAfterFlushingCurrentChunk` prevents regression.

### JavaScript adapter and TypeScript

Files:

- `pkg/chunking/module.go` implements `require("chunking")`, strict codecs, module documentation, and the TypeScript descriptor.
- `pkg/chunking/module_test.go` runs the module inside a real go-go-goja runtime.
- `pkg/xgoja/providers/text/text.go` publishes the module through the `goja-text` provider.
- `pkg/xgoja/providers/text/text_test.go` resolves every provider module and validates its TypeScript descriptor.

JavaScript option objects use lower-camel keys because they are script configuration. Go-backed returned values expose PascalCase fields because Goja projects exported Go fields and because existing goja-text modules use the same convention.

The adapter treats `nil`, JavaScript `undefined`, and `null` as absent optional values. It rejects unknown keys, numeric strings, non-integral numeric values, and arrays with the wrong element type. Errors become JavaScript exceptions through `vm.NewGoError` or `vm.NewTypeError`.

The pinned go-go-goja TypeScript API represents optional parameters with `spec.Param.Optional`, not an optional type wrapper. The module therefore declares:

```go
spec.Param{
    Name: "options",
    Type: spec.Named("LineOptions"),
    Optional: true,
}
```

The generated declaration is `options?: LineOptions`. A broader additive TypeScript-spec enhancement is tracked in go-go-goja issue #92, but this implementation has no dependency on it.

### Generated application and documentation

Files:

- `cmd/goja-text/xgoja.yaml` selects the module, provider help, assets, and root-mounted jsverbs.
- `cmd/goja-text/jsverbs/chunking.js` supplies `chunking blocks`, `chunking pack`, and `chunking recursive`.
- `examples/js/chunking-demo.js` is the copyable exploration starting point.
- `examples/markdown/chunking-sample.md` exercises headings, paragraphs, a fence, a list, and multiple sections.
- `pkg/xgoja/providers/text/doc/chunking-api-reference.md` is the exhaustive runtime reference.
- `pkg/xgoja/providers/text/doc/chunking-user-guide.md` teaches the intended exploration and tokenizer-integration workflow.
- `Makefile` builds, regenerates, and smokes the product surfaces.

`make build-xgoja` regenerates committed host artifacts and builds `dist/goja-text`. `goja-text types` prints the selected module declarations, and both help slugs are queryable through the generated Glazed help system.

## Historical Implementation Sequence

### Phase 1: exact Markdown coordinates — complete

Files:

- `pkg/markdown/types.go`
- `pkg/markdown/convert.go`
- `pkg/markdown/module.go`
- new focused range tests

Implement one internal source-range helper, add byte/rune/end fields, update TypeScript, and verify JS access.

### Phase 2: core chunking package and segmenters — complete

Files:

- `pkg/chunking/types.go`
- `positions.go`
- `segment_lines.go`
- `segment_paragraphs.go`
- `segment_markdown.go`
- `validate.go`

Implement coordinate indexing once and reuse it. Add lossless validators before Markdown segmenters so every later algorithm can assert its output.

### Phase 3: packing — complete

Files:

- `pkg/chunking/pack.go`
- `recursive.go`
- focused table, golden, and fuzz tests

Implement simple packing first, then preweighted packing, then recursion. Do not combine all three before the basic coverage/progress tests pass.

### Phase 4: JavaScript module and provider — complete

Files:

- `pkg/chunking/module.go`
- `pkg/chunking/typescript.go` or declarations in `module.go`
- `pkg/xgoja/providers/text/text.go`
- `pkg/xgoja/providers/text/text_test.go`

The loader must only decode options and wire exports. Domain algorithms remain callable from pure Go tests.

### Phase 5: product surfaces — complete

Files:

- `examples/js/chunking-demo.js`
- `examples/markdown/chunking-sample.md`
- `cmd/goja-text/jsverbs/chunking.js`
- provider help pages
- `README.md`
- `Makefile` smoke target
- regenerated `cmd/goja-text` artifacts

### Phase 6: validation — complete

```bash
go fmt ./...
go test ./... -count=1
GOWORK=off go test ./... -count=1
go test ./pkg/chunking -run=Fuzz -fuzztime=short
make build-xgoja
./dist/goja-text run examples/js/chunking-demo.js
./dist/goja-text help goja-text-chunking-user-guide
make check
make lint
```

The exact fuzz command may be split per target because Go runs one fuzz target at a time.

## Testing and Validation Strategy

### Core invariants

For segmentation:

- ordinals are contiguous;
- ranges are monotonic, non-overlapping, and gapless;
- every byte boundary is valid UTF-8;
- `Span.Text` equals the source slice;
- joined span text equals the source;
- byte/rune/line coordinates agree;
- repeated calls produce identical results.

For packing:

- chunks contain at least one span;
- chunks remain in source order;
- no span disappears;
- duplication occurs only through overlap;
- budget is respected unless `Oversized` is true;
- overlap cannot prevent progress;
- repeated calls produce identical results.

### Fixtures

Tests cover:

- empty and whitespace-only input;
- LF and CRLF;
- Unicode, emoji, and combining marks;
- paragraphs and repeated blank lines;
- headings at levels 1–6;
- nested lists and block quotes;
- fenced and indented code;
- HTML blocks and thematic breaks;
- malformed but parseable Markdown;
- preamble text;
- exact-budget and oversized spans.

### Test levels

```mermaid
flowchart TD
    Pure["Pure Go unit + fuzz tests"] --> Runtime["engine runtime require(chunking) test"]
    Runtime --> Provider["provider descriptor test"]
    Provider --> Generated["generated xgoja binary"]
    Generated --> Smoke["JS demo + jsverb + help smoke"]

    style Pure fill:#dcfce7,stroke:#16a34a
    style Runtime fill:#dbeafe,stroke:#2563eb
    style Generated fill:#fef3c7,stroke:#d97706
```

## Known Limitations and Future Work

- Goldmark extension nodes that are not enabled by the current parser need new source-range fixtures when an extension is introduced.
- Word measurement uses `strings.Fields`, which defines words by Unicode whitespace. It is deterministic but is not linguistic segmentation or model tokenization.
- `PackedChunk.Level` records one compact fallback level. The implementation intentionally does not return a deeply nested recursive trace.
- The coordinate implementation favors correctness and recomputes some rune prefixes. Large-document benchmarks should determine whether a prefix index is needed before adding complexity.
- The final `runes` fallback guarantees progress for byte and rune measurement. Word-budget recursion can still surface one oversized whitespace-free range and should use caller weights when exact model limits matter.
- Sentence segmentation is not included. A future release should evaluate Unicode sentence-boundary behavior and language requirements separately.

## References

- [GitHub issue #9](https://github.com/go-go-golems/goja-text/issues/9)
- [go-go-goja issue #92](https://github.com/go-go-golems/go-go-goja/issues/92) — proposed additive TypeScript declaration builders and structured type nodes; not required by this implementation.
- `pkg/markdown/module.go` — native module, TypeScript, and runtime export conventions.
- `pkg/markdown/types.go` — current Go-backed AST contract.
- `pkg/markdown/convert.go` — Goldmark conversion and current source-position logic.
- `pkg/markdown/module_test.go` — runtime `require("markdown")` test pattern.
- `pkg/extract/types.go` — source-coordinate and result-struct precedent.
- `pkg/sanitize/options.go` — strict lower-camel plain-object option decoding precedent.
- `pkg/xgoja/providers/text/text.go` — provider packaging.
- `pkg/xgoja/providers/text/text_test.go` — provider TypeScript validation.
- `pkg/xgoja/providers/text/doc/markdown-api-reference.md` — help-page style and current Markdown API.
- `cmd/goja-text/xgoja.yaml` — generated product composition.
- `cmd/goja-text/jsverbs/markdown.js` — bundled jsverb style.
- `Makefile` — repository validation and smoke targets.
