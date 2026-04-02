# Architectural Direction

## Overview

This repository represents the Go-native foundation for Bubblestudio and related terminal UI (TUI) applications.

The system is intentionally built around the Go ecosystem:
- Bubble Tea
- Bubbles
- Lip Gloss

It is not targeting Ink or other language/runtime stacks.

---

## Core Principle

All Go-side concerns belong to a single evolving architecture:

- runtime/export loading and rendering
- reusable primitives (list, form, modal, etc.)
- interaction patterns (wizard, dashboards, etc.)
- application shell/runtime behavior

These are not separate products. They are parts of the same Go foundation and must evolve toward convergence.

---

## Current State

There are currently two visible tracks:

1. Export/runtime path (PR #1, PR #2)
2. Primitive/shell/pattern path (PR #3)

These exist at different levels of maturity but are part of the same system.

---

## Direction

- `bubblestudio/` is the canonical Go runtime and UI foundation
- Exported TUIs and constructed applications (e.g., `tenantui`) should be built from this foundation
- The primitive/pattern layer represents the preferred long-term structure for UI behavior
- The existing export/runtime implementation must be fixed for correctness in the short term

Over time, these paths should converge rather than remain parallel implementations.

---

## Constraints

- Do not introduce alternate language runtimes
- Do not treat shell/primitives as throwaway demos
- Do not allow duplicate UI systems to evolve independently
- Prefer convergence over duplication

---

## Near-Term Priorities

1. Fix export/runtime correctness issues (e.g., .tui parsing, identifier generation)
2. Continue developing reusable primitives and patterns
3. Begin aligning runtime/export behavior with the cleaner primitive architecture

---

## Long-Term Goal

A unified Go-native TUI platform where:

- designs can be authored and exported
- exported applications run on a shared runtime foundation
- primitives and patterns are consistent across all generated and hand-built TUIs
