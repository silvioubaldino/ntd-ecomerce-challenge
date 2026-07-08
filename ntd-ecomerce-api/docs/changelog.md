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
[SemVer](https://semver.org/spec/v2.0.0.html). Most recent on top; 1 line per PR.

## Unreleased

- Implemented SPEC-001 (Product catalog CRUD): full Go service (domain, usecase, GORM repository, Gin handlers, bootstrap DI), embedded migration, Docker Compose, and unit tests across every layer.
- Added TDR-001..004 (Go+Gin, GORM+decimal, golang-migrate, layered architecture), SPEC-001 (Product CRUD), and adapted the Go skills to this service's conventions.
- Service scaffold initialized (specs, technical_decisions).
