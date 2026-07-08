---
id: ARCH
type: architecture
title: Architecture overview (living C4)
status: approved
updated: 2026-07-08
parents: []
related: []
---

# Architecture overview (C4 — context + containers)

> **Living document.** Depicts the **current** topology of the monorepo: which parts
> exist and how they connect. Update it in the **same edit** that adds/removes a part
> or integration (see `conventions.md` §A.6 and §A.8). Names in **English** (they carry
> through to the code).

## Diagram (container view)

```mermaid
flowchart TB
    user["User (browser)"]

    subgraph monorepo[Monorepo]
        web["web<br/>(frontend)"]
        api["api<br/>(backend · owns the contracts)"]
        context["context<br/>(shared types/domain)"]
    end

    db[("Database<br/>(Postgres · local)")]

    user -->|HTTP| web
    web -->|HTTP · REST/JSON| api
    api -->|SQL| db
    web -.->|imports types| context
    api -.->|imports types| context
```

## Containers (legend)

| Container | Role | Stack (fill in) |
|-----------|------|-------------------|
| **web** | Frontend; consumes the `api` | _<framework>_ |
| **api** | Backend; business rules and owner of the contracts | _<framework>_ |
| **context** | Types and domain shared between `web` and `api` | _<TypeScript, etc.>_ |
| **Database** | Domain persistence (local) | Postgres (local) |

> Replace the stacks with the real ones once the project is defined. Diagram and table
> must stay in sync — if they diverge, **the table wins**.
