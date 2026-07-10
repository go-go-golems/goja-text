---
Title: "goja-text chunking user guide"
Slug: goja-text-chunking-user-guide
Short: "Build inspectable source-preserving chunks from text and Markdown in JavaScript."
Topics:
- goja-text
- chunking
- markdown
- retrieval
- javascript
Commands:
- goja-text run
- goja-text chunking
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

The chunking module provides the boundary and packing primitives needed before embedding or indexing a document. It keeps each result connected to the original UTF-8 source, making it possible to inspect chunk text, reproduce generation, and cite exact byte ranges.

This guide builds a Markdown chunking pipeline in stages. The stages remain separate because boundary quality, budget policy, tokenizer choice, and retrieval metadata change for different applications.

## Start with a source partition

A segmenter decides where complete units begin and end. Start with Markdown blocks when the source is Markdown and code fences, lists, block quotes, and headings should remain recognizable.

```js
const fs = require("fs");
const chunking = require("chunking");

const source = fs.readFileSync(
  "examples/markdown/chunking-sample.md",
  "utf-8"
);

const blocks = chunking.markdownBlocks(source, {
  atomic: ["fencedCodeBlock", "codeBlock", "htmlBlock"],
});
```

Immediately verify the preservation invariant while developing a new pipeline:

```js
const reconstructed = blocks.Spans
  .map((span) => span.Text)
  .join("");

if (reconstructed !== source) {
  throw new Error("segmenter did not preserve the source");
}
```

Built-in segmenters perform this validation in Go. Keeping the assertion in an exploratory script makes the intended contract visible when custom transformations are added later.

## Inspect spans before choosing a budget

Span inspection shows which structures will be indivisible during ordinary packing. This matters because a single fenced code block may be larger than the planned chunk budget.

```js
for (const span of blocks.Spans) {
  console.log({
    ordinal: span.Ordinal,
    kind: span.Kind,
    bytes: span.EndByte - span.StartByte,
    runes: span.EndRune - span.StartRune,
    atomic: span.Atomic,
    language: span.Language,
  });
}
```

The repository command packages the same inspection as structured Glazed output:

```bash
./dist/goja-text chunking blocks \
  examples/markdown/chunking-sample.md \
  --output table
```

Do this before tuning retrieval results. If the primitive spans do not match the document structures the application needs, changing only the chunk budget will not correct the boundary policy.

## Pack complete spans

Packing groups consecutive spans while their measured size remains within a budget. The packer is greedy and deterministic.

```js
const packed = chunking.pack(blocks.Spans, {
  maxUnits: 220,
  measure: "runes",
  overlap: { unit: "spans", value: 1 },
  oversized: "allow",
});
```

The selected measurement answers a specific operational question:

| Measure | Use it when | Limitation |
| --- | --- | --- |
| `bytes` | Storage and transport sizes matter | Multibyte characters count more than one |
| `runes` | A deterministic Unicode-aware text budget is sufficient | Model tokenization is different |
| `words` | A rough language-independent whitespace count is useful | Punctuation and languages without spaces are not tokenized semantically |

Overlap repeats complete trailing spans. It supplies local context without creating partial source ranges. It also increases total indexed text, so record the overlap setting with an index generation.

## Handle oversized structures explicitly

No complete-span packer can satisfy a budget smaller than one of its input spans. The `oversized` policy makes that condition visible.

With `allow`, the chunk is emitted with `Oversized: true` and a diagnostic:

```js
for (const chunk of packed.Chunks) {
  if (chunk.Oversized) {
    console.error(
      "oversized chunk",
      chunk.Ordinal,
      chunk.Diagnostics.map((item) => item.Code)
    );
  }
}
```

With `error`, packing stops instead. Use `error` in a pipeline that must never publish an index with budget violations. Use `allow` in an exploration tool that needs to show the problematic source and let an operator revise the strategy.

## Use recursive fallback when coarse units are too large

Recursive chunking begins with strong document boundaries and refines only oversized ranges. This preserves sections and blocks when they fit while still making progress on long prose.

