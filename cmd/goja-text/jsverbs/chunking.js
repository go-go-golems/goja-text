const fs = require("fs");
const chunking = require("chunking");

function readFile(file) {
  return fs.readFileSync(file, "utf-8");
}

function parseList(value, fallback) {
  if (!value) return fallback;
  return String(value).split(",").map((item) => item.trim()).filter(Boolean);
}

function spanRow(span) {
  return {
    ordinal: span.Ordinal,
    kind: span.Kind,
    startByte: span.StartByte,
    endByte: span.EndByte,
    startRune: span.StartRune,
    endRune: span.EndRune,
    start: `${span.StartLine}:${span.StartColumn}`,
    end: `${span.EndLine}:${span.EndColumn}`,
    atomic: span.Atomic,
    headingLevel: span.HeadingLevel,
    headingPath: (span.HeadingPath || []).join(" / "),
    language: span.Language,
    preview: span.Text.trim().slice(0, 100),
  };
}

function blocks(file, atomic) {
  const source = readFile(file);
  const result = chunking.markdownBlocks(source, {
    atomic: parseList(atomic, ["fencedCodeBlock", "codeBlock", "htmlBlock"]),
  });
  return result.Spans.map(spanRow);
}

__verb__("blocks", {
  short: "List source-preserving top-level Markdown block spans",
  fields: {
    file: { argument: true, help: "Markdown file to segment" },
    atomic: { default: "fencedCodeBlock,codeBlock,htmlBlock", help: "Comma-separated block kinds that recursive splitting must preserve" },
  },
});

function pack(file, maxUnits, measure, overlap, oversized) {
  const source = readFile(file);
  const segmented = chunking.markdownBlocks(source, {
    atomic: ["fencedCodeBlock", "codeBlock", "htmlBlock"],
  });
  const result = chunking.pack(segmented.Spans, {
    maxUnits,
    measure,
    overlap: { unit: "spans", value: overlap },
    oversized,
  });
  return result.Chunks.map((chunk) => ({
    ordinal: chunk.Ordinal,
    weight: chunk.Weight,
    startByte: chunk.StartByte,
    endByte: chunk.EndByte,
    spanOrdinals: chunk.SpanOrdinals.join(","),
    headingPath: (chunk.HeadingPath || []).join(" / "),
    oversized: chunk.Oversized,
    diagnosticCodes: (chunk.Diagnostics || []).map((item) => item.Code).join(","),
    preview: chunk.Text.trim().slice(0, 120),
  }));
}

__verb__("pack", {
  short: "Pack complete Markdown blocks into deterministic budgeted chunks",
  fields: {
    file: { argument: true, help: "Markdown file to chunk" },
    maxUnits: { type: "int", default: 1200, help: "Maximum bytes, runes, or words per chunk" },
    measure: { type: "choice", choices: ["bytes", "runes", "words"], default: "runes", help: "Budget measurement" },
    overlap: { type: "int", default: 0, help: "Trailing complete spans repeated in the next chunk" },
    oversized: { type: "choice", choices: ["allow", "error"], default: "allow", help: "Policy for one span larger than the budget" },
  },
});

function recursive(file, maxUnits, measure, levels, overlap, oversized) {
  const result = chunking.recursive(readFile(file), {
    maxUnits,
    measure,
    levels: parseList(levels, ["markdownSections", "markdownBlocks", "paragraphs", "lines", "runes"]),
    overlap: { unit: "spans", value: overlap },
    oversized,
    atomic: ["fencedCodeBlock", "codeBlock", "htmlBlock"],
  });
  return result.Chunks.map((chunk) => ({
    ordinal: chunk.Ordinal,
    weight: chunk.Weight,
    startByte: chunk.StartByte,
    endByte: chunk.EndByte,
    level: chunk.Level,
    oversized: chunk.Oversized,
    preview: chunk.Text.trim().slice(0, 120),
  }));
}

__verb__("recursive", {
  short: "Recursively refine oversized text through ordered boundary strategies",
  fields: {
    file: { argument: true, help: "UTF-8 text or Markdown file to chunk" },
    maxUnits: { type: "int", default: 1200, help: "Maximum bytes, runes, or words per chunk" },
    measure: { type: "choice", choices: ["bytes", "runes", "words"], default: "runes", help: "Budget measurement" },
    levels: { default: "markdownSections,markdownBlocks,paragraphs,lines,runes", help: "Comma-separated fallback order" },
    overlap: { type: "int", default: 0, help: "Trailing complete spans repeated in the next chunk" },
    oversized: { type: "choice", choices: ["allow", "error"], default: "allow", help: "Final policy when an atomic span exceeds the budget" },
  },
});
