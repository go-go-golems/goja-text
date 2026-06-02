const fs = require("fs");
const sanitize = require("sanitize");

function fixRules(result) {
  return (result.Fixes || []).map((fix) => fix.Rule);
}

function yaml(file, maxIterations) {
  const input = fs.readFileSync(file, "utf-8");
  const options = sanitize.yaml.options()
    .MaxIterations(maxIterations)
    .TabWidth(2)
    .Build();
  const result = sanitize.yaml.sanitize(input, options);

  return {
    file,
    format: "yaml",
    sanitized: result.Sanitized,
    parseClean: result.ParseClean,
    lintClean: result.LintClean,
    fixCount: (result.Fixes || []).length,
    fixRules: fixRules(result)
  };
}

__verb__("yaml", {
  short: "Repair a YAML file and report applied rules",
  fields: {
    file: {
      argument: true,
      help: "YAML file to repair"
    },
    maxIterations: {
      type: "int",
      default: 5,
      help: "Maximum sanitizer iterations"
    }
  }
});

function json(file, maxIterations) {
  const input = fs.readFileSync(file, "utf-8");
  const options = sanitize.json.options()
    .MaxIterations(maxIterations)
    .Build();
  const result = sanitize.json.sanitize(input, options);

  return {
    file,
    format: "json",
    sanitized: result.Sanitized,
    parseClean: result.ParseClean,
    lintClean: result.LintClean,
    strictParseClean: result.StrictParseClean,
    fixCount: (result.Fixes || []).length,
    fixRules: fixRules(result)
  };
}

__verb__("json", {
  short: "Repair a JSON file and report applied rules",
  fields: {
    file: {
      argument: true,
      help: "JSON file to repair"
    },
    maxIterations: {
      type: "int",
      default: 5,
      help: "Maximum sanitizer iterations"
    }
  }
});

function rules(format) {
  if (format === "json") {
    return sanitize.json.rules();
  }
  if (format === "yaml") {
    return sanitize.yaml.rules();
  }
  throw new Error("format must be json or yaml");
}

__verb__("rules", {
  short: "Show sanitize rule catalogs",
  fields: {
    format: {
      argument: true,
      type: "choice",
      choices: ["json", "yaml"],
      help: "Rule catalog to show"
    }
  }
});
