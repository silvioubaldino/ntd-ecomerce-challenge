---
id: CHANGELOG-context
type: changelog
title: Changelog — context
status: approved
updated: 2026-07-10
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

- Approved AYD-006 - Catalog filters and sorting. [#21](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/21)
- Specced the Checkout & Order design (AYD-005) for both parts: linked SPEC-005@api and SPEC-005@web as the checkout contract turning a guest Cart into a confirmed, immutable Order with a simulated payment. [#18](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/18)
- Approved the Cart contract (AYD-004): the guest Cart endpoints, quantity/stock rules (RN-03) and priced response are finalized as the source of truth for SPEC-004@api and SPEC-004@web. [#17](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/17)
- Added multi-item purchase to the scope: a guest Cart (AYD-004) and Checkout & Order (AYD-005), with the related requirements and glossary terms. [#15](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/15)
- Created AYD-003 (Product catalog search): extends `GET /products` with an optional free-text `q` filter over Product text fields, reusing the AYD-001 list/pagination contract — no new endpoint. [#11](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/11)
- Updated AYD-002 (Product CSV bulk import): added a downloadable blank CSV template to the web import UX (static asset, no new api endpoint). [#09](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/9)
- Created AYD-002 (Product CSV bulk import): upload endpoint contract and per-row validation/reporting rules (RN-01/RN-02), reusing the AYD-001 Product model. [#08](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/8)
- Chose and documented the web stack; linked AYD-001 to SPEC-001@web. [#05](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/5)
- Chose and documented the api/db stack; linked AYD-001 to SPEC-001@api. [#03](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/3)
- Created AYD-001 (Product catalog CRUD): first api↔web contract, Product model, and base REST conventions. [#02](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/2)
- Filled REQ-01 (functional/non-functional requirements, business rules, MVP scope) and added the SKU term to the glossary. [#01](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/1) 
- Docs framework restructured into per-part folders (`ntd-ecomerce-context`, `ntd-ecomerce-api`, `ntd-ecomerce-web`).

