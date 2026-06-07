# goja-text

`goja-text` is a set of Go-backed text modules for [go-go-goja](../go-go-goja). It gives JavaScript scripts a practical way to parse Markdown, repair YAML and JSON, and extract structured-data snippets from larger documents while keeping the important domain logic in Go.

The project is designed for automation that sits at the boundary between prose and structure. That boundary shows up often: a Markdown file contains headings and links; a model response contains fenced JSON; a human-edited YAML file is almost valid but needs repair before downstream validation. `goja-text` provides the primitives for those cases without forcing every script to reimplement parsers in JavaScript.

## What the project provides

The repository currently exposes three JavaScript modules:

| Module | Purpose | Typical first call |
| --- | --- | --- |
| `markdown` | Parse Markdown, render HTML, walk a Go-backed Markdown AST, and collect text content. | `markdown.parse(input)` |
| `sanitize` | Repair and inspect YAML or JSON syntax, with fix metadata and Go-backed option builders. | `sanitize.json.sanitize(input)` |
| `extract` | Find structured-data candidates in Markdown, XML-like tags, frontmatter, or raw text. | `extract.all(input)` |
| `template` | Render Go `text/template` and `html/template` documents with Go-backed builders and Glazed/Sprig helpers. | `template.text().Parse(input)` |

These modules are available both to Go hosts using `go-go-goja` and to the generated `xgoja` binary described by `cmd/goja-text/xgoja.yaml`.

## Build the example xgoja binary

The `cmd/goja-text` directory is a committed generated xgoja command module. It contains the canonical `xgoja.yaml`, the bundled JavaScript verbs, generated `main.go`, embedded assets under `xgoja_embed/`, and its own `go.mod`/`go.sum` so the binary can be rebuilt without a temporary generation workspace.

From this repository directory:

```bash
make build-xgoja
```

For development after editing `cmd/goja-text/xgoja.yaml` or files under `cmd/goja-text/jsverbs`, regenerate the checked-in scaffold directly:

```bash
cd cmd/goja-text
GOWORK=off go generate
GOWORK=off go build -o ../../dist/goja-text .
```

The `go:generate` directive uses `go tool xgoja build --work-dir . --dry-run` to refresh the generated files in place, then runs a small post-generation normalizer and `go mod tidy`. The generated binary is written to `dist/goja-text` by the Makefile.

The build includes:

- `markdown`, `sanitize`, `extract`, and `template` from this repository.
- Core `path` and `yaml` modules from go-go-goja.
- Guarded host `fs` access for examples that read local files.
- Provider-shipped Glazed help pages for every goja-text module.
- Embedded jsverbs examples and practical commands under `cmd/goja-text/jsverbs`.

## Learn from the built-in help

The generated binary includes user-facing Glazed help entries. They are written as a pair for each module: an API reference for quick lookup and a user guide for learning the intended workflow.

```bash
./dist/goja-text help goja-text-markdown-user-guide
./dist/goja-text help goja-text-markdown-api-reference

./dist/goja-text help goja-text-sanitize-user-guide
./dist/goja-text help goja-text-sanitize-api-reference

./dist/goja-text help goja-text-extract-user-guide
./dist/goja-text help goja-text-extract-api-reference

./dist/goja-text help goja-text-template-user-guide
./dist/goja-text help goja-text-template-api-reference
```

The user guides include runnable examples, including `eval`, `run`, and bundled root-mounted JavaScript verb commands.

## Markdown: parse once, query with walk

The Markdown module uses goldmark in Go and exposes the parsed document as Go-backed `MarkdownNode` objects. JavaScript reads exported Go fields with PascalCase names.

```js
const markdown = require("markdown");

const ast = markdown.parse("# Hello\n\nSee [docs](https://example.com).");
console.log(ast.Type);                  // "document"
console.log(ast.Children[0].Type);       // "heading"
console.log(ast.Children[0].Level);      // 1
console.log(markdown.textContent(ast));  // "HelloSee docs."
```

Use `walk()` for document-specific queries:

