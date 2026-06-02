# Tasks

## Phase 0: Repository and module plumbing

- [x] Fix `goja-text/go.mod` module path to `github.com/go-go-golems/goja-text` and align Go version/toolchain with workspace expectations.
- [x] Add local dependency requirements for `github.com/go-go-golems/go-go-goja`, `github.com/dop251/goja`, and `github.com/yuin/goldmark`.
- [x] Remove or rename template placeholder command package `cmd/XXX` if it conflicts with module identity.

## Phase 1: Markdown domain model and parser

- [x] Create `pkg/markdown/types.go` with Go-backed `MarkdownNode`, `WalkContext`, `ValidationResult`, and option/result types.
- [x] Create `pkg/markdown/convert.go` to convert goldmark AST nodes into `*MarkdownNode` trees.
- [x] Implement `Parse(input string, options ...ParseOption) (*MarkdownNode, error)` as pure Go domain logic.
- [x] Implement `RenderHTML(input string, options ...RenderOption) (string, error)` as pure Go domain logic.
- [x] Implement `TextContent(node *MarkdownNode) (string, error)` as pure Go domain logic.
- [x] Implement `Validate(value any) ValidationResult` / `ValidateNode(node *MarkdownNode) ValidationResult` for Go-backed AST invariants.

## Phase 2: goja native module adapter

- [x] Create `pkg/markdown/module.go` implementing `modules.NativeModule` and `modules.TypeScriptDeclarer`.
- [x] Export `parse`, `renderHTML`, `walk`, `textContent`, and `validate`; do not export one-off heading/link extractors.
- [x] Implement `walk(root, visitor, options?)` using `goja.AssertFunction`, passing Go-backed nodes and `WalkContext` to JS.
- [x] Ensure JS examples and tests use Go field names: `node.Type`, `node.Children`, `node.Level`, etc.
- [x] Add TypeScript declaration metadata for `MarkdownNode`, `WalkContext`, `WalkResult`, and `ValidationResult`.

## Phase 3: Tests

- [x] Add pure Go parser/conversion tests for headings, paragraphs, links, lists, code blocks, and text content.
- [x] Add pure Go validation tests for nil/invalid nodes and heading-level invariants.
- [x] Add goja runtime integration tests proving `require("markdown")` works.
- [x] Add JS integration tests proving `parse()` returns Go-backed objects accessible as `node.Type` / `node.Children`.
- [x] Add JS integration tests proving heading/link-style queries can be implemented with `walk()`.
- [x] Run `go test ./... -count=1` from `goja-text`.
- [x] Run `GOWORK=off go test ./... -count=1` from `goja-text` if the local module can resolve dependencies independently.

## Phase 4: xgoja provider and binary

- [x] Create `pkg/xgoja/providers/text/text.go` that wraps the markdown `NativeModule` into xgoja `providerapi.Module` entries.
- [x] Create `xgoja.yaml` with the local `goja-text` provider (`replace: .`), `go-go-goja-core`, and guarded `go-go-goja-host` `fs` access.
- [x] Build the generated binary with `go run ../go-go-goja/cmd/xgoja build -f xgoja.yaml --xgoja-replace ../go-go-goja`.
- [x] Add `examples/js/markdown-demo.js` that reads a Markdown file using `require("fs")`, parses it, and walks headings/links using `walk()`.
- [x] Smoke-test `dist/goja-text eval`, `dist/goja-text run examples/js/markdown-demo.js`, and `dist/goja-text modules`.

## Phase 5: Documentation and ticket upkeep

- [x] Update README with JS usage, Go embedding, xgoja build, and smoke-test examples.
- [ ] Update the diary after each implementation step with commands, failures, commits, and validation results.
- [ ] Update ticket changelog and relate modified files after each meaningful commit.
- [x] Run `docmgr doctor --ticket GOJA-TEXT-001 --stale-after 30` before final handoff.
- [ ] Upload final design/diary bundle to reMarkable after implementation milestones.
