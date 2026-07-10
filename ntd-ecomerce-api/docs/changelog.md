---
id: CHANGELOG-api
type: changelog
title: Changelog — api
status: approved
updated: 2026-07-08
---

# Changelog — api

Changes to this service's implementation (specs, code, local technical decisions).
Changes to the **shared** docs (requirements, design, architecture) go in
`ntd-ecomerce-context`'s changelog.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) +
[SemVer](https://semver.org/spec/v2.0.0.html). Most recent on top; 1 line per PR,
stating **what** shipped — no stack/library names or implementation detail (see
CONV §B.3; that detail lives in the SPEC/TDR).

## Unreleased

- Implemented SPEC-006 - Catalog filters and sorting: the product list now supports narrowing by category and price range and sorting the results, with a new endpoint listing the available categories, and free-text search is scoped away from category.
- Implemented SPEC-005 - Checkout & Order: create a confirmed Order from a guest Cart with a simulated payment, re-checking stock and snapshotting prices while decrementing stock atomically, and retrieve an Order by id. [#18](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/18)
- Implemented SPEC-004 - Cart resource: create a guest cart, add/increment, set and remove items with stock-checked quantities, returning line subtotals and a cart total. [#17](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/17)
- Implemented SPEC-003 - Product catalog search, extending the product list with an optional free-text filter. [#14](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/14)
- Implemented SPEC-002 - Product bulk import from CSV, with per-row validation and rejection reporting. [#09](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/9)
- Implemented SPEC-001 - Product catalog CRUD [#04](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/4)
- Chose and documented the api stack; drafted SPEC-001 (Product CRUD). [#03](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/3)
- Service scaffold initialized (specs, technical_decisions).
