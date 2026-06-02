# Changelog

## 2026-06-02

- Initial workspace created


## 2026-06-02

Step 1: Created structured-data extraction ticket and intern-facing design guide for a new extract module covering Markdown code blocks, XML-like tags, raw JSON/YAML recognition, frontmatter, combined extraction, and sanitize-backed validation.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/ttmp/2026/06/02/GOJA-TEXT-003--goja-text-module-bindings-structured-data-extraction-helpers/design-doc/01-structured-data-extraction-helpers-design-and-implementation-guide.md — Primary design guide
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/ttmp/2026/06/02/GOJA-TEXT-003--goja-text-module-bindings-structured-data-extraction-helpers/reference/01-investigation-diary.md — Initial diary


## 2026-06-02

Step 2: Expanded GOJA-TEXT-003 into detailed phased implementation tasks and marked planning checklist task complete.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/ttmp/2026/06/02/GOJA-TEXT-003--goja-text-module-bindings-structured-data-extraction-helpers/reference/01-investigation-diary.md — Planning diary entry
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/ttmp/2026/06/02/GOJA-TEXT-003--goja-text-module-bindings-structured-data-extraction-helpers/tasks.md — Detailed phase/task checklist


## 2026-06-02

Step 3: Implemented extract package skeleton, source-position helpers, candidate/options types, builder validation, and tests; go test ./... -count=1 passes.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/options_test.go — Builder tests
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/positions.go — Line index and span helpers
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/types.go — Candidate/options types


## 2026-06-02

Step 4: Implemented Markdown fenced code block, XML-like tag, and YAML frontmatter extractors with wrapper tests; go test ./... -count=1 passes.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/frontmatter.go — Frontmatter extraction
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/markdown_fences.go — Markdown fence extraction
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/xml_tags.go — XML-like tag extraction


## 2026-06-02

Step 5: Implemented raw JSON/YAML recognition, sanitize-backed candidate validation, and combined All extraction with tests; go test ./... -count=1 passes.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/all.go — Combined extraction
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/raw.go — Raw recognition
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/validate.go — Candidate validation


## 2026-06-02

Step 6: Implemented extract NativeModule, TypeScript declarations, and JavaScript runtime tests; go test ./... -count=1 passes.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/module.go — NativeModule
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/extract/module_test.go — Runtime tests

