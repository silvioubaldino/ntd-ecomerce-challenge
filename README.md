# 🛒 NTD E-commerce Challenge

Coding challenge full stack (e-commerce). **Monorepo** com três partes e um banco local:

```
ntd-ecomerce-context/   → tipos/domínio compartilhados + doc compartilhada (fonte única)
ntd-ecomerce-api/       → backend (dono dos contratos e regras de negócio)
ntd-ecomerce-web/       → frontend (consome a api)
db (local)              → Postgres rodando localmente
```

> Este README documenta o **framework de docs** (versão enxuta). O código de cada parte
> entra depois.

## Estrutura

```
ntd-ecomerce-context/
  docs/
    requirements.md        # REQ + GLOSSÁRIO (agrupados)
    conventions.md         # docs + código (agrupados)
    architecture.md        # C4 vivo do monorepo
    changelog.md           # mudanças na doc compartilhada
    design/
      AYD-NNN.md            # Análise & Design (contratos entre api/web)
ntd-ecomerce-api/
  docs/
    specs/
      SPEC-NNN.md            # o quê + como (spec + plano num arquivo)
    technical_decisions/
      TDR-NNN.md             # decisões técnicas locais deste serviço
    changelog.md            # mudanças na implementação deste serviço
ntd-ecomerce-web/
  docs/
    specs/
      SPEC-NNN.md
    technical_decisions/
      TDR-NNN.md
    changelog.md
```

## Framework de docs (versão enxuta)

Documentação **spec-driven** — do *o quê* ao *como* — legível por humanos e consumível
por IAs (Claude Code). Cada artefato tem **ID estável**, **status** e declara seus
**relacionamentos** (`parents`/`children`), formando um grafo que a IA percorre.

`ntd-ecomerce-context` guarda a camada **compartilhada** (fonte única da verdade). Cada
serviço (`ntd-ecomerce-api`, `ntd-ecomerce-web`) guarda só o seu: specs, decisões
técnicas locais (TDR) e changelog. Referências entre partes usam `ID@parte`
(ex.: `AYD-003@context`, `SPEC-012@api`) — ver `conventions.md` §A.2.

### Fluxo de trabalho
```
REQ  ──►  AYD  ──►  SPEC (por parte)  ──►  código
(o que    (design    (o que + o como)      (implementa)
 é preciso) da feature)
```
1. **REQ** (`ntd-ecomerce-context`) — o requisito a atender (e os termos no glossário).
2. **AYD** (`ntd-ecomerce-context/docs/design/`) — desenha a feature: quais partes toca
   (`api`/`web`), contratos, modelo e fluxo.
3. **SPEC** (`<parte>/docs/specs/`) — descreve, direto e objetivo, **o que fazer e como**
   (passos + testes), uma por parte afetada.
4. Implementa; decisão técnica local não trivial → **TDR** na parte; registra 1 linha no
   **changelog** correspondente; mantém **architecture** em dia se a topologia mudar.

Detalhes de IDs, frontmatter e ciclo de vida: veja `ntd-ecomerce-context/docs/conventions.md`.

## Como usar com Claude Code
`CLAUDE.md` (raiz) é o ponto de entrada da IA: importa requisitos, convenções e
arquitetura, e descreve o fluxo acima.
