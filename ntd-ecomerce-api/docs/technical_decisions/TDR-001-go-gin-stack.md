---
id: TDR-001
type: tdr
title: Go + Gin as the API language and HTTP framework
status: accepted
updated: 2026-07-08
parents: [AYD-001@context]
related: [TDR-002, TDR-004]
superseded_by: null
---

# TDR-001: Go + Gin as the API language and HTTP framework

## Context
The api part needs a language and HTTP framework to implement the REST contracts
defined in the AYDs (starting with AYD-001@context). The operator's primary backend
experience is Go, and the team's coding conventions (`.claude/skills/go-*`) already
encode a Gin-based clean-architecture style — reusing them lowers delivery risk for
the whole MVP (CRUD, CSV import, search, purchase).

## Decision
- **Go 1.25** as the API language, module name `ntd-ecomerce-api`.
- **Gin** (`github.com/gin-gonic/gin`) as the HTTP framework: routing, JSON binding,
  middleware (recovery, logging, CORS).
- Handler conventions follow the `go-api-handlers` skill: thin handlers, narrow
  usecase interfaces declared at point of use, centralized error mapping to the
  AYD-001 error envelope.

## Alternatives & trade-offs
- **chi**: more idiomatic (pure `net/http`), but the existing conventions/skills are
  Gin-shaped; rewriting them buys no MVP value.
- **net/http (1.22+ routing)**: zero dependencies, but more boilerplate for binding,
  middleware, and route groups.
- **echo**: comparable to Gin; no advantage that justifies diverging from the
  operator's familiarity.
