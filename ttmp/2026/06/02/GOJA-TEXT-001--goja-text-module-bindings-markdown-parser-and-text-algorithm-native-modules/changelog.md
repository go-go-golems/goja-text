# Changelog

## 2026-06-02

- Initial workspace created


## 2026-06-02

Step 1: Explored full go-go-goja architecture, created comprehensive intern-ready design document covering module system, engine factory, REPL, jsverbs, goldmark, and markdown module design. Created diary.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/ttmp/2026/06/02/GOJA-TEXT-001--goja-text-module-bindings-markdown-parser-and-text-algorithm-native-modules/design-doc/01-goja-text-bindings-architecture-design-and-implementation-guide.md — Main design document


## 2026-06-02

Uploaded design guide + diary bundle to reMarkable at /ai/2026/06/02/GOJA-TEXT-001


## 2026-06-02

Step 2: Researched xgoja build system, updated design document to use xgoja as primary testing vehicle. Added Part 4 (xgoja Build System), updated file layout, architecture diagram, implementation phases, and provider pattern. Will re-upload to remarkable.


## 2026-06-02

Re-uploaded updated design guide v2 + diary to reMarkable at /ai/2026/06/02/GOJA-TEXT-001


## 2026-06-02

Step 3: Added intern-facing review document critiquing the goja-text bindings plan/spec, including strengths, corrections, missing evidence, xgoja replace guidance, AST export risks, walk() recommendation, and next-time advice.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/ttmp/2026/06/02/GOJA-TEXT-001--goja-text-module-bindings-markdown-parser-and-text-algorithm-native-modules/design-doc/02-review-of-the-goja-text-bindings-plan-and-spec.md — Intern-facing technical review document


## 2026-06-02

Uploaded v3 bundle (original design + intern-facing review + diary) to reMarkable at /ai/2026/06/02/GOJA-TEXT-001.


## 2026-06-02

Step 4: Updated design to keep Markdown AST as Go-backed objects, use node.Type/node.Children JS access, remove one-off heading/link extraction exports, add walk/textContent/validate primitives, and add xgoja provider replace guidance.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/ttmp/2026/06/02/GOJA-TEXT-001--goja-text-module-bindings-markdown-parser-and-text-algorithm-native-modules/design-doc/01-goja-text-bindings-architecture-design-and-implementation-guide.md — Primary design updated for Go-backed AST and walk primitive


## 2026-06-02

Step 5: Implemented core markdown native module with Go-backed AST, parse/renderHTML/walk/textContent/validate exports, pure Go tests, and goja runtime integration tests.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/markdown/module.go — NativeModule implementation
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/markdown/module_test.go — Runtime integration coverage


## 2026-06-02

Step 6: Added xgoja provider/spec, demo markdown script using host fs, README docs, standalone GOWORK=off validation, and xgoja build/run smoke tests.

### Related Files

- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/pkg/xgoja/providers/text/text.go — provider registration
- /home/manuel/workspaces/2026-06-02/goja-text/goja-text/xgoja.yaml — xgoja build spec

