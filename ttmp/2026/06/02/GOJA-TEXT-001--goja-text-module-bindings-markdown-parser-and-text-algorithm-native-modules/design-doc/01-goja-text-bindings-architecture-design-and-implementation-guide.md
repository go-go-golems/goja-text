---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: go-go-goja/engine/factory.go
      Note: Engine builder/factory/runtime composition pipeline
    - Path: go-go-goja/modules/common.go
      Note: NativeModule interface and Registry - core module system
    - Path: go-go-goja/modules/uidsl/module.go
      Note: Reference for Go struct to JS object projection pattern
    - Path: go-go-goja/modules/yaml/yaml.go
      Note: Reference module implementation - closest pattern to follow
    - Path: go-go-goja/pkg/jsverbs/scan.go
      Note: jsverbs scanner - tree-sitter based JS function discovery
    - Path: go-go-goja/pkg/xgoja/app/factory.go
      Note: xgoja RuntimeFactory - creates engine.Runtime from provider modules
    - Path: go-go-goja/pkg/xgoja/providers/core/core.go
      Note: xgoja core provider - pattern for wrapping NativeModule into providerapi.Module
    - Path: go-go-goja/pkg/xgoja/providers/host/host.go
      Note: xgoja host provider - guarded fs/exec/database modules with config
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---









# Goja Text Bindings: Architecture, Design, and Implementation Guide

## Executive Summary

