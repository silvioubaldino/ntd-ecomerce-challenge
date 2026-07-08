---
id: TDR-004
type: tdr
title: React Hook Form + Zod for forms and client-side validation
status: accepted
updated: 2026-07-08
parents: [AYD-001@context]
related: [TDR-002]
superseded_by: null
---

# TDR-004: React Hook Form + Zod for forms and client-side validation

## Context
The Product create/edit form (AYD-001) collects the `ProductInput` fields with
non-trivial rules the api enforces server-side: required strings with length bounds,
`price`/`weight_kg` as **non-negative decimal strings**, `stock` as a non-negative
integer. The client should validate up front for fast feedback while still surfacing
the api's authoritative errors — the `422 validation_error` `details` map and
`409 sku_already_exists` (TDR-002).

## Decision
- **React Hook Form** manages form state and submission.
- **Zod** defines a `ProductInput` schema that mirrors the AYD-001 rules, validated via
  the RHF resolver. Decimal fields are validated **as strings** (regex/`refine` for
  non-negative decimals) and sent as strings — never parsed to `number` (keeps
  contract fidelity, TDR-002).
- **Server errors win**: on a `422`, map `error.details` (field → problem code) onto
  RHF field errors; on a `409 sku_already_exists`, attach the error to the `sku`
  field. The api remains the source of truth; client validation is a UX layer, not a
  replacement.

## Alternatives & trade-offs
- **Formik + Yup**: comparable, but RHF has better performance (uncontrolled inputs)
  and Zod gives shared TS types from one schema.
- **Uncontrolled form + manual validation**: duplicates the rules by hand and is
  error-prone as fields grow (CSV import, purchase forms later).
- **Client-only validation**: unsafe — the api owns the rules (uniqueness needs the
  server anyway), so we always reconcile with its response.
