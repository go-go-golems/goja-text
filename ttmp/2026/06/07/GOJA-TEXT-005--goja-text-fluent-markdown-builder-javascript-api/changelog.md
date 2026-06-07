# Changelog

## 2026-06-07

- Initial workspace created


## 2026-06-07

Created GOJA-TEXT-005 design package for a fluent Go-backed Markdown builder API, including architecture evidence, API sketches, implementation phases, tests, and diary.

### Related Files

- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/ttmp/2026/06/07/GOJA-TEXT-005--goja-text-fluent-markdown-builder-javascript-api/design-doc/01-markdown-builder-analysis-design-and-implementation-guide.md — Primary intern-ready implementation guide
- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/ttmp/2026/06/07/GOJA-TEXT-005--goja-text-fluent-markdown-builder-javascript-api/reference/01-investigation-diary.md — Chronological investigation diary


## 2026-06-07

Validated GOJA-TEXT-005 with docmgr doctor and uploaded the documentation bundle to reMarkable at /ai/2026/06/07/GOJA-TEXT-005.

### Related Files

- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/ttmp/2026/06/07/GOJA-TEXT-005--goja-text-fluent-markdown-builder-javascript-api/design-doc/01-markdown-builder-analysis-design-and-implementation-guide.md — Uploaded as part of reMarkable bundle
- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/ttmp/2026/06/07/GOJA-TEXT-005--goja-text-fluent-markdown-builder-javascript-api/reference/01-investigation-diary.md — Uploaded as part of reMarkable bundle


## 2026-06-07

Implemented Phase 1 Markdown builder service layer with typed blocks/inlines, fluent builder methods, table rendering, escaping, validation, RenderHTML bridge, and service tests.

### Related Files

- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/builder.go — Fluent builder and validation
- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/builder_render.go — Markdown serializer
- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/builder_table.go — Table builder and inline factory
- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/builder_test.go — Service tests
- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/builder_types.go — Typed service model


## 2026-06-07

Implemented Phase 2 goja module exports for markdown.builder and markdown.inline, with runtime tests for fluent document generation, inline helpers, RenderHTML, and validation errors.

### Related Files

- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/module.go — NativeModule exports
- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/module_test.go — Runtime integration tests


## 2026-06-07

Updated Markdown TypeScript declarations and help pages for markdown.builder, TableBuilder, inline helpers, and generated Markdown workflows.

### Related Files

- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/markdown/module.go — TypeScript declarations
- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/xgoja/providers/text/doc/markdown-api-reference.md — API reference docs
- /home/manuel/workspaces/2026-06-07/goja-render-markdown/goja-text/pkg/xgoja/providers/text/doc/markdown-user-guide.md — User guide docs

