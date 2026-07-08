---
id: CONV
type: conventions
title: Convenções de docs e código
status: approved
updated: 2026-07-08
---

# Convenções

O "contrato" que mantém docs e código consistentes e legíveis por humanos e IAs.
Duas partes: **A) documentação** e **B) código**.

---

## A. Documentação

### A.1 Tipos de documento e IDs
ID = `PREFIXO-NNN`, **estável** (nunca muda, mesmo se o arquivo mudar de nome).

| Prefixo | Tipo | Onde |
|---------|------|------|
| REQ  | Requisitos | `docs/requirements.md` |
| GLO  | Glossário (linguagem ubíqua) | `docs/requirements.md` (seção) |
| AYD  | Análise & Design de feature | `docs/design/` |
| SPEC | Especificação + plano (o quê + como) | `docs/specs/` |
| ARCH | Arquitetura viva (C4) | `docs/architecture.md` |
| CONV | Estas convenções | `docs/conventions.md` |

### A.2 Frontmatter (obrigatório em todo doc)
```yaml
---
id: AYD-001
type: design          # requirements | design | spec | architecture | conventions
status: draft         # draft | review | approved | done
updated: 2026-07-08
parents: [REQ-01]     # o que este doc refina (camada acima)
children: [SPEC-001]  # o que refina este doc
related: [GLO]        # contexto transversal
---
```

### A.3 Ciclo de status
`draft → review → approved` (e `done` para uma SPEC já implementada).
**approved** = fonte da verdade vigente.

### A.4 Linkagem (a "cola" do grafo)
- Refinamento declarado nos dois lados: `children` no pai, `parents` no filho.
- Toda `SPEC` tem um `AYD` em `parents`. Todo `AYD` tem um `REQ` em `parents`.
- Termos de domínio vivem só no **glossário**; os outros docs apenas referenciam.

### A.5 Ciclo de vida
| Tipo | Comportamento |
|------|---------------|
| REQ / GLO / AYD / ARCH / CONV | **Vivo** — edita in-place, atualiza `updated`. |
| SPEC | **Efêmero** — documento de trabalho; vira histórico após implementado (`done`). |

Auditoria mora no **git + changelog**.

### A.6 Propagação de mudanças
Ao alterar um doc vivo: (1) edite e atualize `updated`; (2) registre no `changelog.md`;
(3) marque os `children` afetados como `status: review` e revise-os; (4) se a mudança
adiciona/remove um serviço ou integração, atualize `architecture.md` na **mesma edição**.

### A.7 Idioma
Prosa dos docs em **português**. Entidades, campos, enums, endpoints e eventos em
**inglês** (atravessam para o código). O glossário define o termo canônico em inglês.
**Exceção:** o `changelog.md` é em inglês (padrão Keep a Changelog).

### A.8 Diagramas
**Mermaid embutido no `.md`** (versionável, renderiza no GitHub) — nunca PNG como
canônico. Topologia vigente → `architecture.md`; fluxo de uma feature → o `AYD` dela.
Se o diagrama divergir do texto, **o texto vence**.

---

## B. Código

### B.1 Estilo
- **Nomenclatura:** use os termos do glossário — sempre em **inglês** (variáveis, funções,
  tipos, rotas, entidades). Comentários podem ser em português.
- **Tipos compartilhados** entre `api` e `web` vivem em `context/` (fonte única).
- **Linter/formatter:** configuração e comando por parte (`api`/`web`) — mantenha
  padronizado no monorepo.

### B.2 Testes
- **Estrutura:** AAA (Arrange, Act, Assert).
- **Cobertura:** todo critério de aceite de uma SPEC tem um teste correspondente.
- **Mocks:** só na fronteira (rede, db, relógio); não mocke a unidade testada.

### B.3 Git
- **Branches:** `feature/<id-da-spec>-descricao` (ex.: `feature/SPEC-001-cart`).
- **Commits:** [Conventional Commits](https://www.conventionalcommits.org); referencie o
  ID (SPEC/AYD) quando aplicável.
- **PRs:** vinculam a SPEC; uma linha no `changelog.md` por PR.
