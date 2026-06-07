---
Title: "goja-text Markdown builder API reference"
Slug: goja-text-markdown-builder-api-reference
Short: "Reference for generating Markdown documents with require(\"markdown\").builder()."
Topics:
- goja-text
- markdown
- markdown-builder
- javascript
- xgoja
- api-reference
Commands:
- goja-text
- goja-text eval
- goja-text run
- goja-text markdown builder-example
- goja-text markdown builder-examples
Flags:
- outputPath
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

Use the Markdown builder when JavaScript has structured data and needs to emit a clean Markdown document. The builder avoids brittle string concatenation by giving scripts typed operations for headings, paragraphs, lists, checklists, callouts, code blocks, raw blocks, tables, and inline formatting.

The builder is Go-backed. JavaScript calls exported Go methods such as `Title`, `Paragraph`, `Table`, `RenderString`, and `RenderHTML`. Rendered results expose exported Go fields such as `Text`, `Bytes`, and `Blocks`.

## Loading the API

Load the existing `markdown` module and create a builder. The builder is part of `require("markdown")`; there is no separate module name.

```js
const markdown = require("markdown");

const text = markdown.builder()
  .Title("Status report")
  .Paragraph("Generated from structured data.")
  .RenderString();
```

Use this API when the document shape is assembled programmatically. Use the `template` module when the document shape is fixed and mostly described by a template file.

## Builder lifecycle

A `MarkdownBuilder` starts empty. Each block method appends one Markdown block and returns the same builder so calls can be chained.

```js
const doc = markdown.builder()
  .Title("Release notes")
  .Paragraph("This release contains generated sections.")
  .Heading(2, "Changes")
  .BulletList(["Added builder API", "Added table rendering"]);

const result = doc.Render();
console.log(result.Text);
console.log(result.Bytes);
console.log(result.Blocks);
```

Call `Validate()` when a script wants to inspect errors without rendering. Call `Render()` for a structured result, `RenderString()` for only Markdown text, and `RenderHTML()` to pass the generated Markdown through goldmark.

## MarkdownBuilder methods

### Title(text)

Adds a level-one heading. This is equivalent to `Heading(1, text)`.

```js
markdown.builder().Title("Project report");
```

### Heading(level, text)

Adds a heading from level 1 through 6. Invalid levels are recorded as validation errors and fail rendering.

```js
markdown.builder()
  .Heading(2, "Implementation")
  .Heading(3, "Tests");
```

### Paragraph(...parts) and Text(text)

Adds a paragraph. Ordinary strings are escaped as Markdown text. Multiple parts are concatenated, so inline helper values can be mixed with strings.

```js
const i = markdown.inline();

markdown.builder()
  .Paragraph("Run ", i.Code("go test ./..."), " before committing.");
```

`Text(text)` is convenience sugar for a one-part paragraph.

### Raw(markdown)

Adds trusted raw Markdown as a block. Use this for advanced Markdown that the builder does not model yet. Do not use it for untrusted user input.

```js
markdown.builder()
  .Raw("<details>\n<summary>More</summary>\n\nRaw Markdown here.\n</details>");
```

### ThematicBreak()

Adds a thematic break rendered as `---`.

```js
markdown.builder()
  .Paragraph("Before")
  .ThematicBreak()
  .Paragraph("After");
```

### Blockquote(body)

Adds a blockquote. String bodies are split on newlines and every line receives the `>` prefix.

```js
markdown.builder()
  .Blockquote("This text is quoted.\nSo is this line.");
```

### Callout(kind, title, body)

Adds an Obsidian-style callout block. The kind is uppercased in the marker.

```js
markdown.builder()
  .Callout("warning", "Review needed", "Check table escaping before release.");
```

Output:

```markdown
> [!WARNING] Review needed
> Check table escaping before release.
```

### BulletList(items)

Adds a bullet list from a JavaScript array.

```js
markdown.builder()
  .BulletList(["Parse data", "Build report", "Upload PDF"]);
```

### OrderedList(items, start?)

Adds an ordered list. The optional start value defaults to `1` and must be positive.

```js
markdown.builder()
  .OrderedList(["Design", "Implement", "Review"], 1);
```

### Checklist(items)

Adds a GitHub-style task list. Items may be strings or objects with `text` and `checked` fields.

```js
markdown.builder()
  .Checklist([
    { text: "Service layer", checked: true },
    { text: "CLI docs", checked: false },
  ]);
```

### CodeBlock(language, code)

Adds a fenced code block. The renderer chooses a fence long enough to contain code that itself includes backticks.

```js
markdown.builder()
  .CodeBlock("js", "console.log('hello');");
```

The language name must not contain whitespace.

### Table()

Starts a `TableBuilder`. A table is appended only when `End()` is called.

```js
markdown.builder()
  .Table()
    .Columns("Name", "Status")
    .Row("Parser", "done")
    .Row("Builder", "planned")
    .End()
  .RenderString();
```

## TableBuilder methods

Tables are first-class because they are easy to get wrong by hand. The renderer escapes pipes in cells, turns cell newlines into `<br>`, pads columns for readable source output, and validates that every row has the same number of cells as the header.

### Columns(...columns)

Sets the table columns. Each column can be a string or an object with `label` and optional `align`.

```js
.Table()
  .Columns(
    { label: "Name", align: "left" },
    { label: "Score", align: "right" },
    "Notes"
  )
```

Supported alignments are `default`, `left`, `center`, and `right`.

