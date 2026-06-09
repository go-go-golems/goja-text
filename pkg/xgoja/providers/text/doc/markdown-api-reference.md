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

Use `require("markdown")` when JavaScript needs a Go-backed Markdown parser, HTML renderer, AST traversal primitive, fluent Markdown document builder, or fluent document parser for frontmatter and structured blocks.

The module keeps parsed nodes, generated-document builders, and parsed document helpers as Go objects. JavaScript therefore reads exported Go fields and calls exported Go methods with PascalCase names such as `Type`, `Children`, `Level`, `Destination`, `Title`, `Text`, `Frontmatter`, `FirstHeading`, `RenderString`, and `Table`.

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

### document(source)

Creates a Go-backed fluent document parser for Markdown files that combine prose with leading YAML frontmatter and named structured blocks.

The document builder supports two frontmatter layers:

- permissive typed accessors on the built `FrontmatterView`, such as `String(name, fallback)`;
- optional strict top-level field schema rules, such as `.Field("title").String().Required().End()`.

```js
const doc = markdown.document(source)
  .Frontmatter()
    .YAML().Repair().Optional()
    .Field("title").String().Required().End()
    .Field("number").String().Optional().Default("01").End()
    .End()
  .Blocks()
    .Block("context-window")
      .FromXMLTag("context-window")
      .FromFence("context-window")
      .JSON().Repair().Optional().End()
      .StripFromBody()
      .End()
    .End()
  .Build();

const fm = doc.Frontmatter();
const title = fm.String("title", doc.FirstHeading("Untitled"));
const html = doc.RenderHTML();
const block = doc.Block("context-window");
const snapshot = block ? block.JSONValue() : null;
```

Document builder methods:

- `Frontmatter()` — start frontmatter configuration.
- `Blocks()` — start structured block configuration.
- `Validate()` — validate builder configuration.
- `Build()` — parse and validate the document, returning a `ParsedDocument`.

Frontmatter builder methods:

- `YAML()` — parse leading `---` frontmatter as YAML.
- `Repair()` — repair frontmatter YAML before parsing.
- `Optional()` / `Required()` — decide whether missing frontmatter is allowed.
- `Field(name)` — add a strict top-level field rule.
- `End()` — return to the document builder.

Frontmatter field builder methods:

- `String()`, `Number()`, `Bool()` — require the YAML value to have that parsed type when present.
- `Required()` — fail `Build()` if the field is missing. Empty strings count as missing for string fields.
- `Optional()` — allow the field to be absent.
- `Default(value)` — insert a default into `FrontmatterView` when the field is absent. The default must match the declared type.
- `End()` — return to the frontmatter builder.

Field schema rules are intentionally strict. For example, `.Field("published").Bool().Required()` rejects `published: "yes"` because YAML parsed that value as a string, not a boolean.

Structured block builder methods:

- `Blocks().Block(name)` — configure a named block rule.
- `FromXMLTag(tag)` — extract `<tag>...</tag>` payloads.
- `FromFence(info)` — extract fenced code blocks whose first info word matches `info`.
- `JSON().Repair().End()` — parse matching payloads as JSON, repairing common syntax issues first.
- `StripFromBody()` — remove matching blocks before `Body()`, `AST()`, `FirstHeading()`, and `RenderHTML()`.
- `Optional()` / `Required()` — decide whether a matching block is required.
- `End()` — return to the parent builder.

Parsed document methods:

- `Source()` — original source.
- `Body()` — body after frontmatter removal and configured block stripping.
- `AST()` — parsed Markdown AST for `Body()`.
- `Frontmatter()` — typed frontmatter view.
- `FirstHeading(fallback?)` — first Markdown heading text from the parsed body.
- `RenderHTML()` — body rendered through goldmark.
- `Blocks()` — all extracted blocks.
- `Block(name)` — first extracted block by name, or `null`.

Frontmatter view methods:

- `Has(name)` / `Value(name)` — presence and raw normalized value.
- `String(name, fallback?)`, `Number(name, fallback?)`, `Bool(name, fallback?)` — typed accessors with fallback.
- `Keys()` — stable key list.
- `ToObject()` — shallow object copy for escape-hatch use.

Document block methods:

- `Name()`, `Kind()`, `Text()`, `Raw()`, `StartByte()`, `EndByte()`.
- `JSONValue()` — parsed JSON value when the rule used `JSON()`, or strict on-demand JSON parsing otherwise.

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
