---
id: CONV
type: conventions
title: Convenções de docs e código
status: approved
updated: 2026-07-08
---

# Convenções

O "contrato" que mantém docs e código consistentes e legíveis por humanos e IAs, entre as
três partes do monorepo (`ntd-ecomerce-context`, `ntd-ecomerce-api`, `ntd-ecomerce-web`).
Duas partes: **A) documentação** e **B) código**.

---

## A. Documentação

### A.1 Tipos de documento, IDs e onde moram
ID = `PREFIXO-NNN`, **estável** (nunca muda, mesmo se o arquivo mudar de nome).

| Prefixo | Tipo | Onde | Escopo |
|---------|------|------|--------|
| REQ  | Requisitos | `ntd-ecomerce-context/docs/requirements.md` | compartilhado |
| GLO  | Glossário (linguagem ubíqua) | `ntd-ecomerce-context/docs/requirements.md` (seção) | compartilhado |
| AYD  | Análise & Design de feature | `ntd-ecomerce-context/docs/design/` | compartilhado, cross-parte |
| ARCH | Arquitetura viva (C4) | `ntd-ecomerce-context/docs/architecture.md` | compartilhado |
| CONV | Estas convenções | `ntd-ecomerce-context/docs/conventions.md` | compartilhado |
| SPEC | Especificação + plano (o quê + como) | `<parte>/docs/specs/` | local (api ou web) |
| TDR  | Technical Decision Record (decisão técnica local) | `<parte>/docs/technical_decisions/` | local (api ou web) |

### A.2 Referência entre partes
IDs são **globais no produto**. Para apontar um doc de outra parte, use `ID@parte`:
`AYD-003@context`, `SPEC-012@api`, `SPEC-013@web`. Sem sufixo = mora no `context` (REQ,
GLO, AYD, ARCH, CONV são únicos e não ambíguos).

### A.3 Frontmatter (obrigatório em todo doc)
```yaml
---
id: AYD-001
type: design          # requirements | design | spec | architecture | conventions | tdr
status: draft         # draft | review | approved | done
updated: 2026-07-08
parents: [REQ-01]           # o que este doc refina (camada acima)
children: [SPEC-001@api]    # o que refina este doc (pode ser cross-parte)
related: [GLO]               # contexto transversal
---
```

### A.4 Ciclo de status
`draft → review → approved` (e `done` para uma SPEC já implementada; `proposed →
accepted → superseded` para TDR).
**approved/accepted** = fonte da verdade vigente.

### A.5 Linkagem (a "cola" do grafo)
- Refinamento declarado nos dois lados: `children` no pai, `parents` no filho.
- Toda `SPEC` tem um `AYD` em `parents` (ex.: `[AYD-NNN@context]`). Todo `AYD` tem um `REQ`.
- **1 AYD → N SPECs**, uma por parte afetada. O AYD é a fonte dos contratos.
- Termos de domínio vivem só no **glossário**; os outros docs apenas referenciam.

### A.6 Ciclo de vida
| Tipo | Comportamento |
|------|---------------|
| REQ / GLO / AYD / ARCH / CONV | **Vivo** — edita in-place, atualiza `updated`. |
| SPEC | **Efêmero** — documento de trabalho; vira histórico após implementado (`done`). |
| TDR | **Append-only** — nunca reescreve. Decisão nova substitui a antiga via `superseded_by`. |

Auditoria mora no **git + changelog**.

### A.7 Propagação de mudanças
Ao alterar um doc vivo: (1) edite e atualize `updated`; (2) registre no `changelog.md` da
parte correspondente (compartilhado → `ntd-ecomerce-context/docs/changelog.md`; local →
`<parte>/docs/changelog.md`); (3) marque os `children` afetados (inclusive cross-parte)
como `status: review` e revise-os; (4) se a mudança adiciona/remove uma parte ou
integração, atualize `architecture.md` na **mesma edição**.

### A.8 Idioma
Prosa dos docs em **português**. Entidades, campos, enums, endpoints e eventos em
**inglês** (atravessam para o código). O glossário define o termo canônico em inglês.
**Exceção:** os `changelog.md` são em inglês (padrão Keep a Changelog).

### A.9 Diagramas
**Mermaid embutido no `.md`** (versionável, renderiza no GitHub) — nunca PNG como
canônico. Topologia vigente → `architecture.md`; fluxo de uma feature → o `AYD` dela.
Se o diagrama divergir do texto, **o texto vence**.

---

## B. Código

### B.1 Estilo
- **Nomenclatura:** use os termos do glossário — sempre em **inglês** (variáveis, funções,
  tipos, rotas, entidades). Comentários podem ser em português.
- **Tipos compartilhados** entre `ntd-ecomerce-api` e `ntd-ecomerce-web` vivem em
  `ntd-ecomerce-context/` (fonte única).
- **Linter/formatter:** configuração e comando por parte — mantenha padronizado no
  monorepo.

### B.2 Testes
- **Estrutura:** AAA (Arrange, Act, Assert).
- **Cobertura:** todo critério de aceite de uma SPEC tem um teste correspondente.
- **Mocks:** só na fronteira (rede, db, relógio); não mocke a unidade testada.

### B.3 Git
- **Branches:** `feature/<id-da-spec>-descricao` (ex.: `feature/SPEC-001-cart`).
- **Commits:** [Conventional Commits](https://www.conventionalcommits.org); referencie o
  ID (SPEC/AYD/TDR) quando aplicável.
- **PRs:** vinculam a SPEC; uma linha no `changelog.md` da parte correspondente por PR.
