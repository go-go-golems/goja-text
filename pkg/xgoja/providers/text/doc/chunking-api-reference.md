---
Title: "goja-text chunking JavaScript API reference"
Slug: goja-text-chunking-api-reference
Short: "Reference for source-preserving segmentation and budgeted packing through require(\"chunking\")."
Topics:
- goja-text
- chunking
- markdown
- javascript
- xgoja
Commands:
- goja-text eval
- goja-text run
- goja-text chunking
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

The `chunking` module separates boundary detection from budgeted packing. Segmenters produce exact source spans; packers combine complete spans without reconstructing the input. This separation lets a script compare strategies, attach model-specific weights, and retain citation coordinates in the original UTF-8 document.

Load the module with:

```js
const chunking = require("chunking");
```

Returned values are Go-backed objects, so JavaScript reads PascalCase fields such as `Spans`, `Text`, `StartByte`, and `Chunks`. Option arguments are plain JavaScript objects with lower-camel keys. Unknown option keys and incorrect primitive types are errors.

## Coordinate and preservation contract

Every built-in segmenter returns a gapless partition of valid UTF-8 input. Concatenating `Span.Text` in ordinal order reproduces the original string byte for byte:

```js
const result = chunking.markdownBlocks(source);
if (result.Spans.map((span) => span.Text).join("") !== source) {
  throw new Error("source partition changed the document");
}
```

Byte and rune intervals are zero-based and half-open. Line and column coordinates are one-based, and end coordinates point immediately after the span.

| Field | Meaning |
| --- | --- |
| `StartByte`, `EndByte` | Exact UTF-8 byte interval `[start, end)` |
| `StartRune`, `EndRune` | Unicode code-point interval `[start, end)` |
| `StartLine`, `StartColumn` | One-based start position |
| `EndLine`, `EndColumn` | One-based position immediately after the span |

JavaScript string indices are UTF-16 code units and are not interchangeable with either byte or rune offsets.

## Segment functions

Segment functions accept source text and return a `SegmentResult`. Invalid UTF-8 is rejected before any boundary calculation.

### lines(source, options?)

`lines` partitions source at LF boundaries and recognizes CRLF as one terminator.

```js
const result = chunking.lines("alpha\r\nbeta\n", {
  keepTerminators: true,
});
```

Options:

| Key | Type | Default | Behavior |
| --- | --- | --- | --- |
| `keepTerminators` | boolean | `true` | Attach each terminator to its line. When false, emit explicit `lineTerminator` spans so source remains lossless. |

### paragraphs(source, options?)

`paragraphs` detects runs of blank lines and assigns their bytes according to one explicit ownership policy.

```js
const result = chunking.paragraphs(source, {
  blankLines: "trailing",
});
```

`blankLines` accepts:

- `trailing` — attach the separator to the preceding paragraph;
- `separate` — emit a `paragraphSeparator` span;
- `leading` — attach the separator to the following paragraph.

The default is `trailing` because it keeps paragraph content and the whitespace that follows it together.

### markdownBlocks(source, options?)

`markdownBlocks` parses the document with Goldmark and partitions it at consecutive top-level node starts. Each span preserves original markers, fences, indentation, whitespace, and line endings.

```js
const result = chunking.markdownBlocks(source, {
  atomic: ["fencedCodeBlock", "codeBlock", "htmlBlock"],
});
```

The `atomic` option marks block kinds that recursive splitting should preserve. Atomic metadata does not change `Text` or ranges.

Common `Kind` values include `heading`, `paragraph`, `list`, `blockquote`, `fencedCodeBlock`, `codeBlock`, `htmlBlock`, and `thematicBreak`.

### markdownSections(source, options?)

`markdownSections` creates a flat, non-overlapping partition beginning at accepted headings. Text before the first accepted heading becomes a `preamble` span. Heading ancestry is stored in `HeadingPath` metadata.

```js
const result = chunking.markdownSections(source, {
  maxHeadingLevel: 4,
});
```

`maxHeadingLevel` defaults to 6 and must be between 1 and 6. A level-two section nested below `# API` may have `HeadingPath` equal to `['API', 'Authentication']`.

## pack(spans, options)

`pack` greedily combines complete spans while the selected measurement stays within `maxUnits`. It never cuts a span.

```js
const blocks = chunking.markdownBlocks(source);
const result = chunking.pack(blocks.Spans, {
  maxUnits: 2400,
  measure: "runes",
  overlap: { unit: "spans", value: 1 },
  oversized: "allow",
});
```

