# Tasks

## TODO

- [x] Add tasks here

- [x] Review existing goja-text native module, xgoja provider, docs, and test patterns
- [x] Review go-go-goja module and xgoja provider documentation
- [x] Review Glazed templating helper APIs and identify reusable pieces
- [x] Write intern-oriented template module implementation guide
- [x] Relate key source files and upload final bundle to reMarkable
- [x] Phase 1: Implement Go-backed template service types, builders, validation, function-set selection, and text/html rendering
- [x] Phase 1: Add service tests for text rendering, HTML escaping, named templates, helpers, missing-key errors, and builder validation
- [x] Phase 2: Add goja NativeModule adapter with text/html builders, convenience render helpers, docs string, and TypeScript declarations
- [x] Phase 2: Add runtime integration tests for require("template"), fluent builder rendering, HTML escaping, helpers, and validation errors
- [x] Phase 3: Wire the template module into the goja-text xgoja provider and generated command buildspec
- [x] Phase 3: Add template API reference, user guide, and examples/js/template-demo.js
- [x] Phase 4: Run go test ./... and GOWORK=off go test ./...; fix failures
- [x] Phase 4: Regenerate/build cmd/goja-text if provider or buildspec changes require it
- [x] Phase 5: Update diary, changelog, related files, docmgr doctor, and commit at coherent checkpoints
- [x] Future phase: Design and implement JS callback functions exposed to template FuncMap after runtime-owner review
- [x] Add Glazed help documentation for writing docs with the template API, including troubleshooting and see-also links
- [x] Add useful template jsverbs to the generated goja-text binary and update command tour/fixtures
- [x] Regenerate/build goja-text xgoja binary and smoke-test template jsverbs
