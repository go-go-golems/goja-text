const fs = require("fs");
const sanitize = require("sanitize");

const yamlSource = fs.readFileSync("examples/yaml/broken.yaml", "utf-8");
const yamlConfig = sanitize.yaml.options()
  .MaxIterations(5)
  .TabWidth(2)
  .Build();
const yamlResult = sanitize.yaml.sanitize(yamlSource, yamlConfig);

const jsonSource = fs.readFileSync("examples/json/broken.json", "utf-8");
const jsonConfig = sanitize.json.options()
  .MaxIterations(5)
  .Build();
const jsonResult = sanitize.json.sanitize(jsonSource, jsonConfig);

console.log(JSON.stringify({
  Yaml: {
    Sanitized: yamlResult.Sanitized,
    ParseClean: yamlResult.ParseClean,
    LintClean: yamlResult.LintClean,
    FixRules: yamlResult.Fixes.map((fix) => fix.Rule),
    RuleCount: sanitize.yaml.rules().length,
  },
  Json: {
    Sanitized: jsonResult.Sanitized,
    ParseClean: jsonResult.ParseClean,
    LintClean: jsonResult.LintClean,
    StrictParseClean: jsonResult.StrictParseClean,
    FixRules: jsonResult.Fixes.map((fix) => fix.Rule),
    RuleCount: sanitize.json.rules().length,
  },
  StrictParse: sanitize.json.strictParse('{"ok": true}'),
}, null, 2));
