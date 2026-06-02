const fs = require("fs");
const extract = require("extract");

function readText(file) {
  return fs.readFileSync(file, "utf-8");
}

function rowFor(candidate) {
  return {
    kind: candidate.Kind,
    format: candidate.Format,
    wrapper: candidate.Wrapper,
    label: candidate.Label,
    confidence: candidate.Confidence,
    startRow: candidate.StartRow,
    startColumn: candidate.StartColumn,
    payloadStartRow: candidate.PayloadStartRow,
    payloadStartColumn: candidate.PayloadStartColumn,
    textPreview: (candidate.Text || "").trim().slice(0, 80)
  };
}

function list(file) {
  return extract.all(readText(file)).map(rowFor);
}

__verb__("list", {
  short: "List structured-data candidates in a text file",
  fields: {
    file: {
      argument: true,
      help: "Text or Markdown file to inspect"
    }
  }
});

function validate(file) {
  return extract.all(readText(file)).map((candidate) => {
    const validation = extract.validate(candidate);
    const row = rowFor(candidate);
    row.valid = validation.Valid;
    row.sanitizedPreview = validation.Sanitized ? validation.Sanitized.trim().slice(0, 120) : "";
    row.issueCount = (validation.Issues || []).length;
    row.fixCount = (validation.Fixes || []).length;
    return row;
  });
}

__verb__("validate", {
  short: "Extract and validate structured-data candidates in a text file",
  fields: {
    file: {
      argument: true,
      help: "Text or Markdown file to inspect"
    }
  }
});

function firstValid(file) {
  const candidates = extract.all(readText(file));
  for (const candidate of candidates) {
    const validation = extract.validate(candidate);
    if (validation.Valid) {
      return {
        candidate: rowFor(candidate),
        sanitized: validation.Sanitized
      };
    }
  }
  return null;
}

__verb__("firstValid", {
  short: "Return the first valid structured-data candidate from a text file",
  fields: {
    file: {
      argument: true,
      help: "Text or Markdown file to inspect"
    }
  }
});
