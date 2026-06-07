const fs = require("fs");
const yaml = require("yaml");
const template = require("template");

const helpers = {
  readFile(file) {
    return fs.readFileSync(file, "utf-8");
  },

  writeMaybe(outputPath, text) {
    if (outputPath) {
      fs.writeFileSync(outputPath, text);
      return { outputPath, bytes: text.length };
    }
    return text;
  },

  parseDataFile(dataFile) {
    if (!dataFile) return {};
    const source = helpers.readFile(dataFile);
    if (dataFile.endsWith(".json")) return JSON.parse(source);
    return yaml.parse(source);
  },

  parseFuncSets(funcs) {
    if (!funcs) return ["sprig", "glazed"];
    if (funcs === "none") return ["none"];
    return String(funcs).split(",").map((s) => s.trim()).filter((s) => s.length > 0);
  },

  configureBuilder(builder, options) {
    builder.Name(options.name || "template");
    builder.MissingKey(options.missingKey || "error");
    builder.Funcs(...helpers.parseFuncSets(options.funcs));
    if (options.leftDelim || options.rightDelim) {
      if (!options.leftDelim || !options.rightDelim) {
        throw new Error("leftDelim and rightDelim must be supplied together");
      }
      builder.Delims(options.leftDelim, options.rightDelim);
    }
    return builder;
  }
};

function text(templateFile, dataFile, outputPath, name, templateName, missingKey, funcs, leftDelim, rightDelim) {
  const source = helpers.readFile(templateFile);
  const data = helpers.parseDataFile(dataFile);
  const builder = helpers.configureBuilder(template.text(), { name, missingKey, funcs, leftDelim, rightDelim });
  const set = builder.Parse(source);
  const result = templateName ? set.RenderTemplate(templateName, data) : set.Render(data);
  return helpers.writeMaybe(outputPath, result.Text);
}

__verb__("text", {
  short: "Render a Go text/template file with YAML or JSON data",
  fields: {
    templateFile: { argument: true, help: "Template file to parse" },
    dataFile: { help: "YAML or JSON data file; omitted means empty data" },
    outputPath: { help: "Optional file path to write rendered output" },
    name: { default: "template", help: "Root template name" },
    templateName: { help: "Named template to execute instead of the root template" },
    missingKey: { default: "error", type: "choice", choices: ["default", "invalid", "zero", "error"], help: "Go template missingkey policy" },
    funcs: { default: "sprig,glazed", help: "Comma-separated helper presets: sprig,glazed or none" },
    leftDelim: { help: "Custom left delimiter; requires rightDelim" },
    rightDelim: { help: "Custom right delimiter; requires leftDelim" }
  }
});

function html(templateFile, dataFile, outputPath, name, templateName, missingKey, funcs, leftDelim, rightDelim) {
  const source = helpers.readFile(templateFile);
  const data = helpers.parseDataFile(dataFile);
  const builder = helpers.configureBuilder(template.html(), { name, missingKey, funcs, leftDelim, rightDelim });
  const set = builder.Parse(source);
  const result = templateName ? set.RenderTemplate(templateName, data) : set.Render(data);
  return helpers.writeMaybe(outputPath, result.Text);
}

__verb__("html", {
  short: "Render a Go html/template file with contextual escaping",
  fields: {
    templateFile: { argument: true, help: "HTML template file to parse" },
    dataFile: { help: "YAML or JSON data file; omitted means empty data" },
    outputPath: { help: "Optional file path to write rendered output" },
    name: { default: "template", help: "Root template name" },
    templateName: { help: "Named template to execute instead of the root template" },
    missingKey: { default: "error", type: "choice", choices: ["default", "invalid", "zero", "error"], help: "Go template missingkey policy" },
    funcs: { default: "sprig,glazed", help: "Comma-separated helper presets: sprig,glazed or none" },
    leftDelim: { help: "Custom left delimiter; requires rightDelim" },
    rightDelim: { help: "Custom right delimiter; requires leftDelim" }
  }
});

function inspect(templateFile, mode, name, funcs, missingKey, leftDelim, rightDelim) {
  const source = helpers.readFile(templateFile);
  const builder = helpers.configureBuilder(mode === "html" ? template.html() : template.text(), { name, funcs, missingKey, leftDelim, rightDelim });
  const validation = builder.Validate();
  if (!validation.Valid) return { valid: false, errors: validation.Errors };
  const set = builder.Parse(source);
  return set.Templates().map((info) => ({ name: info.Name, mode: info.Mode, defined: info.Defined, isDefault: info.Name === set.Name }));
}

__verb__("inspect", {
  short: "List templates defined by a Go template file",
  fields: {
    templateFile: { argument: true, help: "Template file to parse" },
    mode: { default: "text", type: "choice", choices: ["text", "html"], help: "Template engine to use" },
    name: { default: "template", help: "Root template name" },
    funcs: { default: "sprig,glazed", help: "Comma-separated helper presets: sprig,glazed or none" },
    missingKey: { default: "error", type: "choice", choices: ["default", "invalid", "zero", "error"], help: "Go template missingkey policy" },
    leftDelim: { help: "Custom left delimiter; requires rightDelim" },
    rightDelim: { help: "Custom right delimiter; requires leftDelim" }
  }
});

function check(templateFile, mode, name, funcs, missingKey, leftDelim, rightDelim) {
  const source = helpers.readFile(templateFile);
  const builder = helpers.configureBuilder(mode === "html" ? template.html() : template.text(), { name, funcs, missingKey, leftDelim, rightDelim });
  const validation = builder.Validate();
  if (!validation.Valid) return validation;
  try {
    const set = builder.Parse(source);
    return { Valid: true, Errors: [], Templates: set.Templates().map((info) => info.Name) };
  } catch (err) {
    return { Valid: false, Errors: [String(err && err.message ? err.message : err)] };
  }
}

__verb__("check", {
  short: "Validate template options and parse a Go template file",
  fields: {
    templateFile: { argument: true, help: "Template file to validate" },
    mode: { default: "text", type: "choice", choices: ["text", "html"], help: "Template engine to use" },
    name: { default: "template", help: "Root template name" },
    funcs: { default: "sprig,glazed", help: "Comma-separated helper presets: sprig,glazed or none" },
    missingKey: { default: "error", type: "choice", choices: ["default", "invalid", "zero", "error"], help: "Go template missingkey policy" },
    leftDelim: { help: "Custom left delimiter; requires rightDelim" },
    rightDelim: { help: "Custom right delimiter; requires leftDelim" }
  }
});

function helperDemo(name) {
  const result = template.text()
    .JSFunc("badge", (value) => `[[${String(value).toUpperCase()}]]`)
    .Parse("{{ badge .Name }}")
    .Render({ Name: name || "goja-text" });
  return { text: result.Text, note: "JSFunc helpers are synchronous and return ordinary values to Go templates" };
}

__verb__("helperDemo", {
  short: "Demonstrate a synchronous JSFunc template helper",
  fields: {
    name: { default: "goja-text", help: "Name to pass through the JS helper" }
  }
});
