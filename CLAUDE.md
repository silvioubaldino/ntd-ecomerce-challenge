# NTD E-commerce Challenge — guia da IA

Monorepo full stack de e-commerce com três partes: **`ntd-ecomerce-context/`**
(tipos/domínio compartilhados + doc compartilhada, fonte única), **`ntd-ecomerce-api/`**
(backend, dono dos contratos) e **`ntd-ecomerce-web/`** (frontend, consome a api), com um
**Postgres local**. A documentação é **spec-driven** e enxuta.

## Comece por aqui
- @ntd-ecomerce-context/docs/requirements.md — requisitos **e glossário** (SEMPRE use estes termos)
- @ntd-ecomerce-context/docs/conventions.md — convenções de docs **e** de código (IDs, `ID@parte`, frontmatter, estilo, git, testes)
- @ntd-ecomerce-context/docs/architecture.md — arquitetura viva (C4) do monorepo

## Onde cada coisa mora
- **REQ + Glossário** → `ntd-ecomerce-context/docs/requirements.md`
- **AYD** (design de feature: contratos api↔web, modelo, fluxo) → `ntd-ecomerce-context/docs/design/`
- **Arquitetura viva (C4)** → `ntd-ecomerce-context/docs/architecture.md`
- **Changelog compartilhado** → `ntd-ecomerce-context/docs/changelog.md`
- **SPEC** (o quê + como de uma feature, por parte) → `ntd-ecomerce-api/docs/specs/` e `ntd-ecomerce-web/docs/specs/`
- **TDR** (decisão técnica local de uma parte) → `ntd-ecomerce-api/docs/technical_decisions/` e `ntd-ecomerce-web/docs/technical_decisions/`
- **Changelog local** de cada serviço → `<parte>/docs/changelog.md`

## Fluxo de uma feature
1. Leia o **REQ** relevante em `ntd-ecomerce-context/docs/requirements.md` (e confirme os
   termos no glossário).
2. Escreva/atualize o **AYD** (`ntd-ecomerce-context/docs/design/AYD-NNN.md`): quais partes
   toca (`api`/`web`), os **contratos** (endpoints/payloads), o modelo de domínio e o
   fluxo. O AYD é a fonte dos contratos.
3. Para cada parte afetada, escreva a **SPEC** (`<parte>/docs/specs/SPEC-NNN.md`,
   `parents: [AYD-NNN@context]`): **direto e objetivo** — o que fazer e como (critérios
   de aceite + passos + testes).
4. Implemente na parte correspondente. Decisão técnica local não trivial (não muda
   contrato) → **TDR** (`<parte>/docs/technical_decisions/`); se mudar contrato/protocolo
   entre api e web, volte para o **AYD** no `ntd-ecomerce-context`.
5. Registre **1 linha** no `changelog.md` da parte (compartilhado → context; local →
   api/web) e, se a topologia mudou (nova parte/integração), atualize
   `ntd-ecomerce-context/docs/architecture.md` no mesmo PR.

## Regras de ouro
- O **glossário** (em `requirements.md`) define o termo canônico em **inglês** — código e
  docs usam esse termo. Adicione o termo lá **antes** de usá-lo.
- Contrato de uma feature muda no **AYD** (context); a SPEC implementa, não redefine.
- Referência entre partes usa `ID@parte` (ex.: `SPEC-012@api`). Sem sufixo = context.
- Prosa dos docs em **português**; entidades/campos/enums/endpoints em **inglês**.
- Mudou um doc vivo? Atualize `updated` e marque `children` afetados como `status: review`.
