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

