# goja-text

`goja-text` provides Go-backed text algorithm modules for [go-go-goja](../go-go-goja). The first module is `require("markdown")`, which parses Markdown with goldmark and exposes the parsed AST as Go-backed objects inside JavaScript.

## Markdown module

The markdown module exports:

- `parse(input)` — parse Markdown into a Go-backed `MarkdownNode` tree
- `renderHTML(input)` — render Markdown to HTML
- `walk(root, visitor)` — traverse a `MarkdownNode` tree with a JavaScript callback
- `textContent(node)` — collect plain text from a node subtree
- `validate(value)` — validate Markdown input or a Go-backed node tree

Parsed nodes expose exported Go field names in JavaScript:

```js
const markdown = require("markdown");
const ast = markdown.parse("# Hello\n\nSee [docs](https://example.com).");

console.log(ast.Type);                  // "document"
console.log(ast.Children[0].Type);       // "heading"
console.log(ast.Children[0].Level);      // 1
console.log(markdown.textContent(ast));  // "HelloSee docs."
```

This is intentional: the project keeps domain AST values as Go-backed objects so Go module functions can validate them and report useful runtime errors when JavaScript passes invalid values back into Go.

## Querying with `walk`

The module does not export one-off helpers such as `extractHeadings` or `extractLinks`. Implement document-specific queries in JavaScript using `walk`:

```js
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
```

Visitor return values:

- `undefined` or `true`: continue normally
- `false` or `"skip"`: skip this node's children
- `"stop"`: stop traversal entirely

## xgoja binary

The included `xgoja.yaml` builds a `dist/goja-text` binary with:

- `markdown` from this repository
- core modules `path` and `yaml`
- guarded host `fs` access for reading files from disk

Build from the `goja-text` module directory:

```bash
go run ../go-go-goja/cmd/xgoja build \
  -f xgoja.yaml \
  --xgoja-replace /home/manuel/workspaces/2026-06-02/goja-text/go-go-goja
```

The `--xgoja-replace` path must be absolute because xgoja builds in a temporary directory.

Smoke tests:

```bash
./dist/goja-text modules --output json
./dist/goja-text eval 'const md = require("markdown"); const ast = md.parse("# Hello"); JSON.stringify({type: ast.Type, text: md.textContent(ast)})'
./dist/goja-text run examples/js/markdown-demo.js
```

## Go embedding

A host Go program can blank-import this package so its `init()` registers the module, then use the go-go-goja engine:

```go
package main

import (
  "context"
  "fmt"

  "github.com/dop251/goja"
  "github.com/go-go-golems/go-go-goja/engine"
  _ "github.com/go-go-golems/goja-text/pkg/markdown"
)

func main() {
  ctx := context.Background()
  factory, err := engine.NewBuilder().UseModuleMiddleware(engine.MiddlewareOnly("markdown")).Build()
  if err != nil { panic(err) }
  rt, err := factory.NewRuntime(engine.WithStartupContext(ctx), engine.WithLifetimeContext(ctx))
  if err != nil { panic(err) }
  defer rt.Close(ctx)

  ret, err := rt.Owner.Call(ctx, "example", func(_ context.Context, vm *goja.Runtime) (any, error) {
    v, err := vm.RunString(`const md = require("markdown"); md.textContent(md.parse("# Hello"));`)
    if err != nil { return nil, err }
    return v.Export(), nil
  })
  if err != nil { panic(err) }
  fmt.Println(ret)
}
```

## Tests

```bash
go test ./... -count=1
GOWORK=off go test ./... -count=1
```
