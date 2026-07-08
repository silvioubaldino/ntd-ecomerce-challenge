---
id: TDR-002
type: tdr
title: GORM over Postgres, decimals via shopspring/decimal
status: accepted
updated: 2026-07-08
parents: [AYD-001@context]
related: [TDR-001, TDR-003]
superseded_by: null
---

# TDR-002: GORM over Postgres, decimals via shopspring/decimal

## Context
AYD-001@context mandates Postgres persistence with `NUMERIC` for money/weight and
decimals traveling as **strings** in JSON — never floats. The repository layer needs
a data-access approach that keeps CRUD fast to write for the MVP while honoring
those constraints.

## Decision
- **GORM v2** (`gorm.io/gorm`) with the Postgres driver (`gorm.io/driver/postgres`,
  pgx-based) as the data-access layer.
- **Schema is owned by SQL migrations (TDR-003), not by GORM**: `AutoMigrate` is
  forbidden; GORM models must mirror the migrated schema.
- **`github.com/shopspring/decimal`** for `price` and `weight_kg`:
  `decimal.Decimal` implements `sql.Scanner`/`driver.Valuer` (maps to `NUMERIC`)
  and marshals to/from JSON strings, satisfying the AYD-001 convention end to end.
- Persistence models live in the repository package with explicit
  `FromDomain`/`ToDomain` mapping — domain types never carry GORM tags.

## Alternatives & trade-offs
- **pgx + sqlc**: type-safe generated code and full SQL control; more setup and
  codegen ceremony than the MVP needs, and less familiar to the operator.
- **pgx + sqlx / manual scan**: simple, but repetitive mapping boilerplate across
  four features.
- **ent**: powerful schema-as-code, but a heavy learning curve for a challenge MVP.
- **GORM `AutoMigrate`** (instead of migrations): convenient but hides schema
  evolution and can't express constraints precisely — rejected; see TDR-003.