```js
const headings = [];

markdown.walk(ast, (node, ctx) => {
  if (node.Type === "heading") {
    headings.push({
      level: node.Level,
      text: markdown.textContent(node),
      depth: ctx.Depth,
    });
  }
});
```

This is the central Markdown design choice. The Go API stays small and reliable, while JavaScript remains responsible for the question it wants to ask of the document.

## Sanitize: repair syntax before domain validation

The sanitize module exposes YAML and JSON namespaces backed by `github.com/go-go-golems/sanitize`.

```js
const sanitize = require("sanitize");

const options = sanitize.json.options()
  .MaxIterations(5)
  .Build();

const result = sanitize.json.sanitize("~~~json\n{'ok': True,}\n~~~\n", options);
console.log(result.Sanitized);
console.log(result.Fixes.map((fix) => fix.Rule));
```

Options are Go-backed builders rather than loose JavaScript objects. That lets Go reject unknown option names, validate values, and report better runtime errors when scripts import dynamic configuration.

## Extract: find candidates before parsing values

The extract module searches larger text for structured-data candidates. A candidate keeps the payload, the raw wrapper, source positions, format guesses, and confidence.

```js
const extract = require("extract");

const candidates = extract.all(`---
title: Demo
---

~~~json
{"ok": true}
~~~
`);

for (const candidate of candidates) {
  const validation = extract.validate(candidate);
  console.log(candidate.Kind, candidate.Format, candidate.StartRow, validation.Valid);
}
```

Extraction deliberately returns evidence rather than trusted parsed values. Scripts can validate, rank, display, or reject candidates according to their own policy.

## Demo scripts

The `examples/js` directory contains small scripts that use the host `fs` module to read fixtures from disk.

```bash
./dist/goja-text run examples/js/markdown-demo.js
./dist/goja-text run examples/js/sanitize-demo.js
./dist/goja-text run examples/js/extract-demo.js
./dist/goja-text run examples/js/template-demo.js
```

## Bundled verbs

The `cmd/goja-text/jsverbs` directory turns the same module patterns into Glazed commands. They are embedded into the generated binary by `cmd/goja-text/xgoja.yaml` and mounted at the generated root command with `commands.jsverbs.mount: root`, so the final binary can teach and exercise itself without reading verb files from disk or requiring an extra `verbs` prefix.

```bash
./dist/goja-text --help
./dist/goja-text examples tour
./dist/goja-text examples fixtures

./dist/goja-text markdown toc examples/markdown/sample.md
./dist/goja-text markdown links examples/markdown/sample.md
./dist/goja-text markdown summary examples/markdown/sample.md

./dist/goja-text sanitize yaml examples/yaml/broken.yaml
./dist/goja-text sanitize json examples/json/broken.json
./dist/goja-text sanitize lintJson examples/json/broken.json
./dist/goja-text sanitize rules json

./dist/goja-text extract list examples/text/structured-data-sample.md
./dist/goja-text extract validate examples/text/structured-data-sample.md
./dist/goja-text extract firstValid examples/text/structured-data-sample.md
./dist/goja-text extract markdownBlocks examples/text/structured-data-sample.md
```

Because jsverbs return structured values, Glazed can render the same command output in formats such as JSON, YAML, or tables.

## Validation

Use the repository Makefile for the normal validation path:

```bash
make check
```

The check target runs Go tests, builds the xgoja binary, and executes the smoke scripts for Markdown, sanitize, and extract.

You can also run the test commands directly:

```bash
go test ./... -count=1
GOWORK=off go test ./... -count=1
```

## Go embedding

A Go host can blank-import the package that registers a module and then build a go-go-goja runtime that exposes it.

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

## Design principles

- Keep parsing, repair, and extraction semantics in Go when Go will later validate or consume the values.
- Expose Go-backed domain objects to JavaScript rather than flattening everything into plain objects too early.
- Use JavaScript for document-specific policies and queries, especially with primitives such as `markdown.walk()` and `extract.all()`.
- Preserve evidence: source spans, wrapper metadata, fix lists, diagnostics, and confidence scores are part of the API because they make automation reviewable.