```js
const recursive = chunking.recursive(source, {
  maxUnits: 140,
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

The order is policy. Put stronger boundaries first and mechanical boundaries last. The final `runes` level creates fixed windows when section, block, paragraph, and line boundaries are still too large.

Atomic kinds stop refinement. An atomic fence larger than the budget therefore remains oversized. Remove that kind from `atomic` only when splitting its exact source text is acceptable to the application.

Run the bundled command to compare the fallback output:

```bash
./dist/goja-text chunking recursive \
  examples/markdown/chunking-sample.md \
  --max-units 140 \
  --measure runes \
  --output json
```

## Integrate an actual model tokenizer

The library intentionally does not select an embedding model or tokenizer. A model-specific pipeline computes one weight for each span and passes those weights to `packWeighted`.

```js
const weighted = chunking.packWeighted(
  blocks.Spans.map((span) => ({
    span,
    weight: tokenizer.count(span.Text),
  })),
  {
    maxWeight: 512,
    overlapWeight: 48,
    oversized: "allow",
  }
);
```

The tokenizer function is application code. Record its model name, vocabulary revision, normalization settings, special-token policy, and library version beside the chunking strategy. Without that metadata, the same source may produce different weights after a model or tokenizer upgrade.

## Store retrieval metadata outside the text primitive

The chunking result provides ranges and structural metadata, but an indexing application still owns document identity and generation metadata. A practical indexed record can include:

```js
const record = {
  documentId,
  generationId,
  chunkOrdinal: chunk.Ordinal,
  sourceRange: [chunk.StartByte, chunk.EndByte],
  headingPath: chunk.HeadingPath,
  strategy: recursive.Spec,
  embeddingModel,
  text: chunk.Text,
  vector,
};
```

Do not use `StartByte` alone as a global chunk identifier. It is meaningful only together with the source document and the exact document revision.

## Compare strategies in an exploration loop

A useful exploration script keeps segmentation, weighting, packing, and evaluation as replaceable functions:

```text
source
  -> segment(strategy)
  -> inspect spans and preservation
  -> measure(model or deterministic unit)
  -> pack(budget and overlap)
  -> embed
  -> search evaluation queries
  -> compare retrieval evidence
```

Pseudocode for a strategy sweep:

```text
for each segmentation strategy:
    spans = segment(source, strategy)
    assert join(spans.text) == source

    for each budget and overlap:
        chunks = pack(spans, budget, overlap)
        vectors = embed(chunks.text)
        scores = evaluate(search(vectors, queries))
        save(strategy, chunks, scores, diagnostics)
```

The module covers the first three operations. Embedding, vector storage, search, and evaluation remain application concerns so they can evolve independently.

## Run the complete example

The included script prints block, packed, and recursive summaries:

```bash
make build-xgoja
./dist/goja-text run examples/js/chunking-demo.js
```

Read `examples/js/chunking-demo.js` before extending it. It is deliberately small enough to copy into an experiment while still checking source preservation and reporting exact coordinates.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| A fence is emitted above budget | It is atomic and larger than `maxUnits` | Increase the budget, accept the diagnostic, or remove the atomic kind |
| Recursive chunks still exceed budget | An atomic span was preserved, or the final policy allows oversized content | Inspect `Oversized`, `Level`, and diagnostics; revise the atomic policy |
| Search results lack heading context | The chosen segmenter does not expose heading paths | Start with `markdownSections`, or attach heading metadata from section spans |
| Indexed size grows unexpectedly | Whole-span overlap repeats text | Reduce `overlap.value` or `overlapWeight` and regenerate the index |
| A tokenizer count does not match `runes` | Model tokens are not Unicode code points | Use `packWeighted` with the exact production tokenizer |
| A script reads `result.spans` as undefined | Results are Go-backed | Read PascalCase fields such as `result.Spans` |

## See Also

- `goja-text help goja-text-chunking-api-reference`
- `goja-text help goja-text-markdown-user-guide`
- `goja-text chunking blocks --help`
- `goja-text chunking pack --help`
- `goja-text chunking recursive --help`
