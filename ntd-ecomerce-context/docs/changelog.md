---
id: CHANGELOG-context
type: changelog
title: Changelog — context
status: approved
updated: 2026-07-08
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
  no implementation detail; reference the PR when useful.

## Unreleased

- Docs framework restructured into per-part folders (`ntd-ecomerce-context`, `ntd-ecomerce-api`, `ntd-ecomerce-web`).
