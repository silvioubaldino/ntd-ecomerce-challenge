---
id: CHANGELOG-context
type: changelog
title: Changelog — context
status: approved
updated: 2026-07-09
---

# Changelog — context

Changes to the **shared** docs (requirements, glossary, architecture, design/AYD,
conventions). Implementation changes for each service go in its own local changelog
(`ntd-ecomerce-api/docs/changelog.md`, `ntd-ecomerce-web/docs/changelog.md`).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

**Policy**:

- **Order:** most recent on top; new entries go **above** the previous ones.
- **Unreleased:** unreleased work accrues under `## Unreleased` (always the top block), with
no date/version. On a release, `## Unreleased` becomes `## [dd-MM-yyyy - vX.Y.Z]` and a
new empty `## Unreleased` is opened above it.
- **One line per PR:** each PR adds a **single line** summarizing what it delivers — general,
no implementation detail (no stack/library names, file paths, or how it was built —
that lives in the SPEC/TDR/PR); reference the PR when useful. See CONV §B.3.



## Unreleased

- Updated AYD-002 (Product CSV bulk import): added a downloadable blank CSV template to the web import UX (static asset, no new api endpoint). [#09](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/9)
- Created AYD-002 (Product CSV bulk import): upload endpoint contract and per-row validation/reporting rules (RN-01/RN-02), reusing the AYD-001 Product model. [#08](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/8)
- Chose and documented the web stack; linked AYD-001 to SPEC-001@web. [#05](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/5)
- Chose and documented the api/db stack; linked AYD-001 to SPEC-001@api. [#03](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/3)
- Created AYD-001 (Product catalog CRUD): first api↔web contract, Product model, and base REST conventions. [#02](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/2)
- Filled REQ-01 (functional/non-functional requirements, business rules, MVP scope) and added the SKU term to the glossary. [#01](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/1) 
- Docs framework restructured into per-part folders (`ntd-ecomerce-context`, `ntd-ecomerce-api`, `ntd-ecomerce-web`).