Options:

| Key | Type | Default | Behavior |
| --- | --- | --- | --- |
| `maxUnits` | positive integer | required | Maximum measured size of an ordinary chunk |
| `measure` | `bytes`, `runes`, or `words` | `runes` | Deterministic measurement function |
| `overlap.unit` | `spans` | `spans` | Overlap is always expressed as complete spans |
| `overlap.value` | nonnegative integer | `0` | Trailing spans repeated in the next chunk when they fit |
| `oversized` | `allow` or `error` | `allow` | Mark and report, or reject, a single span above budget |

Word measurement uses Unicode whitespace boundaries. It is not a model tokenizer.

When overlap would prevent the next new span from fitting, the packer removes the oldest retained overlap spans until it can advance. Overlap may duplicate source across chunks, but no original span disappears.

## packWeighted(items, options)

`packWeighted` uses nonnegative integer weights computed by the caller. This is the integration point for model-specific tokenizers.

```js
const items = blocks.Spans.map((span) => ({
  span,
  weight: tokenizer.count(span.Text),
}));

const result = chunking.packWeighted(items, {
  maxWeight: 512,
  overlapWeight: 64,
  oversized: "allow",
});
```

The module does not validate that a weight corresponds to the span text. Record the tokenizer and model configuration in the calling application when reproducibility matters.

## recursive(source, options)

`recursive` applies increasingly fine boundaries only to ranges that exceed the budget. Nested spans are translated back to absolute source coordinates before packing.

```js
const result = chunking.recursive(source, {
  maxUnits: 1200,
  measure: "runes",
  levels: [
    "markdownSections",
    "markdownBlocks",
    "paragraphs",
    "lines",
    "runes",
  ],
  atomic: ["fencedCodeBlock", "codeBlock", "htmlBlock"],
  overlap: { unit: "spans", value: 0 },
  oversized: "allow",
});
```

Valid levels are `markdownSections`, `markdownBlocks`, `paragraphs`, `lines`, and `runes`. Unknown levels are errors. The final rune level guarantees progress for byte and rune measurement unless an atomic span is preserved.

## Result types

`SegmentResult` fields:

- `Spec` — strategy name, version, and normalized options;
- `SourceBytes`, `SourceRunes` — input size in both coordinate systems;
- `Spans` — ordered source partition;
- `Diagnostics` — result-level warnings or errors.

`Span` fields:

- `Ordinal`, `Kind`, and exact `Text`;
- byte, rune, line, and column coordinates;
- `Atomic`, `Language`, and `HeadingLevel` structure metadata;
- `HeadingPath` derived heading ancestry;
- `Level` recursive fallback level, when applicable.

`PackResult` contains `Spec`, `Chunks`, and aggregate `Diagnostics`. Each `PackedChunk` contains:

- `Text`, coordinates, and `SpanOrdinals`;
- selected `HeadingPath` and fallback `Level`;
- measured `Weight` and `Oversized` status;
- chunk-local `Diagnostics`.

## Diagnostics and errors

Allowed oversized spans produce warnings instead of disappearing. Stable diagnostic and error codes include:

- `invalid_utf8`;
- `invalid_range`;
- `source_range_mismatch`;
- `invalid_weight`;
- `unknown_measure`;
- `unknown_recursive_level`;
- `span_exceeds_budget`.

Scripts should treat `Oversized` and diagnostics as part of the result contract, not as logging text.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `unknown option` | A lower-camel option key is misspelled | Compare the object with the option table; unknown keys are intentionally rejected |
| `maxUnits must be greater than zero` | The pack budget is absent or nonpositive | Supply a positive integer `maxUnits` |
| A chunk exceeds its budget | One atomic or input span is larger than the budget and `oversized` is `allow` | Inspect `Oversized` and diagnostics, increase the budget, change atomic kinds, or use `recursive` |
| Byte offsets differ from JavaScript indices | JavaScript indices count UTF-16 code units | Slice the original UTF-8 data in Go or use the returned `Text`; do not reinterpret byte offsets as JS indices |
| Token counts differ from model limits | `runes` and `words` are deterministic approximations | Compute tokenizer weights and call `packWeighted` |

## See Also

- `goja-text help goja-text-chunking-user-guide`
- `goja-text help goja-text-markdown-api-reference`
- `goja-text chunking blocks --help`
- `goja-text chunking pack --help`
- `goja-text chunking recursive --help`
