---
id: CONV
type: conventions
title: Docs and code conventions
status: approved
updated: 2026-07-08
---

# Conventions

The "contract" that keeps docs and code consistent and readable by humans and AIs,
across the monorepo's three parts (`ntd-ecomerce-context`, `ntd-ecomerce-api`,
`ntd-ecomerce-web`). Two parts: **A) documentation** and **B) code**.

---

## A. Documentation

### A.1 Document types, IDs, and where they live
ID = `PREFIX-NNN`, **stable** (never changes, even if the file is renamed).

| Prefix | Type | Where | Scope |
|---------|------|------|--------|
| REQ  | Requirements | `ntd-ecomerce-context/docs/requirements.md` | shared |
| GLO  | Glossary (ubiquitous language) | `ntd-ecomerce-context/docs/requirements.md` (section) | shared |
| AYD  | Feature Analysis & Design | `ntd-ecomerce-context/docs/design/` | shared, cross-part |
| ARCH | Living architecture (C4) | `ntd-ecomerce-context/docs/architecture.md` | shared |
| CONV | These conventions | `ntd-ecomerce-context/docs/conventions.md` | shared |
| SPEC | Specification + plan (what + how) | `<part>/docs/specs/` | local (api or web) |
| TDR  | Technical Decision Record (local technical decision) | `<part>/docs/technical_decisions/` | local (api or web) |

### A.2 Referencing across parts
IDs are **global across the product**. To point to a doc from another part, use
`ID@part`: `AYD-003@context`, `SPEC-012@api`, `SPEC-013@web`. No suffix = lives in
`context` (REQ, GLO, AYD, ARCH, CONV are unique and unambiguous).

### A.3 Frontmatter (required in every doc)
```yaml
---
id: AYD-001
type: design          # requirements | design | spec | architecture | conventions | tdr
status: draft         # draft | review | approved | done
updated: 2026-07-08
parents: [REQ-01]           # what this doc refines (layer above)
children: [SPEC-001@api]    # what refines this doc (can be cross-part)
related: [GLO]               # cross-cutting context
---
```

### A.4 Status lifecycle
`draft → review → approved` (and `done` for a SPEC already implemented; `proposed →
accepted → superseded` for TDR).
**approved/accepted** = current source of truth.

### A.5 Linking (the graph's "glue")
- Refinement declared on both sides: `children` on the parent, `parents` on the child.
- Every `SPEC` has an `AYD` in `parents` (e.g., `[AYD-NNN@context]`). Every `AYD` has a `REQ`.
- **1 AYD → N SPECs**, one per affected part. The AYD is the source of the contracts.
- Domain terms live only in the **glossary**; other docs just reference them.

### A.6 Lifecycle
| Type | Behavior |
|------|------|
| REQ / GLO / AYD / ARCH / CONV | **Living** — edit in place, update `updated`. |
| SPEC | **Ephemeral** — working document; becomes historical once implemented (`done`). |
| TDR | **Append-only** — never rewritten. A new decision replaces the old one via `superseded_by`. |

Audit trail lives in **git + changelog**.

### A.7 Change propagation
When changing a living doc: (1) edit it and update `updated`; (2) log it in the
corresponding part's `changelog.md` (shared → `ntd-ecomerce-context/docs/changelog.md`;
local → `<part>/docs/changelog.md`); (3) mark affected `children` (including cross-part)
as `status: review` and review them; (4) if the change adds/removes a part or
integration, update `architecture.md` in the **same edit**.

### A.8 Language
All docs are written in **English**, including entities, fields, enums, endpoints, and
events (these carry through to the code). The glossary defines the canonical term for
each concept.

### A.9 Diagrams
**Mermaid embedded in the `.md`** (version-controlled, renders on GitHub) — never a PNG
as the canonical source. Current topology → `architecture.md`; a feature's flow → its
`AYD`. If the diagram diverges from the text, **the text wins**.

---

## B. Code

### B.1 Style
- **Naming:** use the glossary's terms — always in **English** (variables, functions,
  types, routes, entities). Comments may be in English.
- **Shared types** between `ntd-ecomerce-api` and `ntd-ecomerce-web` live in
  `ntd-ecomerce-context/` (single source).
- **Linter/formatter:** configuration and command per part — keep it standardized
  across the monorepo.

### B.2 Tests
- **Structure:** AAA (Arrange, Act, Assert).
- **Coverage:** every acceptance criterion of a SPEC has a corresponding test.
- **Mocks:** only at the boundary (network, db, clock); don't mock the unit under test.

### B.3 Git
- **Branches:** `feature/<spec-id>-description` (e.g., `feature/SPEC-001-cart`).
- **Commits:** [Conventional Commits](https://www.conventionalcommits.org); reference the
  ID (SPEC/AYD/TDR) when applicable.
- **PRs:** link to the SPEC; one line in the corresponding part's `changelog.md` per PR.
