---
id: CHANGELOG-web
type: changelog
title: Changelog — web
status: approved
updated: 2026-07-08
---

# Changelog — web

Changes to this service's implementation (specs, code, local technical decisions).
Changes to the **shared** docs (requirements, design, architecture) go in
`ntd-ecomerce-context`'s changelog.

Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) +
[SemVer](https://semver.org/spec/v2.0.0.html). Most recent on top; 1 line per PR,
stating **what** shipped — no stack/library names or implementation detail (see
CONV §B.3; that detail lives in the SPEC/TDR).

## Unreleased

- Implemented SPEC-003 (Store product search UI: search box with URL-synced term, debounced queries, and paginated results). [#13](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/13)
- Implemented SPEC-002 (Product CSV bulk import UI: upload, import summary, rejected-rows report, blank template download). [#10](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/10)
- Redesigned the catalog UI. [#07](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/7)
- Implemented SPEC-001 (Product catalog CRUD UI: list, create, edit, delete); service now runs via Docker. [#06](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/6)
- Added SPEC-001@web (Product catalog CRUD UI) and its supporting technical decisions. [#05](https://github.com/silvioubaldino/ntd-ecomerce-challenge/pull/5)
- Service scaffold initialized (specs, technical_decisions).
