# Changelog

## 2026-06-08

- Initial workspace created


## 2026-06-08

Created design-first fluent document builder guide for goja-text point 7; implementation deferred for user review.

### Related Files

- /home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/ttmp/2026/06/08/goja-text-document-builder--fluent-markdown-document-builder-helpers/design-doc/01-fluent-document-builder-api-design-and-implementation-guide.md — Primary design and implementation guide
- /home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/ttmp/2026/06/08/goja-text-document-builder--fluent-markdown-document-builder-helpers/reference/01-diary.md — Diary recording the design-first pivot


## 2026-06-08

Implemented minimal Go-backed markdown.document builder without field-level frontmatter schema parsing; tests pass in normal and GOWORK=off modes.

### Related Files

- /home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/pkg/markdown/document.go — Builder/result implementation
- /home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/pkg/markdown/document_module_test.go — Goja integration coverage
- /home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/pkg/markdown/module.go — Module export and TypeScript declarations


## 2026-06-08

Committed minimal markdown.document builder implementation (commit 4cf73d7d0dcde4a1131bb5f2af3be3288250818c).

### Related Files

- /home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/pkg/markdown/document.go — Committed builder implementation
- /home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/pkg/markdown/document_module_test.go — Committed integration tests


## 2026-06-08

Refactored ClubMed slide and handout loaders to use markdown.document and committed the app change (commit 24513e0ecafa54486ae3b8df041175302c0054fd).

### Related Files

- /home/manuel/workspaces/2026-06-07/club-meetup-site/ClubMedMeetup/minitrace-viz/lib/handout-loader.js — Removed duplicated frontmatter/heading parsing
- /home/manuel/workspaces/2026-06-07/club-meetup-site/ClubMedMeetup/minitrace-viz/lib/slide-loader.js — Removed duplicated frontmatter/block parsing
- /home/manuel/workspaces/2026-06-07/club-meetup-site/ClubMedMeetup/minitrace-viz/xgoja.yaml — Local goja-text replace for xgoja build

