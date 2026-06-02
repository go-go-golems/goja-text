const fs = require("fs");
const extract = require("extract");

const source = fs.readFileSync("examples/text/structured-data-sample.md", "utf-8");
const options = extract.options()
  .IncludeDiagnostics(true)
  .Build();

const candidates = extract.all(source, options);
const summary = candidates.map((candidate) => {
  const validation = extract.validate(candidate);
  return {
    Kind: candidate.Kind,
    Format: candidate.Format,
    Wrapper: candidate.Wrapper,
    Label: candidate.Label,
    StartRow: candidate.StartRow,
    TextPreview: candidate.Text.trim().slice(0, 60),
    Valid: validation.Valid,
    SanitizedPreview: validation.Sanitized ? validation.Sanitized.trim().slice(0, 60) : "",
  };
});

console.log(JSON.stringify({
  Count: candidates.length,
  Candidates: summary,
}, null, 2));
