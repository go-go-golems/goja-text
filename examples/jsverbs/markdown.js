const fs = require("fs");
const markdown = require("markdown");

function readMarkdown(file) {
  return fs.readFileSync(file, "utf-8");
}

function headings(file) {
  const ast = markdown.parse(readMarkdown(file));
  const rows = [];

  markdown.walk(ast, (node, ctx) => {
    if (node.Type === "heading") {
      rows.push({
        level: node.Level,
        text: markdown.textContent(node),
        depth: ctx.Depth,
        startLine: node.StartLine,
        startColumn: node.StartColumn
      });
    }
  });

  return rows;
}

__verb__("headings", {
  short: "List Markdown headings from a file",
  fields: {
    file: {
      argument: true,
      help: "Markdown file to parse"
    }
  }
});

function links(file) {
  const ast = markdown.parse(readMarkdown(file));
  const rows = [];

  markdown.walk(ast, (node) => {
    if (node.Type === "link") {
      rows.push({
        text: markdown.textContent(node),
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
  short: "List Markdown links from a file",
  fields: {
    file: {
      argument: true,
      help: "Markdown file to parse"
    }
  }
});

function summary(file) {
  const ast = markdown.parse(readMarkdown(file));
  let headingCount = 0;
  let linkCount = 0;
  let codeBlockCount = 0;

  markdown.walk(ast, (node) => {
    if (node.Type === "heading") headingCount++;
    if (node.Type === "link") linkCount++;
    if (node.Type === "codeBlock" || node.Type === "fencedCodeBlock") codeBlockCount++;
  });

  return {
    file,
    rootType: ast.Type,
    topLevelBlocks: ast.Children.length,
    headingCount,
    linkCount,
    codeBlockCount,
    valid: markdown.validate(ast).Valid
  };
}

__verb__("summary", {
  short: "Summarize a Markdown file",
  fields: {
    file: {
      argument: true,
      help: "Markdown file to parse"
    }
  }
});
