---
Title: Source-Preserving Structure-Aware Chunking JavaScript API
Ticket: GOJA-TEXT-006
Status: active
Topics:
    - goja
    - goja-bindings
    - markdown
    - native-modules
    - text-algorithms
    - xgoja
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources:
    - https://github.com/go-go-golems/goja-text/issues/9
Summary: Design and implementation workspace for a source-preserving chunking native module exposed to JavaScript and generated xgoja applications.
LastUpdated: 2026-07-10T13:09:36.073033733-04:00
WhatFor: Track the architecture, implementation, verification, and delivery of exact text spans, structural segmenters, budgeted packing, recursive fallback, and the require("chunking") API.
WhenToUse: Start here when implementing or reviewing GitHub issue 9, extending the chunking API, or validating source-coordinate and no-data-loss invariants.
---

# Source-Preserving Structure-Aware Chunking JavaScript API

## Overview

This ticket implements [goja-text issue #9](https://github.com/go-go-golems/goja-text/issues/9): a new `require("chunking")` native module for deterministic, source-preserving segmentation and packing. The work extends Markdown AST coordinates, adds generic line/paragraph/Markdown segmenters, implements budgeted and recursive packing, exposes TypeScript and xgoja metadata, and documents the complete API for a new engineer.

The defining invariant is losslessness: built-in segmenters must partition the original UTF-8 source exactly. Packing may duplicate complete source spans only through declared overlap; it must never silently drop text or invent citation ranges.

## Key Links

- [Intern architecture and implementation guide](./design-doc/01-source-preserving-chunking-architecture-and-implementation-guide.md)
- [Chronological implementation diary](./reference/01-chunking-implementation-diary.md)
- [Task checklist](./tasks.md)
- [Implementation changelog](./changelog.md)
- [GitHub issue #9](https://github.com/go-go-golems/goja-text/issues/9)

## Status

Current status: **active**

The ticket and eight implementation tasks exist. Repository architecture and a clean `go test ./... -count=1` baseline have been captured; implementation is in progress.

## Topics

- goja
- goja-bindings
- markdown
- native-modules
- text-algorithms
- xgoja

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- `design-doc/` — intern-facing architecture, API, algorithms, decisions, and implementation guide.
- `reference/` — chronological implementation diary with commands, failures, and review instructions.
- `scripts/` — all ticket-specific experiments and validation helpers.
- `tasks.md` and `changelog.md` — completion state and implementation checkpoints.
