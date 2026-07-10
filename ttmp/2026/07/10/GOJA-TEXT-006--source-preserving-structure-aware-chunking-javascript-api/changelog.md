# Changelog

## 2026-07-10

- Initial workspace created


## 2026-07-10

Mapped the current Markdown, module, provider, generated-host, help, and test architecture and wrote the initial intern implementation design.

### Related Files

- /home/manuel/workspaces/2026-07-10/goja-text-chunking/goja-text/pkg/markdown/module.go — Primary module API evidence


## 2026-07-10

Added exact Markdown byte/rune ranges and end positions with syntax-preserving Goldmark range tests.

### Related Files

- /home/manuel/workspaces/2026-07-10/goja-text-chunking/goja-text/pkg/markdown/source_ranges.go — Exact range implementation


## 2026-07-10

Implemented lossless line, paragraph, Markdown block, and Markdown section segmentation.

### Related Files

- pkg/chunking/segment_markdown.go — Structure-aware source partitions and heading paths


## 2026-07-10

Implemented greedy, caller-weighted, and recursive packing with explicit oversized diagnostics.

### Related Files

- pkg/chunking/pack.go — Budget and overlap invariants


## 2026-07-10

Added require("chunking"), strict JavaScript codecs, TypeScript descriptors, runtime tests, and xgoja provider registration.

### Related Files

- pkg/chunking/module.go — Native JavaScript API and TypeScript declaration
- pkg/xgoja/providers/text/text.go — Provider registration


## 2026-07-10

Added the generated chunking app surface: demo, fixture, jsverbs, embedded help, README, TypeScript output, and smoke targets.

### Related Files

- cmd/goja-text/jsverbs/chunking.js — Root-mounted exploration commands
- pkg/xgoja/providers/text/doc/chunking-user-guide.md — Operational tutorial


## 2026-07-10

Fixed oversized-span marking after a preceding chunk flush and corrected CRLF terminator ownership.

### Related Files

- pkg/chunking/pack.go — Oversized handling before fit logic
- pkg/chunking/segment_lines.go — CRLF terminator integrity


## 2026-07-10

Passed formatting, build, normal and standalone tests, lint, two fuzz targets, TypeScript, help, generated-host, demo, verb, and complete make check validation.

### Related Files

- Makefile — End-to-end generated application validation
- pkg/chunking/pack_test.go — Budget and fuzz regression coverage


## 2026-07-10

Finalized the intern implementation guide, complete diary, golden fixture, review handoff, file relations, and delivery-ready ticket bundle.

### Related Files

- ttmp/2026/07/10/GOJA-TEXT-006--source-preserving-structure-aware-chunking-javascript-api/design-doc/01-source-preserving-chunking-architecture-and-implementation-guide.md — Intern architecture and implementation guide


## 2026-07-10

Implementation, documentation, validation, and reMarkable delivery bundle complete.
