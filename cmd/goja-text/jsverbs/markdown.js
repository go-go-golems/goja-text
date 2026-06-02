const fs = require("fs");
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
