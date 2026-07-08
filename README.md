# 🛒 NTD E-commerce Challenge

Coding challenge full stack (e-commerce). **Monorepo** com três partes e um banco local:

```
context/   → tipos/domínio compartilhados entre api e web
api/       → backend (dono dos contratos e regras de negócio)
web/       → frontend (consome a api)
db (local) → Postgres rodando localmente
```

> Este README documenta o **framework de docs** (versão enxuta). O código das partes
> `context/`, `api/` e `web/` entra depois.

## Framework de docs (versão enxuta)

Documentação **spec-driven** — do *o quê* ao *como* — legível por humanos e consumível
por IAs (Claude Code). Cada artefato tem **ID estável**, **status** e declara seus
**relacionamentos** (`parents`/`children`), formando um grafo que a IA percorre.

Como é um **monorepo**, tudo vive em uma única pasta `docs/` — sem espelhos, sem sync
entre repos.

### Mapa dos documentos
| Arquivo | ID | O que é | Ciclo |
|---------|----|---------|-------|
| `docs/requirements.md` | REQ + **GLO** | Requisitos do produto **+ glossário** (linguagem ubíqua) | vivo |
| `docs/architecture.md` | ARCH | Visão de arquitetura C4 (viva) do monorepo | vivo |
| `docs/conventions.md` | CONV | Convenções de **docs + código** (IDs, frontmatter, estilo, git, testes) | vivo |
| `docs/design/AYD-NNN.md` | AYD | Análise & Design de uma feature (contratos entre api/web) | vivo |
| `docs/specs/SPEC-NNN.md` | SPEC | **O quê + como** de uma feature (spec + plano num arquivo só) | efêmero |
| `docs/changelog.md` | — | Histórico do produto (Keep a Changelog) | append |

### Fluxo de trabalho
```
REQ  ──►  AYD  ──►  SPEC  ──►  código
(o que    (design    (o que +   (implementa)
 é preciso) da feature) o como)
```
1. **REQ** — o requisito a atender (e os termos no glossário).
2. **AYD** — desenha a feature: quais partes toca (`api`/`web`), contratos, modelo e fluxo.
3. **SPEC** — descreve, direto e objetivo, **o que fazer e como** (passos + testes).
4. Implementa; registra 1 linha no **changelog**; mantém **architecture** em dia se a
   topologia mudar.

Detalhes de IDs, frontmatter e ciclo de vida: veja `docs/conventions.md`.

## Como usar com Claude Code
`CLAUDE.md` (raiz) é o ponto de entrada da IA: importa requisitos, convenções e
arquitetura, e descreve o fluxo acima.
