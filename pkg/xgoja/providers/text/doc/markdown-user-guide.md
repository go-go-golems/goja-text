---
Title: "goja-text markdown user guide"
Slug: goja-text-markdown-user-guide
Short: "A guided introduction to parsing, traversing, and rendering Markdown from JavaScript."
Topics:
- goja-text
- markdown
- guide
- javascript
Commands:
- goja-text eval
- goja-text run
- goja-text markdown
- goja-text sanitize
- goja-text extract
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

The markdown module exists for scripts that need to understand Markdown as a document, not merely transform it as text. A Markdown file contains structure: headings introduce sections, links point outside the document, images carry destinations and alternate text, and code fences often contain data that another tool wants to inspect. `require("markdown")` gives JavaScript access to that structure while keeping the parser and the domain objects in Go.

The module also helps scripts create Markdown. Use the fluent `builder()` API when the script has structured data and wants to generate a clean report, table, checklist, prompt, or release note without hand-written string concatenation.

The central parsing idea is simple: parse once, then ask specific questions by walking the tree. The module does not try to predict every question a script might ask. Instead, it exposes a reliable traversal primitive, `walk()`, and lets JavaScript express the document query that matters for the current task.

## The first parse

Start by loading the module and parsing a string. The result is a Go-backed `MarkdownNode`, so JavaScript reads fields with exported Go names such as `Type`, `Children`, and `Level`.

```js
const markdown = require("markdown");

const ast = markdown.parse(`# Title

See [the docs](https://example.com/docs).
`);

console.log(ast.Type);             // "document"
console.log(ast.Children[0].Type); // "heading"
console.log(ast.Children[0].Level); // 1
```

This PascalCase shape is intentional. The same object can cross back into Go functions such as `textContent()` and `walk()`, where Go can validate that it is receiving a real Markdown node rather than a JavaScript object that merely looks similar.

## Turning structure into a useful query

Most scripts should follow a two-step pattern. First parse the document. Then use `walk()` to collect exactly the facts the script needs.

```js
const headings = [];
const links = [];

markdown.walk(ast, (node, ctx) => {
  if (node.Type === "heading") {
    headings.push({
      level: node.Level,
      text: markdown.textContent(node),
      depth: ctx.Depth,
    });
  }

  if (node.Type === "link") {
    links.push({
      text: markdown.textContent(node),
      destination: node.Destination,
      title: node.Title,
    });
  }
});

console.log(JSON.stringify({ headings, links }, null, 2));
```

This pattern is more flexible than a collection of narrow Go helpers. A table-of-contents script, a link checker, and a documentation linter all need the same parser, but they ask different questions. `walk()` keeps the Go API small while giving JavaScript room to describe the policy.

## Controlling traversal

A visitor can return a small control value. Return `"stop"` when the script has found enough information, or return `"skip"` when the current subtree is not relevant.

```js
let firstExternalLink = null;

markdown.walk(ast, (node) => {
  if (node.Type === "link" && /^https?:/.test(node.Destination || "")) {
    firstExternalLink = node.Destination;
    return "stop";
  }
});
```

The control values are deliberately plain JavaScript values:

- `undefined` or `true` means traversal continues normally.
- `false` or `"skip"` skips the current node's children.
- `"stop"` ends the traversal immediately.

## Building Markdown from data

Use `builder()` when the goal is output generation rather than analysis. The builder is Go-backed: JavaScript calls fluent methods, while Go owns block spacing, table formatting, escaping, validation, and final serialization.

```js
const markdown = require("markdown");

const report = markdown.builder()
  .Title("Sprint report")
  .Paragraph("Generated from structured runtime data.")
  .Table()
    .Columns({ label: "Name", align: "left" }, { label: "Status", align: "right" })
    .Row("Parser", "done")
    .Row("Builder", "planned")
    .End()
  .Heading(2, "Next steps")
  .Checklist([
    { text: "Expose goja API", checked: true },
    { text: "Write docs" },
  ])
  .RenderString();

console.log(report);
```

The output is ordinary Markdown:

```markdown
# Sprint report

Generated from structured runtime data.

| Name    | Status  |
| :------ | ------: |
| Parser  | done    |
| Builder | planned |

## Next steps

- [x] Expose goja API
- [ ] Write docs
```

Builder methods intentionally use PascalCase (`Title`, `Paragraph`, `RenderString`) because they are exported Go methods. This matches the rest of the goja-text API.

## Use inline helpers when strings are not enough

Ordinary strings are escaped as text. Use `inline()` when a paragraph or table cell needs a code span, link, emphasis, strong text, or explicit raw Markdown.

```js
const i = markdown.inline();

const text = markdown.builder()
  .Paragraph(
    "Run ",
    i.Code("go test ./..."),
    " and read ",
    i.Link("the guide", "https://example.com"),
    "."
  )
  .RenderString();
```

Prefer normal strings for untrusted data. Use `i.Raw()` and `builder.Raw()` only for trusted Markdown fragments.

## Rendering is a different question

Use `renderHTML()` when the goal is presentation. Use `parse()` and `walk()` when the goal is analysis. Use `builder()` when the goal is Markdown generation.

```bash
./dist/goja-text eval 'const md = require("markdown"); md.renderHTML("## Title\n\n**bold**")'
```

HTML is an output format. The AST is the document model. Keeping that distinction clear prevents scripts from scraping rendered HTML when the original Markdown tree already contains the needed information.

## Running the included examples

The repository includes a conventional demo script and a root-mounted JavaScript verb command source. The demo is useful when you want to see console-oriented JavaScript. The jsverb is useful when you want Glazed to render structured rows.

```bash
./dist/goja-text run examples/js/markdown-demo.js
./dist/goja-text markdown headings examples/markdown/sample.md
```

The `markdown headings` verb reads a file with the host `fs` module, parses it, walks the AST, and returns rows containing heading level, text, and source depth. That is the same pattern shown above, packaged as a command.

## Key points

- The Markdown AST is Go-backed so Go can validate nodes when JavaScript passes them back into module functions.
- JavaScript reads exported fields with PascalCase names such as `Type`, `Children`, `Level`, and `Destination`.
- `walk()` is the primary extension point. Write document-specific queries in JavaScript instead of waiting for a new Go helper.
- Use `renderHTML()` for presentation, `parse()` for structure, and `builder()` for generated Markdown output.
