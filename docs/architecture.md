---
id: ARCH
type: architecture
title: Visão de arquitetura (C4 vivo)
status: approved
updated: 2026-07-08
parents: []
related: []
---

# Visão de arquitetura (C4 — contexto + containers)

> **Documento vivo.** Retrata a topologia **vigente** do monorepo: quais partes existem e
> como se conectam. Atualize na **mesma edição** que adiciona/remove uma parte ou
> integração (ver `conventions.md` §A.6 e §A.8). Nomes em **inglês** (atravessam para o
> código).

## Diagrama (container view)

```mermaid
flowchart TB
    user["User (browser)"]

    subgraph monorepo[Monorepo]
        web["web<br/>(frontend)"]
        api["api<br/>(backend · dono dos contratos)"]
        context["context<br/>(tipos/domínio compartilhados)"]
    end

    db[("Database<br/>(Postgres · local)")]

    user -->|HTTP| web
    web -->|HTTP · REST/JSON| api
    api -->|SQL| db
    web -.->|importa tipos| context
    api -.->|importa tipos| context
```

## Containers (legenda)

| Container | Papel | Stack (preencher) |
|-----------|-------|-------------------|
| **web** | Frontend; consome a `api` | _<framework>_ |
| **api** | Backend; regras de negócio e dono dos contratos | _<framework>_ |
| **context** | Tipos e domínio compartilhados entre `web` e `api` | _<TypeScript, etc.>_ |
| **Database** | Persistência do domínio (local) | Postgres (local) |

> Substitua as stacks pelas reais ao definir o projeto. Diagrama e tabela devem ficar em
> sincronia — se divergirem, **a tabela vence**.
