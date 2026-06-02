const fs = require("fs");
const sanitize = require("sanitize");

function readFile(file) {
  return fs.readFileSync(file, "utf-8");
}

function fixRows(result) {
  return (result.Fixes || []).map((fix, index) => ({
    index,
    rule: fix.Rule,
    description: fix.Description,
    before: fix.Before,
    after: fix.After
  }));
}

function json(file, maxIterations) {
  const options = sanitize.json.options()
    .MaxIterations(maxIterations || 5)
    .Build();
  const result = sanitize.json.sanitize(readFile(file), options);

  return {
    file,
    format: "json",
    sanitized: result.Sanitized,
    parseClean: result.ParseClean,
    lintClean: result.LintClean,
    strictParseClean: result.StrictParseClean,
    fixCount: (result.Fixes || []).length,
    fixes: fixRows(result)
  };
}

__verb__("json", {
  short: "Repair a JSON-like file and report fixes",
  fields: {
    file: { argument: true, help: "JSON or JSON-like file to repair" },
    maxIterations: { type: "int", default: 5, help: "Maximum sanitizer iterations" }
  }
});

function yaml(file, maxIterations, tabWidth) {
  const options = sanitize.yaml.options()
    .MaxIterations(maxIterations || 5)
    .TabWidth(tabWidth || 2)
    .Build();
  const result = sanitize.yaml.sanitize(readFile(file), options);

  return {
    file,
    format: "yaml",
    sanitized: result.Sanitized,
    parseClean: result.ParseClean,
    lintClean: result.LintClean,
    fixCount: (result.Fixes || []).length,
    fixes: fixRows(result)
  };
}

__verb__("yaml", {
  short: "Repair a YAML-like file and report fixes",
  fields: {
    file: { argument: true, help: "YAML or YAML-like file to repair" },
    maxIterations: { type: "int", default: 5, help: "Maximum sanitizer iterations" },
    tabWidth: { type: "int", default: 2, help: "Spaces used when expanding tabs" }
  }
});

function lintJson(file) {
  const issues = sanitize.json.lint(readFile(file));
  return issues.map((issue) => ({
    rule: issue.Rule,
    source: issue.Source,
    description: issue.Description,
    row: issue.Row,
    startByte: issue.StartByte,
    endByte: issue.EndByte
  }));
}

__verb__("lintJson", {
  short: "Lint JSON-like input without returning repaired text",
  fields: {
    file: { argument: true, help: "JSON or JSON-like file to lint" }
  }
});

function lintYaml(file) {
  const issues = sanitize.yaml.lint(readFile(file));
  return issues.map((issue) => ({
    rule: issue.Rule,
    source: issue.Source,
    description: issue.Description,
    row: issue.Row,
    startByte: issue.StartByte,
    endByte: issue.EndByte
  }));
}

__verb__("lintYaml", {
  short: "Lint YAML-like input without returning repaired text",
  fields: {
    file: { argument: true, help: "YAML or YAML-like file to lint" }
  }
});

function rules(format) {
  if (format === "json") return sanitize.json.rules();
  if (format === "yaml") return sanitize.yaml.rules();
  throw new Error("format must be json or yaml");
}

__verb__("rules", {
  short: "List sanitize rules for JSON or YAML",
  fields: {
    format: { argument: true, type: "choice", choices: ["json", "yaml"], help: "Rule catalog to show" }
  }
});
