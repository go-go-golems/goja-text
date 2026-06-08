---
Title: "goja-text markdown JavaScript API reference"
Slug: goja-text-markdown-api-reference
Short: "Reference for require(\"markdown\") in goja-text xgoja runtimes."
Topics:
- goja-text
- markdown
- javascript
- xgoja
Commands:
- goja-text
- goja-text eval
- goja-text run
- help
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

Use `require("markdown")` when JavaScript needs a Go-backed Markdown parser, HTML renderer, AST traversal primitive, or fluent Markdown document builder.

The module keeps parsed nodes and generated-document builders as Go objects. JavaScript therefore reads exported Go fields and calls exported Go methods with PascalCase names such as `Type`, `Children`, `Level`, `Destination`, `Title`, `Text`, `RenderString`, and `Table`.

## Loading

```js
const markdown = require("markdown");
```

The repository xgoja build aliases the module as `markdown` in the `main` runtime.

## Functions

### parse(input)

Parses a Markdown string and returns a Go-backed `MarkdownNode` document.

```js
const ast = markdown.parse("# Title\n\nSee [docs](https://example.com).");
console.log(ast.Type);            // "document"
console.log(ast.Children[0].Type); // "heading"
```

### renderHTML(input)

Renders a Markdown string to HTML through goldmark.

```js
console.log(markdown.renderHTML("**strong**"));
```

### walk(root, visitor, options?)

Traverses a `MarkdownNode` tree depth-first and calls `visitor(node, context)` for every visited node.

```js
markdown.walk(ast, (node, ctx) => {
  console.log(ctx.Depth, node.Type);
});
```

Visitor return values control traversal:

- `undefined` or `true` continues normally.
- `false` or `"skip"` skips the current node's children.
- `"stop"` stops traversal immediately.

### textContent(node)

Collects text below a node.

```js
const heading = ast.Children[0];
console.log(markdown.textContent(heading));
```

### validate(value)

Validates Markdown input or a Go-backed `MarkdownNode` value and returns a validation result.

Use this when scripts accept mixed user input and need a Go-side type check before continuing.

### builder()

Creates a Go-backed fluent Markdown document builder. Use it when JavaScript wants to generate a Markdown document from runtime data without string concatenation or a template file.

```js
const result = markdown.builder()
  .Title("Sprint report")
  .Paragraph("Generated from structured data.")
  .Table()
    .Columns({ label: "Name", align: "left" }, { label: "Status", align: "right" })
    .Row("Parser", "done")
    .Row("Builder", "planned")
    .End()
  .Render();

console.log(result.Text);
console.log(result.Bytes);
```

Builder methods:

- `Title(text)` — add a level-1 heading.
- `Heading(level, text)` — add a level-1 through level-6 heading.
- `Paragraph(...parts)` / `Text(text)` — add escaped paragraph text and inline parts.
- `Raw(markdown)` — add a raw Markdown block. Use only for trusted Markdown.
- `ThematicBreak()` — add `---`.
- `Blockquote(body)` — add a quoted block.
- `Callout(kind, title, body?)` — add an Obsidian-style callout such as `> [!WARNING]`.
- `BulletList(items)` / `OrderedList(items, start?)` — add lists from JavaScript arrays.
- `Checklist(items)` — add `- [ ]` / `- [x]` items from strings or `{ text, checked }` objects.
- `CodeBlock(language, code)` — add a fenced code block; the renderer chooses a safe fence length.
- `Table()` — start a child table builder.
- `Validate()` — return `{ Valid, Errors }`.
- `Render()` — return `{ Text, Bytes, Blocks }`.
- `RenderString()` — return just the Markdown string.
- `RenderHTML()` — render the generated Markdown through goldmark.

### TableBuilder methods

`Table()` returns a child builder. Call `End()` to append the table and return to the parent document builder.

```js
const text = markdown.builder()
  .Table()
    .Columns("Name", { label: "Score", align: "right" })
    .Row("Ada", 42)
    .Row("Linus | Kernel", 100)
    .End()
  .RenderString();
```

Methods:

- `Columns(...columns)` — set headers. A column can be a string or `{ label, align }`.
- `Align(...alignments)` — update alignments after columns are set.
- `Row(...cells)` — append one row.
- `Rows(rows)` — append an array of row arrays.
- `End()` — append the table once and return the parent builder.

The table renderer escapes cell pipes as `\\|`, converts cell newlines to `<br>`, and validates that every row has the same width as the header.

### inline()

Creates explicit inline helper values for paragraphs, headings, and table cells.

```js
const i = markdown.inline();
const text = markdown.builder()
  .Paragraph("Run ", i.Code("go test ./..."), " and read ", i.Link("docs", "https://example.com"), ".")
  .RenderString();
```

Inline helpers:

- `Text(text)` — escaped text.
- `Raw(markdown)` — trusted raw inline Markdown.
- `Code(code)` — code span with safe backtick length.
- `Em(...parts)` — emphasis.
- `Strong(...parts)` — strong emphasis.
- `Link(text, url, title?)` — Markdown link.

## MarkdownNode fields

Common fields include:

- `Type` — normalized node type such as `document`, `heading`, `paragraph`, `text`, `link`, `image`, `codeBlock`, `fencedCodeBlock`, or `htmlBlock`.
- `Children` — child nodes.
- `Text` — text payload for text-like nodes.
- `Level` — heading level.
- `Destination` — link or image destination.
- `Title` — link or image title.
- `Alt` — image alternate text.
- `Language`, `Info` — fenced-code metadata.
- `StartLine`, `StartColumn` — 1-indexed source position for the node start when goldmark exposes it.
- `SourcePos` — compatibility/detail field containing the same `[line, column]` pair.

Prefer checking `node.Type` and then reading the fields that are meaningful for that type.