### Align(...alignments)

Updates alignments after columns are set.

```js
.Table()
  .Columns("Name", "Score")
  .Align("left", "right")
```

### Row(...cells)

Adds one row. Cells can be strings, numbers, booleans, inline helper values, or arrays of inline parts.

```js
const i = markdown.inline();

.Table()
  .Columns("Function", "Description")
  .Row(i.Code("builder()"), "Create a document builder")
```

### Rows(rows)

Adds multiple rows from an array of arrays.

```js
.Table()
  .Columns("Name", "Status")
  .Rows([
    ["Parser", "done"],
    ["Builder", "planned"],
  ])
```

### End()

Appends the table to the parent document and returns the `MarkdownBuilder`. Call `End()` once before continuing with more document blocks.

```js
markdown.builder()
  .Table()
    .Columns("Name", "Status")
    .Row("Parser", "done")
    .End()
  .Heading(2, "After the table");
```

Calling table methods after `End()` records validation errors.

## InlineFactory methods

Call `markdown.inline()` to create explicit inline nodes. Use inline helpers when ordinary escaped strings are not enough.

```js
const i = markdown.inline();
```

### Text(text)

Creates escaped text.

```js
i.Text("literal *stars*")
```

### Raw(markdown)

Creates trusted raw inline Markdown. Use sparingly and never for untrusted data.

```js
i.Raw("<kbd>Ctrl</kbd> + <kbd>C</kbd>")
```

### Code(code)

Creates a code span with safe backtick fencing.

```js
i.Code("go test ./...")
```

### Em(...parts) and Strong(...parts)

Creates emphasis and strong emphasis.

```js
markdown.builder()
  .Paragraph("Status: ", i.Strong("complete"));
```

### Link(text, url, title?)

Creates a Markdown link.

```js
i.Link("docs", "https://example.com/docs", "Documentation")
```

## Render results

`Render()` returns a Go-backed result object.

```js
const result = markdown.builder().Title("Report").Render();

console.log(result.Text);   // rendered Markdown
console.log(result.Bytes);  // byte length of Text
console.log(result.Blocks); // number of document blocks
```

`RenderString()` returns only `result.Text`. `RenderHTML()` renders the generated Markdown through the same goldmark path used by `markdown.renderHTML(input)`.

## Validation behavior

The builder accumulates structural errors and reports them during `Validate()`, `Render()`, `RenderString()`, or `RenderHTML()`.

Common validation failures include:

- heading level outside 1..6,
- table without columns,
- table row width mismatch,
- unknown table alignment,
- ordered-list start less than 1,
- code-block language containing whitespace,
- table mutation after `End()`.

Example:

```js
const result = markdown.builder()
  .Heading(9, "bad")
  .Table().Columns("A", "B").Row("only one cell").End()
  .Validate();

if (!result.Valid) {
  console.error(result.Errors.join("\n"));
}
```

## Complete example

This example generates a small report with a table, checklist, and inline code.

```js
const markdown = require("markdown");
const i = markdown.inline();

const rows = [
  { name: "Service layer", status: "done", owner: "Ada" },
  { name: "Goja module", status: "done", owner: "Linus" },
  { name: "Docs", status: "review", owner: "Intern" },
];

const table = markdown.builder()
  .Title("Markdown Builder Report")
  .Paragraph("Generated by ", i.Code('require("markdown").builder()'), ".")
  .Table()
    .Columns(
      { label: "Name", align: "left" },
      { label: "Status", align: "center" },
      { label: "Owner", align: "left" }
    );

for (const row of rows) {
  table.Row(row.name, row.status, row.owner);
}

const output = table.End()
  .Heading(2, "Next steps")
  .Checklist([
    { text: "Run tests", checked: true },
    { text: "Review generated Markdown", checked: false },
  ])
  .RenderString();

console.log(output);
```

## CLI examples

The generated binary bundles example data for the builder API.

```bash
./dist/goja-text markdown builder-examples
./dist/goja-text markdown builder-example report
./dist/goja-text markdown builder-example api-table
./dist/goja-text markdown builder-example report --output-path report.md
```

These commands are examples of using the JavaScript API from jsverbs. They are not required when writing normal scripts; scripts can call `require("markdown").builder()` directly.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `RenderString()` throws a validation error | The builder recorded an invalid heading, list, table, or code-block option | Call `Validate()` before rendering and inspect `Errors` |
| Table rows fail validation | A row has fewer or more cells than the header | Ensure every `Row()` has exactly the same cell count as `Columns()` |
| The table does not appear in output | `Table()` returns a child builder and `End()` was not called | Call `End()` once before rendering or adding more parent blocks |
| Pipes appear to break table cells | Raw Markdown was used inside table cells | Prefer ordinary strings or `inline().Code()` so the renderer can escape `|` |
| Markdown is visually wrapped in a table in CLI output | Glazed renders returned strings as a `value` field | Use `--output-path` or call the API from a script when raw Markdown output is needed |
| Lowercase fields such as `result.text` are undefined | Results are Go-backed objects | Use PascalCase fields such as `result.Text`, `result.Bytes`, and `result.Blocks` |

## See also

- `goja-text-markdown-api-reference` for parsing, traversal, validation, and HTML rendering.
- `goja-text-markdown-user-guide` for the broader Markdown module workflow.
- `goja-text-template-api-reference` when a fixed template file is a better fit than programmatic document assembly.
- `goja-text-template-writing-documentation` for template-based documentation generation.
