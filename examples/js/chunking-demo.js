const fs = require("fs");
const chunking = require("chunking");

const file = "examples/markdown/chunking-sample.md";
const source = fs.readFileSync(file, "utf-8");

const blocks = chunking.markdownBlocks(source, {
  atomic: ["fencedCodeBlock", "codeBlock", "htmlBlock"],
});

const packed = chunking.pack(blocks.Spans, {
  maxUnits: 220,
  measure: "runes",
  overlap: { unit: "spans", value: 1 },
  oversized: "allow",
});

const recursive = chunking.recursive(source, {
  maxUnits: 140,
  measure: "runes",
  levels: ["markdownSections", "markdownBlocks", "paragraphs", "lines", "runes"],
  overlap: { unit: "spans", value: 0 },
  oversized: "allow",
});

console.log(JSON.stringify({
  file,
  sourceBytes: blocks.SourceBytes,
  sourceRunes: blocks.SourceRunes,
  sourcePreserved: blocks.Spans.map((span) => span.Text).join("") === source,
  blocks: blocks.Spans.map((span) => ({
    ordinal: span.Ordinal,
    kind: span.Kind,
    bytes: span.EndByte - span.StartByte,
    range: [span.StartByte, span.EndByte],
    atomic: span.Atomic,
    language: span.Language,
  })),
  packed: packed.Chunks.map((chunk) => ({
    ordinal: chunk.Ordinal,
    weight: chunk.Weight,
    range: [chunk.StartByte, chunk.EndByte],
    spanOrdinals: chunk.SpanOrdinals,
    oversized: chunk.Oversized,
  })),
  recursive: recursive.Chunks.map((chunk) => ({
    ordinal: chunk.Ordinal,
    weight: chunk.Weight,
    range: [chunk.StartByte, chunk.EndByte],
    level: chunk.Level,
  })),
}, null, 2));