This document is a **complete intern-ready guide** for building native Go→JavaScript text algorithm bindings in the `goja-text` workspace. We start with a **Markdown parser module** (`require("markdown")`) that parses Markdown on the Go side using [yuin/goldmark](https://github.com/yuin/goldmark) and exposes the resulting abstract syntax tree (AST) as native Go objects that goja automatically projects into JavaScript. The guide covers every layer of the system you need to understand — from the goja VM itself, through the module registration system, the engine factory, the REPL, the jsverbs system, and finally the concrete implementation plan for the markdown module and its future siblings.

The `goja-text` repository lives at `/home/manuel/workspaces/2026-06-02/goja-text/` and is a Go workspace (`go.work`) containing four modules:

- **`go-go-goja/`** — the core JavaScript runtime, module system, REPL, and engine infrastructure
- **`goja-text/`** — the new module where we will add text algorithm native modules
- **`glazed/`** — the Glazed CLI framework (commands, schemas, structured output)
- **`sanitize/`** — an existing text-repair service that demonstrates tree-sitter-based text processing

Our new code lives in `goja-text/`, but it depends on and extends `go-go-goja/`.

---

## Problem Statement and Scope

We need a way to call high-performance Go text algorithms from JavaScript running inside goja. The first algorithm is a Markdown parser. Rather than implementing a Markdown parser in JavaScript (slow, incomplete) or using a C bindings approach (complex), we parse Markdown in Go using goldmark and project the resulting AST into JavaScript as plain Go objects that goja's value converter automatically maps to JS objects.

### What we want

- `const ast = require("markdown").parse("# Hello\nWorld")` — returns a tree of JS objects
- `require("markdown").renderHTML("# Hello")` — returns an HTML string
- `require("markdown").walk(ast, callback)` — lets JavaScript implement custom AST queries against Go-backed nodes
- `require("markdown").validate(astOrInput)` — validates Go-backed AST objects and returns valuable runtime diagnostics
- Future: `require("text/diff")`, `require("text/slug")`, `require("text/template")`, etc.

### What we do NOT want

- No JavaScript-side parsing — all heavy lifting in Go
- No custom serialization — let goja's native Go↔JS conversion do the work
- No new framework code — reuse the existing `modules.NativeModule` / `engine.Factory` patterns exactly

---

## Part 1: The goja Runtime — How Go and JavaScript Talk to Each Other

### What is goja?

[goja](https://github.com/dop251/goja) is a Go implementation of ECMAScript (JavaScript). It is **not** a V8 or Node.js binding — it is a pure Go interpreter. This means:

- You create a `*goja.Runtime` (the VM) from Go code
- You run JavaScript source strings through it (`vm.RunString("...")` or `vm.RunProgram(...)`)
- You set Go values into the JS environment (`vm.Set("name", goValue)`)
- You read JS values back from Go (`vm.Get("name").Export()`)
- **Go objects are projected into JS automatically** — a `map[string]any` becomes a plain JS object, a `[]any` becomes a JS array, a struct becomes a JS object with fields as properties

### The goja Value Conversion Pipeline

This is the single most important thing to understand about writing native modules. When you return a Go value from a function that goja calls, the runtime converts it:

| Go type | JavaScript type | Example |
|---|---|---|
| `string` | `string` | `"hello"` → `"hello"` |
| `int`, `int64`, `float64` | `number` | `42` → `42`, `3.14` → `3.14` |
| `bool` | `boolean` | `true` → `true` |
| `map[string]any` | `object` | `map[string]any{"a": 1}` → `{a: 1}` |
| `[]any` or `[]T` | `Array` | `[]string{"x","y"}` → `["x","y"]` |
| `nil` | `null` | `nil` → `null` |
| `struct` with exported fields | `object` | `type X struct{ Name string }` → `{Name: "..."}` |
| `func(...)` with Go signatures | `function` | `func(s string) int` → callable JS function |
| `error` return | thrown exception | `return nil, fmt.Errorf("bad")` → throws |

**Key insight**: When your module's Loader sets an export function like `func(input string) (*MarkdownNode, error)`, goja wraps it so that:
- The JS caller passes a string → goja converts it to `string`
- Your Go function returns a `*MarkdownNode` → goja projects exported Go fields into JS properties such as `node.Type` and `node.Children`
- Your Go function returns an `error` → goja throws it as a JS exception

This is the preferred pattern for this project when the returned value has domain behavior or invariants. Map-based results are still useful for generic data modules such as `yaml.parse(input)`, but the Markdown AST should remain a Go-backed object graph so Go code can validate nodes and produce helpful runtime errors.

### Pseudocode: How a Function Call Flows

```
JS side:                          Go side (module loader):
─────────                         ──────────────────────
const ast =                       modules.SetExport(exports, "markdown",
  require("markdown")               "parse", func(input string) (*MarkdownNode, error) {
    .parse("# Hello");                 var doc goldmark.AST
                                       // ... parse ...
    ast.Children[0].Type              return convertAST(source, doc.Root), nil
                                       // ^ returns *MarkdownNode
                                     })
                                  )

Flow:
1. JS calls require("markdown")
2. goja_nodejs/require finds the registered loader
3. Loader runs, sets exports.parse = Go function
4. JS calls exports.parse("# Hello")
5. goja converts "# Hello" → Go string
6. Go function runs, returns `*MarkdownNode`
7. goja projects exported Go fields into JS properties
8. JS receives a Go-backed object tree
```

---

## Part 2: The Module System — How Native Go Code Gets into JS

### The `modules.NativeModule` Interface

Every native module implements this interface (defined in `go-go-goja/modules/common.go`):

```go
type NativeModule interface {
    Name() string                                    // The require() name, e.g. "markdown"
    Doc() string                                     // Documentation string
    Loader(vm *goja.Runtime, moduleObj *goja.Object) // Populates moduleObj.exports
}
```

**File reference**: `go-go-goja/modules/common.go`

The `Loader` function receives the VM and a module object. You get the exports object and set functions/values on it:

```go
func (mod m) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)
    modules.SetExport(exports, mod.Name(), "parse", func(input string) (any, error) {
        // ... parse markdown, return Go objects ...
    })
}
```

### The `modules.SetExport` Helper

Defined in `go-go-goja/modules/exports.go`:

```go
func SetExport(exports settableObject, moduleName, name string, fn interface{}) {
    if err := exports.Set(name, fn); err != nil {
        log.Error().Str("module", moduleName).Str("export", name).Err(err).Msg("modules: failed to set export")
    }
}
```

This is a thin error-logging wrapper around `goja.Object.Set()`. Use it for every export to get consistent error reporting.

### Self-Registration via `init()`

Modules register themselves in the global `DefaultRegistry` during Go's `init()` phase:

```go
func init() {
    modules.Register(&m{})
}
```

This means **just importing the package** (even with a blank import `_ "..."`) makes the module available to `require()`. The engine's `runtime.go` has blank imports for all built-in modules:

**File reference**: `go-go-goja/engine/runtime.go` — lines with `_ "github.com/go-go-golems/go-go-goja/modules/..."`

### The `modules.TypeScriptDeclarer` Interface (Optional)

If you want your module to participate in `.d.ts` generation, implement:

```go
type TypeScriptDeclarer interface {
    TypeScriptModule() *spec.Module
}
```

**File reference**: `go-go-goja/modules/typing.go`, `go-go-goja/pkg/tsgen/spec/types.go`

This returns a structured description of the module's functions, parameters, and return types that the `gen-dts` command renders into TypeScript declarations. The `yaml` module has a full example:

```go
func (m) TypeScriptModule() *spec.Module {
    return &spec.Module{
        Name: "yaml",
        Functions: []spec.Function{
            {
                Name: "parse",
                Params: []spec.Param{
                    {Name: "input", Type: spec.String()},
                },
                Returns: spec.Any(),
            },
            // ...
        },
    }
}
```

---

## Part 3: The Engine — Runtime Lifecycle and Factory Pattern

### The Builder → Factory → Runtime Pipeline

The engine package (`go-go-goja/engine/`) provides an explicit, immutable composition pipeline for creating JavaScript runtimes:

```
engine.NewBuilder(opts...)
  .WithModules(specs...)          // add module registrations
  .UseModuleMiddleware(mw...)     // filter which modules are loaded
  .WithRuntimeInitializers(inits...)  // add post-setup hooks
  .Build() → *Factory             // immutable, reusable

factory.NewRuntime(opts...) → *Runtime  // creates a fresh VM
```

**File references**:
- `go-go-goja/engine/factory.go` — `FactoryBuilder`, `Factory`, `Build()`, `NewRuntime()`
- `go-go-goja/engine/runtime.go` — `Runtime` struct, `Close()`, lifecycle methods

### The `Runtime` Object

A `*engine.Runtime` bundles everything:

| Field | Type | Purpose |
|---|---|---|
| `VM` | `*goja.Runtime` | The JavaScript virtual machine |
| `Require` | `*require.RequireModule` | The `require()` implementation |
| `Loop` | `*eventloop.EventLoop` | Node.js-style event loop for async |
| `Owner` | `runtimeowner.RuntimeOwner` | Thread-safe VM access scheduler |
| `Values` | `map[string]any` | Runtime-scoped key-value store |
| `runtimeCtx` | `context.Context` | Lifetime context for cancellation |

### Module Middleware — Controlling What Gets Loaded

Middleware functions compose a pipeline that selects which modules from the default registry are available:

| Middleware | Behavior | Example |
|---|---|---|
| `MiddlewareSafe()` | Override — only data-safe modules | `crypto, events, path, time, timer` |
| `MiddlewareOnly(names...)` | Override — only named modules | `MiddlewareOnly("markdown", "yaml")` |
| `MiddlewareExclude(names...)` | Transform — remove from default set | `MiddlewareExclude("exec", "fs")` |
| `MiddlewareAdd(names...)` | Transform — add to default set | `MiddlewareAdd("markdown")` |

**File reference**: `go-go-goja/engine/module_middleware.go`

### How to Add a New Module to the Engine

There are two ways:

1. **Blank import in `engine/runtime.go`** — for built-in modules that should always be available when the engine package is imported
2. **`engine.NewBuilder().WithModules(spec)`** — for modules that are opt-in or live in external packages

For the `markdown` module in `goja-text/`, we will use approach 2: the `goja-text` binary imports `go-go-goja/engine` and then explicitly adds the markdown module spec.

---

## Part 4: The xgoja Build System — Declarative Binary Generation

### What is xgoja?

xgoja is a **declarative build system** for generating custom goja-powered binaries. Instead of writing a Go `main.go` by hand, you write an `xgoja.yaml` spec that declares:

- Which **provider packages** to include (each package bundles one or more modules)
- Which **runtime profiles** to create (each profile selects modules and their config)
- Which **commands** to enable (`eval`, `run`, `repl`, `jsverbs`)
- Optional **jsverbs sources**, **help sources**, and **embedded assets**

The `xgoja build` command reads this spec, generates a Go program that imports the selected providers, and compiles a self-contained binary. The generated binary is a full-featured CLI with `eval`, `run`, `repl` (TUI), and `jsverbs` subcommands — all configured with exactly the modules you declared.

**File references**:
- `go-go-goja/cmd/xgoja/root.go` — xgoja CLI root
- `go-go-goja/cmd/xgoja/cmd_build.go` — `build` command
- `go-go-goja/cmd/xgoja/internal/buildspec/spec.go` — YAML spec data model
- `go-go-goja/cmd/xgoja/internal/generate/main.go` — code generation entry point
- `go-go-goja/cmd/xgoja/internal/generate/templates/main.go.tmpl` — generated `main.go` template

### The Provider Model

An xgoja **provider package** is a Go package that calls `providerapi.Register()` to register modules, verb sources, help sources, and capabilities into a `providerapi.Registry`. Each provider package has a unique string ID (e.g. `go-go-goja-core`, `go-go-goja-host`).

**File references**:
- `go-go-goja/pkg/xgoja/providerapi/registry.go` — `Registry`, `Package`, `ResolveModule()`
- `go-go-goja/pkg/xgoja/providerapi/module.go` — `Module` struct, `ModuleFactory`, `ModuleContext`
- `go-go-goja/pkg/xgoja/providers/core/core.go` — core provider (path, yaml, crypto, time, timer, events)
- `go-go-goja/pkg/xgoja/providers/host/host.go` — host provider (guarded fs, exec, database)

#### Core Provider

The core provider (`go-go-goja-core`) exposes data-safe modules that don't touch the host system:

```go
// go-go-goja/pkg/xgoja/providers/core/core.go
func Register(registry *providerapi.Registry) error {
    entries := make([]providerapi.Entry, 0, len(coreModuleNames))
    for _, name := range coreModuleNames {
        mod := modules.GetModule(name)  // from modules.DefaultRegistry
        entries = append(entries, nativeModuleEntry(mod))
    }
    return registry.Package(PackageID, entries...)
}
```

Each entry is a `providerapi.Module` that wraps the existing `modules.NativeModule`:

```go
func nativeModuleEntry(mod modules.NativeModule) providerapi.Module {
    return providerapi.Module{
        Name:        mod.Name(),
        DefaultAs:   mod.Name(),
        Description: mod.Doc(),
        New: func(providerapi.ModuleContext) (require.ModuleLoader, error) {
            return mod.Loader, nil  // just returns the NativeModule's Loader
        },
    }
}
```

This is the **key bridge**: existing `NativeModule` implementations are automatically compatible with xgoja's provider system. You don't need to rewrite your module — you just wrap it in a `providerapi.Module` entry.

#### Host Provider

The host provider (`go-go-goja-host`) exposes modules that **can** touch the host system (filesystem, process execution, database). Each module requires **explicit opt-in** via config:

```yaml
# xgoja.yaml — enabling host filesystem access
runtimes:
  main:
    modules:
      - package: go-go-goja-host
        name: fs
        as: fs
        config:
          allow: true  # explicit opt-in required!
```

The `fs` module in the host provider uses the same `modules.NativeModule` but configured with `fsmod.New(fsmod.WithBackend(fsmod.OSBackend{}))` when `allow: true` is set. This is how you get full host filesystem access in your generated binary.

### The xgoja.yaml Spec

An `xgoja.yaml` file declares the complete binary configuration:

```yaml
# xgoja.yaml
name: my-text-tool
target:
  kind: xgoja
  output: dist/my-text-tool
packages:
  - id: go-go-goja-core
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/core
  - id: go-go-goja-host
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/host
runtimes:
  main:
    modules:
      - package: go-go-goja-core
        name: yaml
        as: yaml
      - package: go-go-goja-host
        name: fs
        as: fs
        config:
          allow: true
commands:
  eval:
    enabled: true
    runtime: main
  run:
    enabled: true
    runtime: main
  repl:
    enabled: true
    runtime: main
```

### What the Generated Binary Provides

When you build this spec, the resulting `my-text-tool` binary has:

- `my-text-tool eval 'const yaml = require("yaml"); ...'` — evaluate JS inline
- `my-text-tool run script.js` — run a JS file with module roots from the script's directory
- `my-text-tool repl` — full TUI REPL with the configured modules
- `my-text-tool modules` — list available modules

Inside any of these, `require("fs")` gives host filesystem access, `require("yaml")` gives YAML parsing, etc.

### Building with xgoja

```bash
# From go-go-goja root
xgoja build -f xgoja.yaml

# With local go-go-goja checkout (for development)
xgoja build -f xgoja.yaml --xgoja-replace /path/to/go-go-goja

# Dry run (validate spec, show plan)
xgoja build -f xgoja.yaml --dry-run
```

The build process:
1. Reads and validates `xgoja.yaml`
2. Generates `main.go`, `go.mod`, and embedded resources in a temp directory
3. Runs `go mod tidy` and `go build`
4. Copies the binary to the output path

---

## Part 5: The goja-repl — Interactive Testing (Alternative to xgoja)

The `goja-repl` command (`go-go-goja/cmd/goja-repl/`) is a full-featured JavaScript REPL with CLI, TUI, and JSON server modes. It can be used for quick interactive testing, but **xgoja is the preferred approach** for goja-text because it gives us a custom-named binary with our own module set without modifying the go-go-goja codebase.

**File references**:
- `go-go-goja/cmd/goja-repl/main.go` — entry point
- `go-go-goja/cmd/goja-repl/root.go` — Cobra root, session management
- `go-go-goja/cmd/goja-repl/cmd_eval.go` — `eval` command
- `go-go-goja/cmd/goja-repl/cmd_bindings.go` — `bindings` command (inspect what's in scope)
- `go-go-goja/cmd/goja-repl/cmd_run.go` — `run` command (execute a .js file)

### How to Test a Module in goja-repl (If You Prefer It Over xgoja)

```bash
# Start TUI REPL with specific modules enabled
go run ./cmd/goja-repl --enable-module markdown tui

# Or run a script file
go run ./cmd/goja-repl --enable-module markdown run test-markdown.js
```

The limitation is that `goja-repl` only loads modules registered in `modules.DefaultRegistry` — so you'd need a blank import of your module package somewhere that `goja-repl` imports. With xgoja, you avoid this constraint entirely.

### The replapi.App Layer

The `replapi` package (`go-go-goja/pkg/replapi/`) is the high-level API that the CLI, TUI, and JSON server all use. It wraps:

- `engine.Factory` — creates runtimes
- `replsession.Service` — manages live sessions
- `repldb.Store` — optional SQLite persistence

**File reference**: `go-go-goja/pkg/replapi/app.go`

### What is goja-repl?

The `goja-repl` command (`go-go-goja/cmd/goja-repl/`) is a full-featured JavaScript REPL with CLI, TUI, and JSON server modes. It is the primary tool for **interactive testing** of new modules.

**File references**:
- `go-go-goja/cmd/goja-repl/main.go` — entry point
- `go-go-goja/cmd/goja-repl/root.go` — Cobra root, session management
- `go-go-goja/cmd/goja-repl/cmd_eval.go` — `eval` command
- `go-go-goja/cmd/goja-repl/cmd_bindings.go` — `bindings` command (inspect what's in scope)
- `go-go-goja/cmd/goja-repl/cmd_run.go` — `run` command (execute a .js file)

### How to Test a New Module in the REPL

Once the markdown module is registered and the `goja-repl` binary is built with it, you can:

```bash
# Start TUI REPL with markdown module enabled
go run ./cmd/goja-repl --enable-module markdown tui

# Or run a script file
go run ./cmd/goja-repl --enable-module markdown run test-markdown.js

# Or evaluate a one-liner
go run ./cmd/goja-repl eval --session-id test1 --source 'const m = require("markdown"); console.log(m.parse("# hi"))'
```

### The replapi.App Layer

The `replapi` package (`go-go-goja/pkg/replapi/`) is the high-level API that the CLI, TUI, and JSON server all use. It wraps:

- `engine.Factory` — creates runtimes
- `replsession.Service` — manages live sessions
- `repldb.Store` — optional SQLite persistence

When you need to programmatically create a REPL with your module, you compose the engine factory with your module's middleware and pass it to `replapi.New()`.

**File reference**: `go-go-goja/pkg/replapi/app.go`

---

## Part 6: The jsverbs System — Declarative JavaScript Commands

### What are jsverbs?

jsverbs is a system that scans JavaScript source files for **top-level functions** and exposes them as **Glazed CLI commands**. A JavaScript file can declare verbs, sections, and field metadata using special comment-style annotations (`__package__`, `__verb__`, `__section__`), and the scanner builds a command tree from them.

**File references**:
- `go-go-goja/pkg/jsverbs/model.go` — data model (`Registry`, `VerbSpec`, `FileSpec`, etc.)
- `go-go-goja/pkg/jsverbs/scan.go` — tree-sitter-based scanner
- `go-go-goja/pkg/jsverbs/command.go` — Glazed command bridge
- `go-go-goja/pkg/jsverbs/runtime.go` — runtime invocation (creates VM, runs JS function)
- `go-go-goja/pkg/jsverbs/binding.go` — parameter binding plans

### How jsverbs Work (End to End)

```
1. Scan: jsverbs.ScanDir("examples/jsverbs/basic")
   → parses .js files with tree-sitter
   → extracts function signatures, __verb__, __section__ metadata
   → builds Registry of VerbSpecs

2. Build Commands: registry.Commands()
   → for each VerbSpec, builds a cmds.CommandDescription
   → with schema sections, fields, flags
   → wraps in Command/WriterCommand implementing GlazeCommand/WriterCommand

3. Invoke: command.RunIntoGlazeProcessor(ctx, values, processor)
   → builds engine.Factory + Runtime
   → require()'s the JS module
   → calls the JS function with arguments from parsed CLI flags
   → returns the JS function's return value
   → converts to Glazed rows
```

### How jsverbs Relate to Text Modules

jsverbs let you write JavaScript that **uses** native text modules and expose that JavaScript as a CLI command. For example:

```javascript
// examples/text/markdown-stats.js
const markdown = require("markdown");

function stats(input) {
  const ast = markdown.parse(input);
  return {
    headings: ast.children.filter(n => n.type === "heading").length,
    paragraphs: ast.children.filter(n => n.type === "paragraph").length,
  };
}
```

After scanning, this becomes `goja-text markdown-stats --input "# Hello\nWorld"` — a fully wired CLI command.

---

## Part 7: The YAML Module — An Existing Pattern to Follow

The `yaml` module (`go-go-goja/modules/yaml/yaml.go`) is our closest reference implementation. It:

1. Defines a stateless struct `m` that implements `modules.NativeModule` and `modules.TypeScriptDeclarer`
2. Registers itself via `init()`
3. Exports three functions: `parse`, `stringify`, `validate`
4. Uses plain Go return types (`any`, `string`, `map[string]any`) that goja converts automatically

### Full YAML Module Implementation (Annotated)

```go
package yamlmod

type m struct{}
var _ modules.NativeModule = (*m)(nil)    // compile-time interface check
var _ modules.TypeScriptDeclarer = (*m)(nil)

func (m) Name() string { return "yaml" }

func (m) TypeScriptModule() *spec.Module { /* ... spec ... */ }

func (m) Doc() string { return `... documentation string ...` }

func (mod m) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)

    // parse(input: string) -> any | throws
    modules.SetExport(exports, mod.Name(), "parse", func(input string) (any, error) {
        var out any
        if err := yaml.Unmarshal([]byte(input), &out); err != nil {
            return nil, fmt.Errorf("yaml.parse: %w", err)
        }
        return out, nil
    })

    // stringify(value: any, options?) -> string
    modules.SetExport(exports, mod.Name(), "stringify", func(value any, options map[string]any) (string, error) {
        // ... encode with options.indent ...
    })

    // validate(input: string) -> { valid, errors? }
    modules.SetExport(exports, mod.Name(), "validate", func(input string) map[string]any {
        // ... decode, collect errors ...
    })
}

func init() { modules.Register(&m{}) }
```

**Key pattern**: Every export function returns Go types that goja can convert. `map[string]any` becomes a JS object, error returns become thrown exceptions, and `string` returns become JS strings. There is no manual JS object construction.

---

## Part 8: The ui.dsl Module — Exposing Go Structs as JS Objects

The `ui.dsl` module (`go-go-goja/modules/uidsl/`) is an important precedent because it shows how to expose **structured Go objects** (not just maps) as JS values. It defines Go struct types like `Element`, `Text`, `Document` and returns them from module functions. goja automatically projects the exported struct fields as JS properties.

**File references**:
- `go-go-goja/modules/uidsl/module.go` — Loader function, creates Element/Text/etc objects
- `go-go-goja/modules/uidsl/node.go` — Go struct types (Node interface, Element, Text, etc.)

### Pattern: Go Struct → JS Object

```go
type Element struct {
    Tag      string
    Attrs    []Attr
    Children []Node
}

// In the Loader:
exports.Set("div", func(call goja.FunctionCall) goja.Value {
    return vm.ToValue(&Element{Tag: "div", Attrs: attrs, Children: children})
})
```

When JS receives this `Element`, it can access `.Tag`, `.Attrs`, `.Children` as regular properties. This is exactly how we will expose markdown AST nodes — as Go structs with exported fields.

---

## Part 9: The goldmark Markdown Parser

### Why goldmark?

[yuin/goldmark](https://github.com/yuin/goldmark) is:
- The most widely used Go Markdown parser (used by Hugo)
- CommonMark compliant
- Extensible via transformers and plugins
- Already an **indirect dependency** of `go-go-goja` (via `glazed` → `glamour`)

It produces an AST that we can walk and convert to Go objects.

### goldmark Architecture

```
Input (string/bytes)
  → goldmark.New().Convert()        → HTML string (simple path)
  → goldmark.New().Parse()          → ast.Node tree (AST path)
  → goldmark.New().Parser().Parse(text.NewReader(source))
                                     → ast.Node (lower-level)
```

The AST nodes are defined in `github.com/yuin/goldmark/ast`:

| Node type | Go struct | Key fields |
|---|---|---|
| Document | `*ast.Document` | `BaseBlock`, meta |
| Heading | `*ast.Heading` | `Level int` |
| Paragraph | `*ast.Paragraph` | `BaseBlock` |
| TextBlock | `*ast.TextBlock` | `BaseBlock` |
| Text | `*ast.Text` | `Segment`, `Raw` |
| Emphasis | `*ast.Emphasis` | `Level int` |
| CodeBlock | `*ast.CodeBlock` | `BaseBlock` |
| FencedCodeBlock | `*ast.FencedCodeBlock` | `Language`, `Info` |
| Link | `*ast.Link` | `Destination`, `Title` |
| Image | `*ast.Image` | `Destination`, `Title` |
| List | `*ast.List` | `Ordered`, `Start`, `Marker` |
| ListItem | `*ast.ListItem` | `Offset` |
| Blockquote | `*ast.Blockquote` | `BaseBlock` |
| ThematicBreak | `*ast.ThematicBreak` | `BaseBlock` |
| HTMLBlock | `*ast.HTMLBlock` | `BaseBlock` |

### Walking the AST

```go
import (
    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/ast"
    "github.com/yuin/goldmark/text"
)

source := []byte("# Hello\n\nWorld\n")
md := goldmark.New()
doc := md.Parser().Parse(text.NewReader(source))

ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
    if entering {
        switch n.Kind() {
        case ast.KindHeading:
            h := n.(*ast.Heading)
            fmt.Printf("Heading level=%d\n", h.Level)
        case ast.KindText:
            t := n.(*ast.Text)
            fmt.Printf("Text: %s\n", string(t.Value(source)))
        }
    }
    return ast.WalkContinue, nil
})
```

---

## Part 10: Proposed Design — The Markdown Module

### Module Name and Exports

The module will be accessible as `require("markdown")` and expose a small set of primitives. We intentionally do **not** expose one-off query helpers as Go exports. Document-specific queries should be implemented in JavaScript with the general `walk` primitive while preserving Go-backed AST nodes.

| Export | Signature | Description |
|---|---|---|
| `parse(input)` | `parse(input: string, options?: ParseOptions) => MarkdownNode` | Parse Markdown, return a Go-backed AST root projected into JS |
| `renderHTML(input)` | `renderHTML(input: string, options?: RenderOptions) => string` | Parse + render to HTML in one call |
| `walk(root, visitor, options?)` | `walk(root: MarkdownNode, visitor: WalkVisitor, options?: WalkOptions) => void` | Traverse the Go-backed AST and call a JS visitor for every node |
| `textContent(node)` | `textContent(node: MarkdownNode) => string` | Extract plain text from a node subtree |
| `validate(value)` | `validate(value: string | MarkdownNode) => ValidationResult` | Validate parser input or a Go-backed AST object and return runtime diagnostics |

### AST Object Model

The parsed AST will be a tree of **Go structs** that goja projects into JS. This is intentional and important for the project's general goja style: Go-side objects preserve runtime type information, make builder-pattern validation possible, and let module functions produce precise runtime error messages when a JavaScript caller passes the wrong kind of object.

By default, goja exposes exported Go struct fields using their Go field names. Therefore JavaScript callers should expect `node.Type`, `node.Children`, `node.Text`, `node.Level`, and so on unless a runtime explicitly installs a `FieldNameMapper`. We are **not** designing the primary API around lowercase `node.type` / `node.children` map objects.

Each node has:

- `Type` — a string identifying the node kind (`"document"`, `"heading"`, `"paragraph"`, `"text"`, etc.)
- `Children` — an array of child nodes (for container nodes)
- `Text` — the inline text content (for leaf nodes), extracted from source
- Kind-specific fields like `Level` (heading), `Language` (code block), `Destination` (link)

The JSON tags are still useful for diagnostics, optional JSON export, and TypeScript documentation, but they are not the primary JS property naming mechanism.

#### Go Struct Definitions

```go
// MarkdownNode is a single node in the parsed Markdown AST.
// It is designed to be projected directly into JavaScript by goja.
type MarkdownNode struct {
    Type        string          `json:"type"`        // "document", "heading", "paragraph", "text", etc.
    Children    []*MarkdownNode `json:"children,omitempty"`
    Text        string          `json:"text,omitempty"`
    Level       int             `json:"level,omitempty"`
    Language    string          `json:"language,omitempty"`
    Destination string          `json:"destination,omitempty"`
    Title       string          `json:"title,omitempty"`
    Alt         string          `json:"alt,omitempty"`
    Ordered     bool            `json:"ordered,omitempty"`
    Start       int             `json:"start,omitempty"`
    Marker      string          `json:"marker,omitempty"`
    Info        string          `json:"info,omitempty"`
    Raw         string          `json:"raw,omitempty"`
    SourcePos   [2]int          `json:"sourcePos,omitempty"` // [line, col] of node start
}
```

#### JavaScript API Result Shape

```javascript
const ast = require("markdown").parse("# Hello\n\nWorld\n");

// ast is a Go-backed *MarkdownNode projected into JS.
// Use exported Go field names by default:
console.log(ast.Type);              // "document"
console.log(ast.Children[0].Type);  // "heading"
console.log(ast.Children[0].Level); // 1
console.log(ast.Children[0].Children[0].Text); // "Hello"
```

This is a deliberate part of the API. If callers want lowercase JSON-style objects for interchange, we can later add an explicit `toJSON(node)` / `toPlainObject(node)` helper. The default parse result remains a Go-backed object.

### Conversion Function: goldmark AST → MarkdownNode Tree

The core conversion walks the goldmark AST and builds `MarkdownNode` objects:

```go
func convertAST(source []byte, n ast.Node) *MarkdownNode {
    if n == nil {
        return nil
    }

    node := &MarkdownNode{
        Type: nodeType(n),
    }

    // Extract kind-specific fields
    switch n.Kind() {
    case ast.KindHeading:
        h := n.(*ast.Heading)
        node.Level = h.Level
    case ast.KindFencedCodeBlock:
        fcb := n.(*ast.FencedCodeBlock)
        if fcb.Info != nil {
            node.Language = string(fcb.Info.Value(source))
        }
    case ast.KindLink:
        l := n.(*ast.Link)
        node.Destination = string(l.Destination)
        node.Title = string(l.Title)
    case ast.KindImage:
        img := n.(*ast.Image)
        node.Destination = string(img.Destination)
        node.Title = string(img.Title)
        node.Alt = string(img.Title) // goldmark stores alt differently
    case ast.KindList:
        l := n.(*ast.List)
        node.Ordered = l.IsOrdered()
        if l.IsOrdered() {
            node.Start = l.Start
        }
        node.Marker = string(l.Marker)
    case ast.KindText:
        t := n.(*ast.Text)
        node.Text = string(t.Value(source))
    case ast.KindString:
        s := n.(*ast.String)
        node.Text = string(s.Value)
    case ast.KindHTMLBlock:
        // Collect raw HTML content
        node.Raw = collectHTMLBlock(source, n)
    case ast.KindCodeSpan:
        // Collect inline code content
        node.Text = collectCodeSpan(source, n)
    }

    // Walk children (for container nodes)
    child := n.FirstChild()
    for child != nil {
        node.Children = append(node.Children, convertAST(source, child))
        child = child.NextSibling()
    }

    return node
}

func nodeType(n ast.Node) string {
    switch n.Kind() {
    case ast.KindDocument:      return "document"
    case ast.KindHeading:       return "heading"
    case ast.KindParagraph:     return "paragraph"
    case ast.KindText:          return "text"
    case ast.KindString:        return "string"
    case ast.KindEmphasis:      return "emphasis"
    case ast.KindCodeSpan:      return "codeSpan"
    case ast.KindCodeBlock:     return "codeBlock"
    case ast.KindFencedCodeBlock: return "fencedCodeBlock"
    case ast.KindLink:          return "link"
    case ast.KindImage:         return "image"
    case ast.KindList:          return "list"
    case ast.KindListItem:      return "listItem"
    case ast.KindBlockquote:    return "blockquote"
    case ast.KindThematicBreak: return "thematicBreak"
    case ast.KindHTMLBlock:     return "htmlBlock"
    case ast.KindTextBlock:     return "textBlock"
    default:                    return string(n.Kind())
    }
}
```

### Module Implementation Skeleton

```go
package markdownmod

import (
    "fmt"

    "github.com/dop251/goja"
    "github.com/go-go-golems/go-go-goja/modules"
    "github.com/go-go-golems/go-go-goja/pkg/tsgen/spec"
    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/text"
)

type m struct{}
var _ modules.NativeModule = (*m)(nil)
var _ modules.TypeScriptDeclarer = (*m)(nil)

func (m) Name() string { return "markdown" }

func (m) Doc() string {
    return `
The markdown module provides Markdown parsing and rendering.

Functions:
  parse(input, options?): Parse Markdown into a Go-backed AST object tree.
  renderHTML(input, options?): Parse and render Markdown to HTML.
  walk(root, visitor, options?): Traverse a Go-backed AST using a JS callback.
  textContent(node): Extract plain text from a node subtree.
  validate(value): Validate parser input or Go-backed AST objects, return diagnostics.
`
}

func (m) TypeScriptModule() *spec.Module {
    return &spec.Module{
        Name: "markdown",
        RawDTS: []string{
            "export interface MarkdownNode {",
            "  Type: string;",
            "  Children?: MarkdownNode[];",
            "  Text?: string;",
            "  Level?: number;",
            "  Language?: string;",
            "  Destination?: string;",
            "  Title?: string;",
            "  Alt?: string;",
            "  Ordered?: boolean;",
            "  Start?: number;",
            "  Marker?: string;",
            "  Info?: string;",
            "  Raw?: string;",
            "  SourcePos?: [number, number];",
            "}",
            "export interface WalkContext {",
            "  Parent?: MarkdownNode;",
            "  Depth: number;",
            "  Index: number;",
            "  Path: number[];",
            "}",
            "export type WalkResult = void | boolean | 'skip' | 'stop';",
        },
        Functions: []spec.Function{
            {Name: "parse", Params: []spec.Param{{Name: "input", Type: spec.String()}}, Returns: spec.Named("MarkdownNode")},
            {Name: "renderHTML", Params: []spec.Param{{Name: "input", Type: spec.String()}}, Returns: spec.String()},
            {Name: "walk", Params: []spec.Param{{Name: "root", Type: spec.Named("MarkdownNode")}, {Name: "visitor", Type: spec.Any()}}, Returns: spec.Void()},
            {Name: "textContent", Params: []spec.Param{{Name: "node", Type: spec.Named("MarkdownNode")}}, Returns: spec.String()},
            {Name: "validate", Params: []spec.Param{{Name: "value", Type: spec.Any()}}, Returns: spec.Object(
                spec.Field{Name: "Valid", Type: spec.Boolean()},
                spec.Field{Name: "Errors", Type: spec.Array(spec.String()), Optional: true},
            )},
        },
    }
}

func (mod m) Loader(vm *goja.Runtime, moduleObj *goja.Object) {
    exports := moduleObj.Get("exports").(*goja.Object)

    // parse(input: string) => MarkdownNode tree
    modules.SetExport(exports, mod.Name(), "parse", func(input string) (any, error) {
        source := []byte(input)
        md := goldmark.New()
        doc := md.Parser().Parse(text.NewReader(source))
        return convertAST(source, doc), nil
    })

    // renderHTML(input: string) => string
    modules.SetExport(exports, mod.Name(), "renderHTML", func(input string) (string, error) {
        source := []byte(input)
        md := goldmark.New()
        var buf bytes.Buffer
        if err := md.Convert(source, &buf); err != nil {
            return "", fmt.Errorf("markdown.renderHTML: %w", err)
        }
        return buf.String(), nil
    })

    // walk(root: MarkdownNode, visitor: function, options?) => void
    modules.SetExport(exports, mod.Name(), "walk", func(root *MarkdownNode, visitor goja.Value, options ...map[string]any) error {
        fn, ok := goja.AssertFunction(visitor)
        if !ok {
            return fmt.Errorf("markdown.walk: visitor must be a function")
        }
        state := walkState{Stopped: false}
        return walkMarkdownNode(vm, root, nil, fn, 0, 0, nil, &state)
    })

    // textContent(node: MarkdownNode) => string
    modules.SetExport(exports, mod.Name(), "textContent", func(node *MarkdownNode) (string, error) {
        if node == nil {
            return "", fmt.Errorf("markdown.textContent: node must be a MarkdownNode")
        }
        return collectText(node), nil
    })

    // validate(value: string | MarkdownNode) => ValidationResult
    modules.SetExport(exports, mod.Name(), "validate", func(value any) ValidationResult {
        switch v := value.(type) {
        case string:
            // Parsing markdown is permissive; this validates that conversion produces sane Go objects.
            source := []byte(v)
            md := goldmark.New()
            doc := md.Parser().Parse(text.NewReader(source))
            return validateMarkdownNode(convertAST(source, doc))
        case *MarkdownNode:
            return validateMarkdownNode(v)
        default:
            return ValidationResult{
                Valid: false,
                Errors: []string{fmt.Sprintf("markdown.validate: expected string or MarkdownNode, got %T", value)},
            }
        }
    })
}

func init() {
    modules.Register(&m{})
}
```

### Traversal Helper Pseudocode

`walk` keeps traversal on the Go side while letting JavaScript provide query logic. This preserves the project's Go-object pattern: JavaScript receives Go-backed `*MarkdownNode` values, and Go can validate the root object and report precise runtime errors.

```go
type WalkContext struct {
    Parent *MarkdownNode `json:"Parent,omitempty"`
    Depth  int           `json:"Depth"`
    Index  int           `json:"Index"`
    Path   []int         `json:"Path"`
}

type walkState struct {
    Stopped bool
}

type ValidationResult struct {
    Valid  bool     `json:"Valid"`
    Errors []string `json:"Errors,omitempty"`
}

func walkMarkdownNode(
    vm *goja.Runtime,
    node *MarkdownNode,
    parent *MarkdownNode,
    fn goja.Callable,
    depth int,
    index int,
    path []int,
    state *walkState,
) error {
    if state == nil || state.Stopped || node == nil {
        return nil
    }

    ctx := WalkContext{Parent: parent, Depth: depth, Index: index, Path: append([]int(nil), path...)}
    result, err := fn(goja.Undefined(), vm.ToValue(node), vm.ToValue(ctx))
    if err != nil {
        return err
    }

    skipChildren := false
    if result != nil && !goja.IsUndefined(result) && !goja.IsNull(result) {
        switch v := result.Export().(type) {
        case bool:
            skipChildren = !v
        case string:
            switch v {
            case "skip":
                skipChildren = true
            case "stop":
                state.Stopped = true
                return nil
            default:
                return fmt.Errorf("markdown.walk: unsupported visitor return string %q", v)
            }
        }
    }

    if skipChildren {
        return nil
    }
    for i, child := range node.Children {
        childPath := append(append([]int(nil), path...), i)
        if err := walkMarkdownNode(vm, child, node, fn, depth+1, i, childPath, state); err != nil {
            return err
        }
        if state.Stopped {
            return nil
        }
    }
    return nil
}

func collectText(node *MarkdownNode) string {
    if node == nil {
        return ""
    }
    var b strings.Builder
    var visit func(*MarkdownNode)
    visit = func(n *MarkdownNode) {
        if n == nil {
            return
        }
        if n.Text != "" {
            b.WriteString(n.Text)
        }
        for _, child := range n.Children {
            visit(child)
        }
    }
    visit(node)
    return b.String()
}

func validateMarkdownNode(root *MarkdownNode) ValidationResult {
    var errors []string
    var visit func(*MarkdownNode, string)
    visit = func(n *MarkdownNode, path string) {
        if n == nil {
            errors = append(errors, path+": nil child")
            return
        }
        if n.Type == "" {
            errors = append(errors, path+": Type is required")
        }
        if n.Type == "heading" && (n.Level < 1 || n.Level > 6) {
            errors = append(errors, fmt.Sprintf("%s: heading Level must be 1..6, got %d", path, n.Level))
        }
        for i, child := range n.Children {
            visit(child, fmt.Sprintf("%s.Children[%d]", path, i))
        }
    }
    visit(root, "root")
    return ValidationResult{Valid: len(errors) == 0, Errors: errors}
}
```

### JavaScript Query Examples Using `walk`

```javascript
const markdown = require("markdown");
const ast = markdown.parse(source);

function collectHeadingTexts(ast) {
  const headings = [];
  markdown.walk(ast, (node) => {
    if (node.Type === "heading") {
      headings.push({
        Level: node.Level,
        Text: markdown.textContent(node),
        Node: node,
      });
    }
  });
  return headings;
}

function collectLinks(ast) {
  const links = [];
  markdown.walk(ast, (node) => {
    if (node.Type === "link") {
      links.push({
        Destination: node.Destination,
        Title: node.Title || "",
        Text: markdown.textContent(node),
        Node: node,
      });
    }
  });
  return links;
}
```

These helpers are intentionally JavaScript examples, not Go module exports.

---

## Part 11: File Layout, Package Structure, and xgoja Spec

### Where the New Code Goes

Since `goja-text/` is a separate Go module in the workspace, the markdown module and its xgoja provider will live there:

```
goja-text/
├── go.mod                                  # github.com/go-go-golems/goja-text
├── xgoja.yaml                              # xgoja build spec for the goja-text binary
├── pkg/
│   ├── markdown/                           # The markdown module package
│   │   ├── markdown.go                     # NativeModule + Loader implementation
│   │   ├── convert.go                      # goldmark AST → MarkdownNode conversion
│   │   ├── types.go                        # MarkdownNode struct definition
│   │   ├── markdown_test.go                # Module-level tests
│   │   └── convert_test.go                 # Conversion tests
│   └── xgoja/
│       └── providers/
│           └── text/                        # xgoja provider for text modules
│               ├── text.go                 # Register() function, Module entries
│               └── doc.go                  # Package doc
├── examples/
│   └── js/
│       └── markdown-demo.js               # Demo script for the xgoja binary
└── verbs/                                  # jsverbs scripts (optional)
    └── markdown/
        └── stats.js                        # Example jsverbs script
```

### The xgoja Provider Package

The `goja-text/pkg/xgoja/providers/text/` package wraps the markdown `NativeModule` as a `providerapi.Module`, exactly like the core provider does for yaml/crypto/etc:

```go
// pkg/xgoja/providers/text/text.go
package text

import (
    "fmt"

    "github.com/dop251/goja_nodejs/require"
    "github.com/go-go-golems/go-go-goja/modules"
    _ "github.com/go-go-golems/goja-text/pkg/markdown" // triggers init() registration
    "github.com/go-go-golems/go-go-goja/pkg/xgoja/providerapi"
)

const PackageID = "goja-text"

var textModuleNames = []string{
    "markdown",
}

// Register exposes the text algorithm modules as an xgoja provider package.
func Register(registry *providerapi.Registry) error {
    entries := make([]providerapi.Entry, 0, len(textModuleNames))
    for _, name := range textModuleNames {
        mod := modules.GetModule(name)
        if mod == nil {
            return fmt.Errorf("text module %q is not registered", name)
        }
        entries = append(entries, nativeModuleEntry(mod))
    }
    return registry.Package(PackageID, entries...)
}

func nativeModuleEntry(mod modules.NativeModule) providerapi.Module {
    return providerapi.Module{
        Name:        mod.Name(),
        DefaultAs:   mod.Name(),
        Description: mod.Doc(),
        New: func(providerapi.ModuleContext) (require.ModuleLoader, error) {
            return mod.Loader, nil
        },
    }
}
```

**Key pattern**: The blank import `_ "github.com/go-go-golems/goja-text/pkg/markdown"` triggers the markdown package's `init()`, which calls `modules.Register(&m{})`. Then `nativeModuleEntry()` wraps the `NativeModule` into a `providerapi.Module` that xgoja can use. This is the **exact same pattern** as `go-go-goja/pkg/xgoja/providers/core/core.go`.

### The xgoja.yaml Build Spec

The `goja-text/xgoja.yaml` file declares our custom binary:

```yaml
# goja-text/xgoja.yaml
name: goja-text
target:
  kind: xgoja
  output: dist/goja-text
packages:
  - id: goja-text
    import: github.com/go-go-golems/goja-text/pkg/xgoja/providers/text
    replace: .
  - id: go-go-goja-core
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/core
  - id: go-go-goja-host
    import: github.com/go-go-golems/go-go-goja/pkg/xgoja/providers/host
runtimes:
  main:
    modules:
      # Text algorithm modules
      - package: goja-text
        name: markdown
        as: markdown
      # Core data modules
      - package: go-go-goja-core
        name: yaml
        as: yaml
      - package: go-go-goja-core
        name: path
        as: path
      # Host filesystem access (for loading .md files from disk)
      - package: go-go-goja-host
        name: fs
        as: fs
        config:
          allow: true
commands:
  eval:
    enabled: true
    runtime: main
  run:
    enabled: true
    runtime: main
  repl:
    enabled: true
    runtime: main
  jsverbs:
    enabled: true
    runtime: main
```

This spec gives us a `goja-text` binary with:
- `goja-text eval 'const md = require("markdown"); md.parse("# Hi")'` — inline evaluation
- `goja-text run script.js` — run a JS file (with `require("markdown")`, `require("fs")`, `require("yaml")`)
- `goja-text repl` — interactive TUI with all modules loaded
- `goja-text verbs ...` — jsverbs commands from `.js` files in `verbs/`

### Building the goja-text Binary

```bash
# From the goja-text directory
cd goja-text

# Build with xgoja (from go-go-goja)
go run ../go-go-goja/cmd/xgoja build -f xgoja.yaml --xgoja-replace ../go-go-goja

# Or, if xgoja is already installed:
xgoja build -f xgoja.yaml --xgoja-replace ../go-go-goja

# Run it
dist/goja-text eval 'const md = require("markdown"); console.log(md.renderHTML("# Hello"))'

# TUI REPL
dist/goja-text repl

# Run a script that loads markdown from disk
dist/goja-text run examples/js/markdown-demo.js
```

### Why xgoja Instead of a Custom CLI?

| Approach | Pros | Cons |
|---|---|---|
| **xgoja** (chosen) | No hand-written `main.go`; automatic `eval`/`run`/`repl`/`verbs` commands; modular via provider packages; fs module for disk access | Requires `xgoja build` step; slightly more indirection |
| Custom CLI | Direct control; no code generation | Must hand-write all commands; more boilerplate; every new text module requires more wiring |
| goja-repl with flags | Quick for testing | Can't add modules from external Go packages without blank-importing them into go-go-goja itself |

The xgoja approach wins because:
1. **It keeps go-go-goja untouched** — all new code lives in `goja-text/`
2. **It gives us `require("fs")`** for loading markdown files from disk (via the host provider)
3. **It gives us a full REPL** for free (the `repl` command with TUI)
4. **It gives us jsverbs** for free (the `verbs` command)
5. **Adding future text modules is trivial** — just add them to the provider package and update `xgoja.yaml`

### go.mod Dependencies

The `goja-text/go.mod` needs to depend on `go-go-goja` and `goldmark`:

```
module github.com/go-go-golems/goja-text

go 1.26.1

require (
    github.com/go-go-golems/go-go-goja v0.0.0  // local workspace
    github.com/yuin/goldmark v1.8.2
)
```

The workspace `go.work` already has `use ./goja-text`, so the local module resolution works automatically.

---

## Part 12: Implementation Plan (Phased)

### Phase 1: Core Markdown Module (MVP)

**Goal**: `require("markdown").parse()` and `require("markdown").renderHTML()` working end-to-end via xgoja binary.

Steps:
1. Create `goja-text/pkg/markdown/types.go` — define `MarkdownNode` struct
2. Create `goja-text/pkg/markdown/convert.go` — implement `convertAST()` and `nodeType()`
3. Create `goja-text/pkg/markdown/markdown.go` — implement `NativeModule`, `TypeScriptDeclarer`, Loader with `parse` and `renderHTML`
4. Create `goja-text/pkg/xgoja/providers/text/text.go` — xgoja provider wrapping the markdown module
5. Create `goja-text/xgoja.yaml` — build spec declaring markdown + core + host modules
6. Build the binary: `go run ../go-go-goja/cmd/xgoja build -f xgoja.yaml --xgoja-replace ../go-go-goja`
7. Verify: `dist/goja-text eval 'const md = require("markdown"); console.log(md.parse("# Hi"))'`
8. Write Go tests in `markdown_test.go` — create engine runtime, call `require("markdown")`, exercise `parse()` and `renderHTML()`
9. Add `examples/js/markdown-demo.js` — demo script that uses `require("fs")` to load a `.md` file and `require("markdown")` to parse it

### Phase 2: Traversal and Runtime Validation

**Goal**: add `walk()`, `textContent()`, and `validate()` so JavaScript can implement document-specific queries while Go retains runtime type checking and precise error reporting.

Steps:
1. Implement `walk(root, visitor, options?)` — validate `root` is a `*MarkdownNode`, validate `visitor` is a JS function, traverse Go-side nodes, and report useful errors
2. Implement `textContent(node)` — validate node type and collect plain text recursively
3. Implement `validate(value)` — accept string or `*MarkdownNode`, verify Go object invariants, and return `ValidationResult{Valid, Errors}`
4. Add TypeScript declarations for `MarkdownNode`, `WalkContext`, `WalkResult`, and `ValidationResult`
5. Add tests that prove JavaScript can implement heading/link extraction using `walk()` and `node.Type` / `node.Children`

### Phase 3: jsverbs Integration

**Goal**: JavaScript functions that use the markdown module become CLI commands via xgoja's built-in jsverbs support.

Steps:
1. Create `verbs/markdown/` directory with .js files
2. Add `jsverbs` source to `xgoja.yaml` (path + embed config)
3. Use `__verb__` and `__section__` annotations to declare commands in the JS files
4. Rebuild the binary — jsverbs commands appear automatically under `goja-text verbs ...`
5. Test end-to-end: `goja-text verbs markdown-stats --input "# Hello\nWorld"`

### Phase 4: Future Text Modules

After the markdown module proves the pattern, add:
- `require("text/diff")` — unified diff computation
- `require("text/slug")` — URL slug generation
- `require("text/template")` — Go text/template execution from JS
- `require("text/wordcount")` — word/character/sentence counting
- `require("text/outline")` — extract document outline from Markdown

Each follows the exact same pattern: `modules.NativeModule` + `init()` + Go return types.

---

## Part 13: Testing Strategy

### Unit Tests (Go)

Create a runtime directly, no subprocess:

```go
func TestMarkdownParse(t *testing.T) {
    ctx := context.Background()

    factory, err := engine.NewBuilder().
        UseModuleMiddleware(engine.MiddlewareOnly("markdown")).
        Build()
    require.NoError(t, err)

    rt, err := factory.NewRuntime(
        engine.WithStartupContext(ctx),
        engine.WithLifetimeContext(ctx),
    )
    require.NoError(t, err)
    defer rt.Close(ctx)

    v, err := rt.VM.RunString(`require("markdown").parse("# Hello")`)
    require.NoError(t, err)

    result := v.Export()
    node, ok := result.(*markdown.MarkdownNode)
    require.True(t, ok)
    require.Equal(t, "document", node.Type)
}
```

### Integration Tests (JS Scripts via xgoja binary)

Create `.js` files that exercise the full API using the xgoja-built binary:

```javascript
// examples/js/markdown-test.js
const md = require("markdown");
const fs = require("fs");

// Test parse: Go-backed struct fields use exported Go names.
const ast = md.parse("# Hello\n\nWorld\n");
if (ast.Type !== "document") throw new Error("expected document");
if (ast.Children[0].Type !== "heading") throw new Error("expected heading");
if (ast.Children[0].Level !== 1) throw new Error("expected level 1");

// Test renderHTML
const html = md.renderHTML("# Hello");
if (!html.includes("<h1>Hello</h1>")) throw new Error("expected h1 tag");

// Test file loading via fs module
const readmeContent = fs.readFileSync("./README.md", "utf-8");
const readmeAst = md.parse(readmeContent);
console.log("README has", readmeAst.Children.length, "top-level blocks");

// Test query composition using walk(), not Go-side extract helpers.
const headings = [];
md.walk(readmeAst, (node) => {
  if (node.Type === "heading") {
    headings.push({ Level: node.Level, Text: md.textContent(node) });
  }
});
console.log("README headings", headings);

console.log("All tests passed");
```

Run with: `dist/goja-text run examples/js/markdown-test.js`

### Manual Interactive Tests (via xgoja TUI REPL)

```bash
# Build the binary
xgoja build -f xgoja.yaml --xgoja-replace ../go-go-goja

# Start TUI REPL
dist/goja-text repl

# In the REPL:
js> const md = require("markdown")
js> md.parse("# Hello")
js> md.renderHTML("**bold**")
js> const fs = require("fs")
js> const content = fs.readFileSync("/tmp/test.md", "utf-8")
js> const ast = md.parse(content)
js> const headings = []
js> md.walk(ast, (node) => { if (node.Type === "heading") headings.push(md.textContent(node)) })
js> headings
```

---

## Part 14: Risks, Alternatives, and Open Questions

### Risks

| Risk | Mitigation |
|---|---|
| goldmark AST changes between versions | Pin goldmark version in go.mod; conversion function isolates us from internal changes |
| Go struct field projection has surprising JS property names | Document and test that the public API uses exported Go field names (`node.Type`, `node.Children`); only add a separate `toPlainObject()` helper if lowercase JSON-style objects are needed |
| Performance on large documents | Profile; add optional `maxDepth` / `maxNodes` limits to `parse()` |
| Module namespace conflicts with future text modules | Use `markdown` now; migrate to `text/markdown` when `text/` namespace is introduced |

### Alternatives Considered

1. **JSON serialization bridge** — Convert AST to JSON string, parse on JS side. Rejected: double serialization is wasteful and loses type information.
2. **JavaScript markdown parser** (marked.js, etc.) — Rejected: defeats the purpose of Go performance; also, these parsers are CommonMark-incomplete.
3. **goja custom Object construction** (vm.NewObject(), obj.Set()) — Rejected: more verbose than struct projection; struct projection is idiomatic and matches the uidsl pattern.
4. **Cgo-based parser** (cmark-gmc) — Rejected: unnecessary complexity; goldmark is pure Go and fast enough.

### Open Questions

1. **Should `MarkdownNode.SourcePos` use 1-indexed lines** (editor convention) or 0-indexed? Leaning toward 1-indexed for human familiarity.
2. **Should inline formatting nodes** (emphasis, strong, code span) be included in the default `parse()` output, or should there be a `parse({inlineDetail: true})` option? Leaning toward always including them — the tree is small enough.
3. **How to handle frontmatter** (YAML/TOML headers)? goldmark supports this via extensions. Should we enable it by default or make it opt-in?
4. ~~Should the module live in `go-go-goja/modules/markdown/` or `goja-text/pkg/markdown/`?~~ **Resolved**: `goja-text/pkg/markdown/` with an xgoja provider in `goja-text/pkg/xgoja/providers/text/`. This keeps go-go-goja untouched.

---

## Part 15: Key File Reference Index

| File | Purpose |
|---|---|
| `go-go-goja/modules/common.go` | `NativeModule` interface, `Registry`, `Register()` |
| `go-go-goja/modules/exports.go` | `SetExport()` helper |
| `go-go-goja/modules/typing.go` | `TypeScriptDeclarer` interface |
| `go-go-goja/modules/yaml/yaml.go` | **Reference module** — YAML parse/stringify/validate |
| `go-go-goja/modules/uidsl/module.go` | **Reference module** — Go struct → JS object projection |
| `go-go-goja/modules/uidsl/node.go` | Go struct types (Node, Element, Text, etc.) |
| `go-go-goja/engine/runtime.go` | `Runtime` struct, blank imports, lifecycle |
| `go-go-goja/engine/factory.go` | `FactoryBuilder`, `Factory`, `Build()`, `NewRuntime()` |
| `go-go-goja/engine/module_middleware.go` | Middleware pipeline (Safe, Only, Exclude, Add) |
| `go-go-goja/engine/module_specs.go` | `RuntimeModuleSpec`, `NativeModuleSpec` |
| `go-go-goja/engine/runtime_modules.go` | `RuntimeModuleContext`, `RuntimeModuleSpec` interface |
| `go-go-goja/cmd/xgoja/root.go` | xgoja CLI root (build, doctor, inspect, list-modules) |
| `go-go-goja/cmd/xgoja/cmd_build.go` | xgoja `build` command — reads spec, generates, compiles |
| `go-go-goja/cmd/xgoja/internal/buildspec/spec.go` | xgoja.yaml spec data model (Spec, Runtime, ModuleInstance, etc.) |
| `go-go-goja/cmd/xgoja/internal/generate/main.go` | Code generation entry point |
| `go-go-goja/cmd/xgoja/internal/generate/templates/main.go.tmpl` | Generated `main.go` template |
| `go-go-goja/pkg/xgoja/app/root.go` | Generated binary root command, eval/run/repl/verbs commands |
| `go-go-goja/pkg/xgoja/app/factory.go` | `RuntimeFactory` — creates engine.Runtime from provider modules |
| `go-go-goja/pkg/xgoja/app/host.go` | `Host` — attaches commands to cobra root |
| `go-go-goja/pkg/xgoja/app/spec.go` | Runtime spec types (mirrors buildspec for embedded JSON) |
| `go-go-goja/pkg/xgoja/providerapi/module.go` | `Module`, `ModuleFactory`, `ModuleContext` |
| `go-go-goja/pkg/xgoja/providerapi/registry.go` | `Registry`, `Package`, `ResolveModule()` |
| `go-go-goja/pkg/xgoja/providers/core/core.go` | Core provider — wraps yaml/crypto/path/time/timer/events |
| `go-go-goja/pkg/xgoja/providers/host/host.go` | Host provider — guarded fs/exec/database |
| `go-go-goja/pkg/xgoja/testprovider/provider.go` | Test fixture provider — shows full provider package pattern |
| `go-go-goja/cmd/goja-repl/root.go` | goja-repl CLI root, module flags, app construction |
| `go-go-goja/cmd/jsverbs-example/main.go` | jsverbs integration example |
| `go-go-goja/pkg/jsverbs/model.go` | jsverbs data model |
| `go-go-goja/pkg/jsverbs/scan.go` | tree-sitter scanner |
| `go-go-goja/pkg/jsverbs/runtime.go` | jsverbs invocation (creates runtime, runs JS) |
| `go-go-goja/pkg/replapi/app.go` | REPL application facade |
| `go-go-goja/pkg/tsgen/spec/types.go` | TypeScript declaration spec types |
| `go-go-goja/go.mod` | Module dependencies (includes goldmark indirect) |
| `go-go-goja/examples/xgoja/01-core-provider/xgoja.yaml` | Example: core provider only |
| `go-go-goja/examples/xgoja/02-host-provider/xgoja.yaml` | Example: host provider with fs/exec/database |

---

## Part 16: Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                    xgoja build pipeline                             │
│                                                                      │
│  xgoja.yaml ──▶ xgoja build ──▶ generated main.go + go.mod ──▶ go build ──▶ goja-text binary  │
│                                                                      │
│  packages:                       provider Register() calls:          │
│    goja-text           ──▶ text.Register()     ──▶ markdown module    │
│    go-go-goja-core     ──▶ core.Register()     ──▶ yaml, crypto, ...   │
│    go-go-goja-host     ──▶ host.Register()     ──▶ fs, exec, database  │
└──────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────┐
│                    goja-text binary (generated by xgoja)             │
│                                                                      │
│  Commands:                                                           │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐               │
│  │ eval     │ │ run      │ │ repl/TUI │ │ verbs    │               │
│  │ (inline) │ │ (.js)    │ │ (interac)│ │ (jsverbs)│               │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘               │
│       │             │            │             │                      │
│       └─────────────┴────────────┴─────────────┘                      │
│                           │                                          │
│                           ▼                                          │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │              RuntimeFactory (xgoja/app)                      │   │
│  │  Creates engine.Runtime with selected provider modules       │   │
│  │                                                              │   │
│  │  Runtime profile "main":                                    │   │
│  │    goja-text.markdown  ──▶ require("markdown")              │   │
│  │    go-go-goja-core.yaml ──▶ require("yaml")                │   │
│  │    go-go-goja-core.path ──▶ require("path")                │   │
│  │    go-go-goja-host.fs   ──▶ require("fs") [allow: true]    │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                           │                                          │
│                           ▼                                          │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                  engine.Runtime                              │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────────┐      │   │
│  │  │ VM       │ │ Require  │ │ Loop     │ │ Owner      │      │   │
│  │  │ (goja)   │ │ (nodejs) │ │ (event)  │ │ (runtime)  │      │   │
│  │  └────┬─────┘ └──────────┘ └──────────┘ └────────────┘      │   │
│  │       │                                                      │   │
│  │       ▼                                                      │   │
│  │  ┌────────────────────────────────────────────────────────┐  │   │
│  │  │  JavaScript Environment                               │  │   │
│  │  │  const md = require("markdown");                      │  │   │
│  │  │  const fs = require("fs");                            │  │   │
│  │  │  const src = fs.readFileSync("doc.md", "utf-8");      │  │   │
│  │  │  const ast = md.parse(src);                           │  │   │
│  │  │  // ast is a goja-projected MarkdownNode tree         │  │   │
│  │  └────────────────────────────────────────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  Go Side: goldmark → MarkdownNode conversion                 │   │
│  │                                                              │   │
│  │  source []byte                                               │   │
│  │    → goldmark.Parser().Parse(reader)                         │   │
│  │    → ast.Node tree                                           │   │
│  │    → convertAST(source, root)                                │   │
│  │    → *MarkdownNode tree                                      │   │
│  │    → goja projects into JS objects                           │   │
│  └──────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Part 17: Appendix — Existing Module Patterns at a Glance

### The Three Registration Paths

```
Path 1: Built-in (blank import in engine/runtime.go)
  _ "github.com/go-go-golems/go-go-goja/modules/yaml"
  → Available in ALL runtimes by default
  → Controlled by middleware (can exclude)
  → Only for modules inside go-go-goja itself

Path 2: xgoja Provider (providerapi.Module in provider package)
  providerapi.Module{Name: "markdown", New: func(ctx) (require.ModuleLoader, error) {...}}
  → Available in runtimes built from xgoja.yaml specs that list the module
  → Can pass config from xgoja.yaml into the module factory
  → No global state; fully declarative
  → THIS IS THE PATH WE USE for goja-text

Path 3: Plugin (hashiplugin SDK)
  → Out-of-process gRPC plugin
  → Separate binary, discovered at runtime
  → For heavy/isolated extensions

For the markdown module, we use Path 2:
the NativeModule's Loader is wrapped in a providerapi.Module
inside goja-text/pkg/xgoja/providers/text/text.go.
This is exactly how go-go-goja-core wraps yaml, crypto, etc.
```

### The NativeModule + xgoja Provider Implementation Checklist

When creating a new native module for xgoja, follow this checklist:

**In the module package** (`goja-text/pkg/markdown/`):
- [ ] Create package directory: `pkg/<name>/`
- [ ] Define the module struct: `type m struct{}`
- [ ] Add compile-time check: `var _ modules.NativeModule = (*m)(nil)`
- [ ] Implement `Name() string` — return the `require()` name
- [ ] Implement `Doc() string` — return documentation
- [ ] Implement `Loader(vm, moduleObj)` — set exports
- [ ] Optionally implement `TypeScriptDeclarer` for `.d.ts` generation
- [ ] Call `modules.Register(&m{})` in `init()`
- [ ] Write tests using `engine.NewBuilder().UseModuleMiddleware(engine.MiddlewareOnly("<name>")).Build()`

**In the provider package** (`goja-text/pkg/xgoja/providers/text/`):
- [ ] Add `nativeModuleEntry()` wrapper (same pattern as `core.go`)
- [ ] Add the module name to the `textModuleNames` list
- [ ] Add blank import of the module package to trigger `init()`
- [ ] Ensure `Register()` is called by the generated xgoja binary

**In the xgoja.yaml**:
- [ ] Add the provider package to `packages:`
- [ ] Add the module to the runtime profile's `modules:` list
- [ ] Rebuild the binary
