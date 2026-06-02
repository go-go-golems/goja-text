---
Title: "Investigation Diary"
Ticket: GOJA-TEXT-003
Status: active
Topics:
  - goja
  - goja-bindings
  - text-algorithms
  - native-modules
  - markdown
  - json
  - yaml
  - structured-data
  - xml
  - extraction
DocType: reference
Intent: "Chronological diary for structured-data extraction helper design and implementation"
Owners: []
RelatedFiles: []
---

# Investigation Diary

## Goal

Capture the design and implementation process for GOJA-TEXT-003: structured-data extraction helpers for code blocks, XML-like wrappers, raw JSON/YAML recognition, frontmatter, and related deterministic extraction primitives.

---

## Step 1: Create Structured Data Extraction Ticket and Design Guide

Created GOJA-TEXT-003 after closing GOJA-TEXT-002. The new ticket designs an `extract` module that locates structured payload candidates inside messy text while preserving source spans and wrapper metadata. The design intentionally separates extraction from parsing and repair: extraction finds candidates, while validation can use the existing sanitize package for JSON/YAML.

### Prompt Context

**User prompt (verbatim):** "do it. then close the ticket.

Then open a new ticket where we are going to provide helpers for extracting structured data from text:

- codeblocks from markdown
- xml tag wrapped 
- recognizing raw json / yaml
- other suggestions you might have.

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Finish GOJA-TEXT-002 by adding validation targets and closing it, then create a new docmgr ticket and intern-ready design guide for a structured-data extraction module.

**Inferred user intent:** Continue evolving goja-text from format-specific parsing/sanitizing modules into higher-level helpers for extracting structured payloads from unstructured text.

### What I did

- Added Makefile validation targets to GOJA-TEXT-002 and validated with `make check`.
- Closed GOJA-TEXT-002.
- Created GOJA-TEXT-003 with topics covering extraction, structured data, Markdown, XML, JSON, and YAML.
- Added primary design doc: `design-doc/01-structured-data-extraction-helpers-design-and-implementation-guide.md`.
- Added this diary document.
- Designed a new `require("extract")` module with helpers for:
  - Markdown fenced code blocks
  - XML-like tag wrappers
  - raw JSON/YAML recognition
  - YAML frontmatter
  - combined extraction
  - sanitize-backed validation
- Included decision records, algorithms, pseudocode, file layout, testing strategy, risks, open questions, and implementation checklist.

### Why

Structured data appears in many wrappers inside text. The Markdown and sanitize modules already parse and repair formats, but callers still need deterministic span-preserving extraction helpers. The new module should locate candidates and preserve provenance before validation or parsing.

### What worked

- The existing Markdown module provides a useful reference for codeblock semantics.
- The sanitize module provides validation and repair semantics for JSON/YAML candidates.
- The builder/config pattern from GOJA-TEXT-002 gives a clear options approach for `extract.options()`.

### What didn't work

- N/A — this step produced documentation only.

### What I learned

- Extraction should be treated as a separate responsibility from parsing and repair. Returning parsed values too early would discard source-span and wrapper metadata.

### What was tricky to build

- The main design challenge was avoiding overclaiming. XML-like tags should not be documented as full XML parsing, and raw YAML recognition must be conservative to avoid false positives.

### What warrants a second pair of eyes

- Whether `extract.all` should keep overlapping candidates by default or apply an overlap policy.
- Whether XML-like tag extraction should support nested same-name tags in Phase 1.
- Whether TOML/JSON frontmatter should be Phase 1 or deferred.

### What should be done in the future

- Implement the extraction package in phases, starting with source-position helpers and Markdown fenced code blocks.
- Add runtime tests before xgoja integration.

### Code review instructions

- Start with the decision records in the design doc.
- Then review extractor-specific algorithms for Markdown fences, XML-like tags, raw structured recognition, and frontmatter.
- Check that the implementation plan keeps domain extraction code separate from goja module wiring.

### Technical details

- Ticket path: `ttmp/2026/06/02/GOJA-TEXT-003--goja-text-module-bindings-structured-data-extraction-helpers/`
- Design doc: `design-doc/01-structured-data-extraction-helpers-design-and-implementation-guide.md`

---

## Step 2: Expand GOJA-TEXT-003 into Detailed Implementation Phases

Converted the initial high-level task list into detailed implementation phases. The ticket now has explicit tasks for planning, source-position infrastructure, candidate/options types, wrapper extractors, raw JSON/YAML recognition, sanitize-backed validation, native module wiring, xgoja integration, examples, README updates, validation commands, and reMarkable delivery.

### Prompt Context

**User prompt (verbatim):** "Now add a detailed set of phases + tasks for each phase to the ticket, then work on them one by one, committing at appropriate intervals, and keeping a detailed log."

**Assistant interpretation:** Replace the coarse GOJA-TEXT-003 task list with phase-level and subtask-level checkboxes, then proceed through implementation with commits and diary updates.

**Inferred user intent:** Make the implementation process reviewable and resumable, not a single large untracked coding pass.

### What I did

- Rewrote `tasks.md` with six phases:
  - planning and scaffolding
  - source positions, candidate types, and options
  - wrapper extractors
  - raw structured recognition and validation
  - native module and JavaScript runtime tests
  - xgoja integration, docs, validation, and delivery
- Marked task 0.1 complete.

### Why

The initial task list was useful but too coarse for step-by-step implementation. Detailed phases make commit boundaries and diary entries clearer.

### What worked

- The task list now maps directly to implementable files and validation checkpoints.

### What didn't work

- N/A — planning step only.

### What I learned

- The extraction work has enough surface area that separate commits for positions/types, wrapper extractors, raw/validation, and xgoja integration will make review easier.

### What was tricky to build

- The task split needed to keep extraction concerns separate from goja module wiring so domain logic can be tested independently.

### What warrants a second pair of eyes

- Whether raw YAML recognition should be implemented before or after wrapper extraction; the current plan puts it after wrapper extraction to reduce false-positive risk.

### What should be done in the future

- Implement Phase 1 next: line index helpers, candidate types, and options builder.

### Code review instructions

- Review `tasks.md` for implementation sequencing.

### Technical details

- Updated file: `tasks.md`
