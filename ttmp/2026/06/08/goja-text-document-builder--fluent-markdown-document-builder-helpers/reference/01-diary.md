---
Title: Diary
Ticket: goja-text-document-builder
Status: active
Topics:
    - goja
    - modules
    - markdown
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ClubMedMeetup/ttmp/2026/06/08/xgoja-modules-improvement-second-edition--improve-xgoja-and-goja-modules-from-clubmedmeetup-usage-patterns-second-edition/design-doc/01-xgoja-and-goja-module-improvement-map-second-edition.md
      Note: Source design ticket identifying point 7 for goja-text document helpers
    - Path: goja-text/ttmp/2026/06/08/goja-text-document-builder--fluent-markdown-document-builder-helpers/design-doc/01-fluent-document-builder-api-design-and-implementation-guide.md
      Note: Design guide produced in the design-first step
    - Path: goja-text/ttmp/2026/06/08/goja-text-document-builder--fluent-markdown-document-builder-helpers/tasks.md
      Note: Task checklist for design review and later implementation
ExternalSources: []
Summary: Chronological diary for the design-first goja-text document builder ticket.
LastUpdated: 2026-06-08T18:15:00-04:00
WhatFor: Use this to resume or review the design-first document-builder workflow before any implementation begins.
WhenToUse: Read before implementing markdown.document() or refactoring ClubMedMeetup/minitrace-viz loaders.
---


# Diary

## Goal

This diary captures the design-first work for adding Go-backed fluent document-level helpers to `goja-text` and planning the later `ClubMedMeetup/minitrace-viz` loader refactor.

## Step 1: Create the goja-text ticket and pivot to design-first API review

I created a dedicated docmgr ticket under `goja-text/ttmp` using the requested `docmgr --root goja-text/ttmp ...` workflow, then wrote the design/implementation guide before making any code changes. The design centers on a fluent Go-backed `markdown.document(source)` builder so Go, not JavaScript option maps, owns validation and runtime invariants.

I briefly started sketching a concrete Go file too early, then stopped and removed that draft when the user clarified that the document must be reviewed first. The current repo state should therefore contain only the new ticket documentation, not an implementation of the document helper.

### Prompt Context

**User prompt (verbatim):** "Look at @ClubMedMeetup/ttmp/2026/06/08/xgoja-modules-improvement-second-edition--improve-xgoja-and-goja-modules-from-clubmedmeetup-usage-patterns-second-edition/ and the design ticket, and then tacakle point 7 for goja-text and document level helpers, and refactoring @ClubMedMeetup/minitrace-viz/ accordingly as well.

Create a docmgr ticket in goja-text/ttmp/ using the `docmgr --root goja-text/ttmp ...` flag, and Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Follow-up user prompt (verbatim):** "Go for an elegant fluent builder based API where the building is done on the go side to be able to enforce solid invariants at runtime as well, instead of relying on JS maps/objects"

**Follow-up user prompt (verbatim):** "oko but write the document / design / implementaiton guide first before implementing, I need to review it"

**Assistant interpretation:** Create a new goja-text docmgr ticket, study the existing second-edition ClubMed xgoja design ticket and current source files, then write a detailed intern-facing design and implementation guide for a Go-backed fluent Markdown document builder API before implementation.

**Inferred user intent:** The user wants to review and shape the API contract before code exists, especially to ensure the API is builder-based and invariant-preserving rather than a loose JavaScript object-map helper.

**Commit (code):** N/A — no code changes were intentionally kept.

### What I did

- Created ticket `goja-text-document-builder` under `goja-text/ttmp/2026/06/08/goja-text-document-builder--fluent-markdown-document-builder-helpers`.
- Added design doc `design-doc/01-fluent-document-builder-api-design-and-implementation-guide.md`.
- Added this diary doc `reference/01-diary.md`.
- Read the relevant point 7 section in the ClubMed second-edition design ticket.
- Inspected current `goja-text` module/provider files and the target `ClubMedMeetup/minitrace-viz/lib/slide-loader.js` and `handout-loader.js` files.
- Wrote a design that proposes `markdown.document(source)` plus nested Go-backed builders for frontmatter and structured blocks.
- Removed an early draft implementation file at `goja-text/pkg/markdown/document.go` after the user clarified design-first review.

### Why

- A separate goja-text ticket keeps the reusable module design independent from the ClubMed application ticket.
- The builder-first design directly addresses the user's request to enforce invariants on the Go side.
- Deferring implementation makes API review possible before compatibility and naming choices harden.

### What worked

- `docmgr --root goja-text/ttmp ticket create-ticket` created the requested ticket workspace.
- `docmgr --root goja-text/ttmp doc add` created the design doc and diary doc.
- Existing `goja-text` code already has strong precedents for Go-backed fluent builders: `markdown.builder()`, `extract.options()`, `sanitize.*.options()`, and `template.text()`.
- The ClubMed loaders provide clear evidence for a reusable document-level helper: duplicated frontmatter parsing, heading extraction, block extraction, JSON repair, and body stripping.

### What didn't work

- I initially started an implementation draft before the design was ready for review. I corrected this by deleting the draft file:
  - `rm -f goja-text/pkg/markdown/document.go`
- Running `git status --short` from the workspace root failed because the root is not itself a Git repository:
  - `fatal: not a git repository (or any of the parent directories): .git`
  - I then checked repository-specific status with `git -C goja-text status --short`.

### What I learned

- The existing `markdown` module intentionally exposes Go-backed objects with PascalCase methods/fields, so the new builder should follow that convention instead of inventing a lower-camel JavaScript object API.
- The first implementation slice can be much smaller than the full design: frontmatter builder, block extraction builder, parsed document methods, and typed frontmatter accessors are enough to refactor the ClubMed loaders.
- Field-schema builders are valuable but should probably be a second slice unless review says they are required immediately.

