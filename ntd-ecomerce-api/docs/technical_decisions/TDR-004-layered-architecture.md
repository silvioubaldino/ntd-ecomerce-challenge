---
id: TDR-004
type: tdr
title: Layered architecture with a dedicated domain package
status: accepted
updated: 2026-07-08
parents: [AYD-001@context]
related: [TDR-001, TDR-002]
superseded_by: null
---

# TDR-004: Layered architecture with a dedicated domain package

## Context
The api needs a code organization that keeps handler/business/persistence concerns
separated, supports usecases depending on other usecases (e.g. purchase will need
product stock logic), and avoids Go import cycles.

## Decision
Layer-per-package (not package-per-feature), with entities in a leaf `domain`
package:

```
cmd/api/main.go                          orchestration only
internal/domain/                         entities + domain errors (imports nothing internal)
internal/usecase/                        business logic, one file per feature
internal/infrastructure/repository/      GORM models + repositories
internal/infrastructure/api/             Gin handlers + error mapping
internal/bootstrap/                      DI wiring (registry + per-feature setup)
migrations/                              SQL migrations (TDR-003)
```

Cycle avoidance rules:
- `internal/domain` is the only shared vocabulary; every layer imports it, it
  imports no other `internal` package.
- All usecases live in the single `package usecase`, so one usecase calling another
  can never cycle; the dependency is still declared as a **narrow interface in the
  consumer's file** and wired in `internal/bootstrap` — never a concrete struct.
- Handlers and repositories also depend on interfaces declared at point of use,
  keeping compile-time direction strictly `api → usecase → repository → domain`.

## Alternatives & trade-offs
- **Package-per-feature** (`internal/product/{handler,service,repository}.go`):
  reads well per feature but cross-feature service calls (purchase → product)
  create import cycles, forcing awkward extra packages — the exact problem this
  decision avoids.
- **Hexagonal / ports & adapters**: same testability via explicit ports, more
  ceremony (dedicated port packages) than a 4-feature MVP warrants; the
  interface-at-point-of-use rule gives the same decoupling.
- **Flat (main + handlers + store)**: fastest start, but degrades once CSV import,
  search, and purchase land.
