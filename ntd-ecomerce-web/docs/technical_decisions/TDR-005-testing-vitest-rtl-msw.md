---
id: TDR-005
type: tdr
title: Vitest + React Testing Library + MSW for tests
status: accepted
updated: 2026-07-08
parents: [AYD-001@context]
related: [TDR-002, TDR-004]
superseded_by: null
---

# TDR-005: Vitest + React Testing Library + MSW for tests

## Context
Conventions (CONV §B.2) require a test per SPEC acceptance criterion, AAA structure,
and mocking **only at the boundary** (network, db, clock) — never the unit under test.
For the web part that boundary is the **network** (the api). The tests must exercise
real components and hooks while faking the AYD-001 HTTP responses (success envelopes,
`422`/`409`/`404` errors, pagination).

## Decision
- **Vitest** as the test runner (native to Vite, Jest-compatible API).
- **React Testing Library** to render and drive components as a user would (queries by
  role/label, user-event), asserting observable output.
- **MSW (Mock Service Worker)** to intercept `/api/...` calls at the network boundary
  and return AYD-001-shaped responses. This keeps components, the typed client
  (TDR-002), and form logic (TDR-004) unmocked — only the api is faked.

## Alternatives & trade-offs
- **Jest**: needs extra config to align with Vite/ESM; Vitest is the native fit.
- **Mocking the fetch client / hooks directly**: violates CONV §B.2 (mocks the unit
  under test); MSW mocks the true boundary, so client parsing and error mapping stay
  under test.
- **E2E (Playwright/Cypress) as the primary suite**: valuable later, but heavier and
  slower than needed to cover the SPEC's acceptance criteria for the MVP; can be added
  when the purchase flow (RF-04) justifies full-stack E2E.
