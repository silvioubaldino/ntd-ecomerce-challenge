---
id: TDR-003
type: tdr
title: Schema migrations with golang-migrate, embedded and run at boot
status: accepted
updated: 2026-07-08
parents: [AYD-001@context]
related: [TDR-002]
superseded_by: null
---

# TDR-003: Schema migrations with golang-migrate, embedded and run at boot

## Context
The Postgres schema (starting with the `products` table, AYD-001@context) needs
versioned, repeatable evolution that works both in local dev and in the
Docker-composed stack (RNF-01) without manual steps.

## Decision
- **golang-migrate** (`github.com/golang-migrate/migrate/v4`) with versioned SQL
  up/down files in `ntd-ecomerce-api/migrations/`
  (`NNNNNN_description.up.sql` / `.down.sql`).
- Migrations are **embedded** (`embed.FS`) and applied automatically at API boot,
  before the HTTP server starts. `docker compose up` therefore yields a ready
  schema with no extra container or manual command.
- Down migrations are written for every up migration but only run manually
  (local dev rollback).

## Alternatives & trade-offs
- **goose**: equivalent capability; golang-migrate is the wider de-facto standard.
- **Postgres `/docker-entrypoint-initdb.d` init scripts**: no versioning, only runs
  on an empty volume — breaks on schema evolution.
- **Separate migrate container/CLI step in compose**: more explicit, but adds a
  moving part; boot-time embedding is simpler for a single-API stack. Revisit if
  multiple services ever share the schema.
