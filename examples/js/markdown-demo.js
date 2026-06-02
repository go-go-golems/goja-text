const fs = require("fs");
const markdown = require("markdown");

const source = fs.readFileSync("examples/markdown/sample.md", "utf-8");
const ast = markdown.parse(source);

const headings = [];
const links = [];
markdown.walk(ast, (node, ctx) => {
  if (node.Type === "heading") {
    headings.push({
      Level: node.Level,
      Text: markdown.textContent(node),
      Depth: ctx.Depth,
    });
  }
  if (node.Type === "link") {
    links.push({
      Destination: node.Destination,
      Text: markdown.textContent(node),
    });
  }
});

console.log(JSON.stringify({
  RootType: ast.Type,
  TopLevelBlocks: ast.Children.length,
  Headings: headings,
  Links: links,
  Valid: markdown.validate(ast).Valid,
}, null, 2));
