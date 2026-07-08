# NTD E-commerce Challenge — AI guide

Full-stack e-commerce monorepo with three parts: **`ntd-ecomerce-context/`**
(shared types/domain + shared docs, single source of truth), **`ntd-ecomerce-api/`**
(backend, owns the contracts), and **`ntd-ecomerce-web/`** (frontend, consumes the api),
with a **local Postgres**. Documentation is **spec-driven** and lean.

## Start here
- @ntd-ecomerce-context/docs/requirements.md — requirements **and glossary** (ALWAYS use these terms)
- @ntd-ecomerce-context/docs/conventions.md — docs **and** code conventions (IDs, `ID@part`, frontmatter, style, git, tests)
- @ntd-ecomerce-context/docs/architecture.md — living architecture (C4) of the monorepo

## Where things live
- **REQ + Glossary** → `ntd-ecomerce-context/docs/requirements.md`
- **AYD** (feature design: api↔web contracts, model, flow) → `ntd-ecomerce-context/docs/design/`
- **Living architecture (C4)** → `ntd-ecomerce-context/docs/architecture.md`
- **Shared changelog** → `ntd-ecomerce-context/docs/changelog.md`
- **SPEC** (what + how of a feature, per part) → `ntd-ecomerce-api/docs/specs/` and `ntd-ecomerce-web/docs/specs/`
- **TDR** (local technical decision for a part) → `ntd-ecomerce-api/docs/technical_decisions/` and `ntd-ecomerce-web/docs/technical_decisions/`
- **Local changelog** for each service → `<part>/docs/changelog.md`

## Feature workflow
1. Read the relevant **REQ** in `ntd-ecomerce-context/docs/requirements.md` (and confirm
   the terms in the glossary).
2. Write/update the **AYD** (`ntd-ecomerce-context/docs/design/AYD-NNN.md`): which parts
   it touches (`api`/`web`), the **contracts** (endpoints/payloads), the domain model,
   and the flow. The AYD is the source of the contracts.
3. For each affected part, write the **SPEC** (`<part>/docs/specs/SPEC-NNN.md`,
   `parents: [AYD-NNN@context]`): **direct and objective** — what to do and how
   (acceptance criteria + steps + tests).
4. Implement in the corresponding part. Non-trivial local technical decision (doesn't
   change the contract) → **TDR** (`<part>/docs/technical_decisions/`); if it changes
   the contract/protocol between api and web, go back to the **AYD** in
   `ntd-ecomerce-context`.
5. Log **1 line** in the part's `changelog.md` (shared → context; local → api/web) and,
   if the topology changed (new part/integration), update
   `ntd-ecomerce-context/docs/architecture.md` in the same PR.

## Golden rules
- The **glossary** (in `requirements.md`) defines the canonical term in **English** —
  code and docs use that term. Add the term there **before** using it.
- A feature's contract changes in the **AYD** (context); the SPEC implements, it doesn't
  redefine.
- References between parts use `ID@part` (e.g., `SPEC-012@api`). No suffix = context.
- All docs are written in **English**.
- Changed a living doc? Update `updated` and mark affected `children` as `status: review`.
