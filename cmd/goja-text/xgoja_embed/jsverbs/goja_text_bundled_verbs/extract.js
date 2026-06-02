const fs = require("fs");
const extract = require("extract");

function readFile(file) {
  return fs.readFileSync(file, "utf-8");
}

function candidateRow(candidate) {
  return {
    kind: candidate.Kind,
    format: candidate.Format,
    wrapper: candidate.Wrapper,
    label: candidate.Label,
    confidence: candidate.Confidence,
    start: `${candidate.StartRow}:${candidate.StartColumn}`,
    end: `${candidate.EndRow}:${candidate.EndColumn}`,
    payloadStart: `${candidate.PayloadStartRow}:${candidate.PayloadStartColumn}`,
    payloadEnd: `${candidate.PayloadEndRow}:${candidate.PayloadEndColumn}`,
    startRow: candidate.StartRow,
    startColumn: candidate.StartColumn,
    endRow: candidate.EndRow,
    endColumn: candidate.EndColumn,
    payloadStartByte: candidate.PayloadStartByte,
    payloadEndByte: candidate.PayloadEndByte,
    payloadStartRow: candidate.PayloadStartRow,
    payloadStartColumn: candidate.PayloadStartColumn,
    payloadEndRow: candidate.PayloadEndRow,
    payloadEndColumn: candidate.PayloadEndColumn,
    textPreview: (candidate.Text || "").trim().slice(0, 100)
  };
}

function list(file, minConfidence) {
  const options = extract.options()
    .MinConfidence(minConfidence || 0)
    .Build();
  return extract.all(readFile(file), options).map(candidateRow);
}

__verb__("list", {
  short: "List structured-data candidates in text",
  fields: {
    file: { argument: true, help: "Text or Markdown file to inspect" },
    minConfidence: { type: "float", default: 0, help: "Minimum candidate confidence" }
  }
});

function validate(file, minConfidence) {
  const options = extract.options()
    .MinConfidence(minConfidence || 0)
    .Build();
  return extract.all(readFile(file), options).map((candidate) => {
    const validation = extract.validate(candidate);
    const row = candidateRow(candidate);
    row.valid = validation.Valid;
    row.sanitizedPreview = validation.Sanitized ? validation.Sanitized.trim().slice(0, 160) : "";
    row.errorCount = (validation.Errors || []).length;
    row.fixCount = (validation.Fixes || []).length;
    return row;
  });
}

__verb__("validate", {
  short: "Extract and validate structured-data candidates",
  fields: {
    file: { argument: true, help: "Text or Markdown file to inspect" },
    minConfidence: { type: "float", default: 0, help: "Minimum candidate confidence" }
  }
});

function firstValid(file) {
  const candidates = extract.all(readFile(file));
  for (const candidate of candidates) {
    const validation = extract.validate(candidate);
    if (validation.Valid) {
      return {
        candidate: candidateRow(candidate),
        sanitized: validation.Sanitized
      };
    }
  }
  return null;
}

__verb__("firstValid", {
  short: "Return the first valid structured-data candidate",
  fields: {
    file: { argument: true, help: "Text or Markdown file to inspect" }
  }
});

function markdownBlocks(file) {
  return extract.markdownCodeBlocks(readFile(file)).map(candidateRow);
}

__verb__("markdownBlocks", {
  short: "Extract only Markdown fenced code-block candidates",
  fields: {
    file: { argument: true, help: "Markdown file to inspect" }
  }
});
