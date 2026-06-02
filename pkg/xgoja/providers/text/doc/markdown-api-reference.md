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

Use `require("markdown")` when JavaScript needs a Go-backed Markdown parser, HTML renderer, and AST traversal primitive.

The module keeps parsed nodes as Go objects. JavaScript therefore reads exported Go fields with PascalCase names such as `Type`, `Children`, `Level`, `Destination`, `Title`, and `Text`.

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
