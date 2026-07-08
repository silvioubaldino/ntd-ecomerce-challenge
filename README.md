# 🛒 NTD E-commerce Challenge

Full-stack coding challenge (e-commerce). **Monorepo** with three parts and a local database:

```
ntd-ecomerce-context/   → shared types/domain + shared docs (single source of truth)
ntd-ecomerce-api/       → backend (owns contracts and business rules)
ntd-ecomerce-web/       → frontend (consumes the api)
db (local)              → Postgres running locally
```

> This README documents the **docs framework** (lean version). Each part's code
> comes next.

## Structure

```
ntd-ecomerce-context/
  docs/
    requirements.md        # REQ + GLOSSARY (grouped)
    conventions.md         # docs + code (grouped)
    architecture.md        # living C4 of the monorepo
    changelog.md           # changes to the shared docs
    design/
      AYD-NNN.md            # Analysis & Design (contracts between api/web)
ntd-ecomerce-api/
  docs/
    specs/
      SPEC-NNN.md            # what + how (spec + plan in one file)
    technical_decisions/
      TDR-NNN.md             # local technical decisions for this service
    changelog.md            # changes to this service's implementation
ntd-ecomerce-web/
  docs/
    specs/
      SPEC-NNN.md
    technical_decisions/
      TDR-NNN.md
    changelog.md
```

## Docs framework (lean version)

**Spec-driven** documentation — from *what* to *how* — readable by humans and
consumable by AIs (Claude Code). Each artifact has a **stable ID**, a **status**, and
declares its **relationships** (`parents`/`children`), forming a graph the AI can walk.

`ntd-ecomerce-context` holds the **shared** layer (single source of truth). Each
service (`ntd-ecomerce-api`, `ntd-ecomerce-web`) holds only its own: specs, local
technical decisions (TDR), and changelog. References between parts use `ID@part`
(e.g., `AYD-003@context`, `SPEC-012@api`) — see `conventions.md` §A.2.

### Workflow
```
REQ  ──►  AYD  ──►  SPEC (per part)  ──►  code
(what's   (feature   (what + how)         (implements)
 needed)   design)
```
1. **REQ** (`ntd-ecomerce-context`) — the requirement to meet (and the terms in the glossary).
2. **AYD** (`ntd-ecomerce-context/docs/design/`) — designs the feature: which parts it
   touches (`api`/`web`), contracts, model, and flow.
3. **SPEC** (`<part>/docs/specs/`) — describes, directly and objectively, **what to do
   and how** (steps + tests), one per affected part.
4. Implement; non-trivial local technical decision → **TDR** in the part; log 1 line in
   the corresponding **changelog**; keep **architecture** up to date if the topology
   changed.

Details on IDs, frontmatter, and lifecycle: see `ntd-ecomerce-context/docs/conventions.md`.

## Using it with Claude Code
`CLAUDE.md` (root) is the AI's entry point: it imports requirements, conventions, and
architecture, and describes the workflow above.
