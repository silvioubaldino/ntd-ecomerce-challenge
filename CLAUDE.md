# NTD E-commerce Challenge — guia da IA

Monorepo full stack de e-commerce: **`context/`** (tipos/domínio compartilhados),
**`api/`** (backend, dono dos contratos) e **`web/`** (frontend, consome a api), com um
**Postgres local**. A documentação é **spec-driven** e enxuta.

## Comece por aqui
- @docs/requirements.md — requisitos **e glossário** (SEMPRE use estes termos)
- @docs/conventions.md — convenções de docs **e** de código (IDs, frontmatter, estilo, git, testes)
- @docs/architecture.md — arquitetura viva (C4) do monorepo

## Onde cada coisa mora
- **REQ + Glossário** → `docs/requirements.md`
- **AYD** (design de feature: contratos api↔web, modelo, fluxo) → `docs/design/`
- **SPEC** (o quê + como de uma feature) → `docs/specs/`
- **Arquitetura viva (C4)** → `docs/architecture.md`
- **Changelog** → `docs/changelog.md`

## Fluxo de uma feature
1. Leia o **REQ** relevante em `docs/requirements.md` (e confirme os termos no glossário).
2. Escreva/atualize o **AYD** (`docs/design/AYD-NNN.md`): quais partes toca (`api`/`web`),
   os **contratos** (endpoints/payloads), o modelo de domínio e o fluxo. O AYD é a fonte
   dos contratos.
3. Escreva a **SPEC** (`docs/specs/SPEC-NNN.md`, `parents: [AYD-NNN]`): **direto e
   objetivo** — o que fazer e como (critérios de aceite + passos + testes).
4. Implemente em `api/`/`web/`, com os tipos compartilhados em `context/`.
5. Registre **1 linha** no `docs/changelog.md` e, se a topologia mudou (novo serviço/
   integração), atualize `docs/architecture.md` no mesmo PR.

## Regras de ouro
- O **glossário** (em `requirements.md`) define o termo canônico em **inglês** — código e
  docs usam esse termo. Adicione o termo lá **antes** de usá-lo.
- Contrato de uma feature muda no **AYD**; a SPEC implementa, não redefine.
- Prosa dos docs em **português**; entidades/campos/enums/endpoints em **inglês**.
- Mudou um doc vivo? Atualize `updated` e marque `children` afetados como `status: review`.