### What was tricky to build

- The main design tension is elegance versus scope. A nested fluent builder is more verbose than an options map, but it is the right shape for Go-side validation. The proposed solution keeps the full fluent shape in the design while identifying a minimal first implementation slice.
- Another tricky point is package ownership: the feature belongs in the `markdown` module, but it wants frontmatter extraction and YAML/JSON repair behavior that currently lives in `extract` and `sanitize`. The implementation guide calls out package-cycle risk and suggests shared internals if direct reuse creates a cycle.

### What warrants a second pair of eyes

- Review whether the proposed nested API shape is the desired level of fluency:
  - `.Frontmatter().YAML().Repair().Optional().End()`
  - `.Blocks().Block("context-window").FromXMLTag(...).FromFence(...).JSON().Repair().End().StripFromBody().End().End()`
- Review whether first-slice implementation should include field declaration builders or only typed `FrontmatterView` accessors.
- Review whether generic `JSONValue()` is acceptable for structured blocks or whether the first version should include stronger JSON shape validation.

### What should be done in the future

- After design review, implement the minimal `markdown.document()` builder in `goja-text` with tests first.
- Only after `goja-text` tests pass, refactor `ClubMedMeetup/minitrace-viz/lib/slide-loader.js` and `handout-loader.js`.
- Upload the reviewed/approved guide bundle to reMarkable if still desired.

### Code review instructions

- Start with `goja-text/ttmp/2026/06/08/goja-text-document-builder--fluent-markdown-document-builder-helpers/design-doc/01-fluent-document-builder-api-design-and-implementation-guide.md`.
- Check the proposed API contract, decision records, and review checklist before looking at any implementation.
- Confirm no implementation file such as `goja-text/pkg/markdown/document.go` is present from this step.
- Validate documentation with:
  - `docmgr --root goja-text/ttmp doctor --ticket goja-text-document-builder --stale-after 30`

### Technical details

Key source files that shaped the design:

- `/home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/pkg/markdown/module.go`
- `/home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/pkg/markdown/builder.go`
- `/home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/pkg/extract/frontmatter.go`
- `/home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/pkg/sanitize/module.go`
- `/home/manuel/workspaces/2026-06-07/club-meetup-site/goja-text/pkg/xgoja/providers/text/text.go`
- `/home/manuel/workspaces/2026-06-07/club-meetup-site/ClubMedMeetup/minitrace-viz/lib/slide-loader.js`
- `/home/manuel/workspaces/2026-06-07/club-meetup-site/ClubMedMeetup/minitrace-viz/lib/handout-loader.js`

## Step 2: Accept minimal builder scope and add implementation tasks

The user approved moving forward with the minimal builder slice and explicitly scoped out view-level/frontmatter field-schema parsing for now. I updated the ticket task list so the implementation can proceed one step at a time: tests first, then Go-backed builder/result types, then module wiring, then validation, then the ClubMed loader refactor if the local build strategy permits it.

This step is intentionally documentation-only. It records the accepted scope before any code changes so the later implementation commits can be reviewed against a concrete checklist.

### Prompt Context

**User prompt (verbatim):** "ok, let's build the one without the view-level frontmatter schema parsing for now. Add tasks to the ticket, then work on them one by one, committing at a ppropriate intervals, keeping a detailed diary as you work."

**Assistant interpretation:** Proceed with the minimal `markdown.document()` builder implementation, skip field-level/schema frontmatter parsing for the first slice, add concrete tasks to the ticket, commit work at sensible boundaries, and keep the diary current.

**Inferred user intent:** The user wants implementation to begin, but with disciplined ticket bookkeeping and a smaller API surface that can be validated before adding schema-driven field parsing.

**Commit (code):** Pending — this step prepares documentation/task state for the first commit.

### What I did

- Updated `tasks.md` with implementation tasks for the minimal builder.
- Kept field-level frontmatter schema parsing out of the immediate task list.
- Left the design document intact so it can continue to describe future schema-builder extensions, while the tasks constrain the first implementation slice.

### Why

- The task list now matches the approved scope.
- Committing the documentation/task plan before code gives reviewers a clean baseline.

### What worked

- The existing design already had a "minimal API for first implementation" section, so no design rewrite was needed before implementation.

### What didn't work

- N/A for this planning step.

### What I learned

- The first implementation should focus on parser-policy consolidation: frontmatter section parsing/repair, typed accessor fallback, structured block extraction/repair, body stripping, first heading, and HTML rendering.
- Field declarations such as `.Field("title").String().Optional()` should remain future work.

### What was tricky to build

- The tricky part is keeping the implementation small while preserving the fluent Go-backed shape. The task list now separates the minimal builder from the future field-schema extension so the first code pass does not grow too wide.

### What warrants a second pair of eyes

- Confirm that the ClubMed refactor should happen before a released goja-text version exists. `ClubMedMeetup/minitrace-viz/xgoja.yaml` currently selects `goja-text` without a local `replace`, so runtime validation of the refactor may require a temporary local build strategy.

### What should be done in the future

- Implement failing Goja tests first.
- Commit documentation/task baseline.
- Implement the minimal builder and module wiring.
- Decide whether the ClubMed refactor should be committed now or after tagging/replacing goja-text.

### Code review instructions

- Review `tasks.md` first to confirm the approved implementation scope.
- Then review later code commits against the task checklist.

### Technical details

- Minimal scope excludes frontmatter field/schema builder methods.
- Minimal scope includes typed accessors on a built `FrontmatterView`, for example `String(name, fallback)`.
