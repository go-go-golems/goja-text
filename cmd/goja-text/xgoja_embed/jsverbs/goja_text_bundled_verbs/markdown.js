const fs = require("fs");
const assets = require("fs:assets");
const yaml = require("yaml");
const markdown = require("markdown");

function readFile(file) {
  return fs.readFileSync(file, "utf-8");
}

function slugify(text) {
  return String(text || "")
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\s-]/g, "")
    .replace(/\s+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");
}

function toc(file, maxLevel) {
  const ast = markdown.parse(readFile(file));
  const rows = [];
  const limit = maxLevel || 6;

  markdown.walk(ast, (node) => {
    if (node.Type !== "heading" || node.Level > limit) {
      return;
    }
    const text = markdown.textContent(node);
    rows.push({
      level: node.Level,
      text,
      anchor: slugify(text),
      startLine: node.StartLine,
      startColumn: node.StartColumn
    });
  });

  return rows;
}

__verb__("toc", {
  short: "Build a Markdown table of contents",
  fields: {
    file: { argument: true, help: "Markdown file to parse" },
    maxLevel: { type: "int", default: 3, help: "Maximum heading level to include" }
  }
});

function links(file) {
  const ast = markdown.parse(readFile(file));
  const rows = [];

  markdown.walk(ast, (node) => {
    if (node.Type === "link") {
      rows.push({
        kind: "link",
        text: markdown.textContent(node),
        destination: node.Destination,
        title: node.Title,
        startLine: node.StartLine,
        startColumn: node.StartColumn
      });
    }
    if (node.Type === "image") {
      rows.push({
        kind: "image",
        text: node.Alt,
        destination: node.Destination,
        title: node.Title,
        startLine: node.StartLine,
        startColumn: node.StartColumn
      });
    }
  });

  return rows;
}

__verb__("links", {
  short: "List Markdown links and images",
  fields: {
    file: { argument: true, help: "Markdown file to parse" }
  }
});

function summary(file) {
  const ast = markdown.parse(readFile(file));
  const counts = {
    file,
    rootType: ast.Type,
    topLevelBlocks: ast.Children.length,
    headings: 0,
    links: 0,
    images: 0,
    codeBlocks: 0,
    htmlBlocks: 0,
    valid: markdown.validate(ast).Valid
  };

  markdown.walk(ast, (node) => {
    if (node.Type === "heading") counts.headings++;
    if (node.Type === "link") counts.links++;
    if (node.Type === "image") counts.images++;
    if (node.Type === "codeBlock" || node.Type === "fencedCodeBlock") counts.codeBlocks++;
    if (node.Type === "htmlBlock" || node.Type === "rawHTML") counts.htmlBlocks++;
  });

  return counts;
}

__verb__("summary", {
  short: "Summarize Markdown document structure",
  fields: {
    file: { argument: true, help: "Markdown file to parse" }
  }
});

const builderExampleSpecs = {
  report: {
    dataPath: "/markdown-builder/report.yaml",
    description: "Sprint-style Markdown report with status table and checklist"
  },
  "api-table": {
    dataPath: "/markdown-builder/api-table.yaml",
    description: "API reference table generated from structured function metadata"
  }
};

const builderHelpers = {
  readAssetYaml(path) {
    return yaml.parse(assets.readFileSync(path, "utf-8"));
  },

  writeMaybe(outputPath, text) {
    if (outputPath) {
      fs.writeFileSync(outputPath, text);
      return { outputPath, bytes: text.length };
    }
    return text;
  },

  renderReport(data) {
    const doc = markdown.builder()
      .Title(data.title)
      .Paragraph(data.summary)
      .Table()
        .Columns(
          { label: "Name", align: "left" },
          { label: "Status", align: "center" },
          { label: "Owner", align: "left" }
        );

    for (const row of data.statusRows || []) {
      doc.Row(row.name, row.status, row.owner);
    }

    return doc.End()
      .Heading(2, "Next steps")
      .Checklist(data.nextSteps || [])
      .RenderString();
  },

  renderApiTable(data) {
    const i = markdown.inline();
    const table = markdown.builder()
      .Title(data.title)
      .Paragraph(data.summary)
      .Table()
        .Columns(
          { label: "Function", align: "left" },
          { label: "Returns", align: "left" },
          "Description"
        );

    for (const fn of data.functions || []) {
      table.Row(i.Code(fn.name), i.Code(fn.returns), fn.description);
    }

    return table.End().RenderString();
  }
};

function builderExamples() {
  return Object.keys(builderExampleSpecs).map((name) => ({
    name,
    description: builderExampleSpecs[name].description,
    dataPath: builderExampleSpecs[name].dataPath,
    command: `goja-text markdown builder-example ${name}`
  }));
}

__verb__("builderExamples", {
  short: "List embedded Markdown builder examples"
});

function builderExample(name, outputPath) {
  const key = name || "report";
  const spec = builderExampleSpecs[key];
  if (!spec) {
    throw new Error(`unknown Markdown builder example ${key}; choose one of ${Object.keys(builderExampleSpecs).join(", ")}`);
  }
  const data = builderHelpers.readAssetYaml(spec.dataPath);
  const text = key === "api-table" ? builderHelpers.renderApiTable(data) : builderHelpers.renderReport(data);
  return builderHelpers.writeMaybe(outputPath, text);
}

__verb__("builderExample", {
  short: "Render one embedded Markdown builder example",
  fields: {
    name: { argument: true, default: "report", type: "choice", choices: ["report", "api-table"], help: "Embedded builder example to render" },
    outputPath: { help: "Optional file path to write rendered Markdown" }
  }
});
