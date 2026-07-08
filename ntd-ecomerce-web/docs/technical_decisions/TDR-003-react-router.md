---
id: TDR-003
type: tdr
title: React Router for client-side routing
status: accepted
updated: 2026-07-08
parents: [AYD-001@context]
related: [TDR-001]
superseded_by: null
---

# TDR-003: React Router for client-side routing

## Context
As an SPA (TDR-001), the web part needs client-side navigation across distinct views.
AYD-001 already implies a catalog list plus create/edit routes, and later AYDs add a
storefront/search and a purchase flow — so the routing choice must accommodate nested
layouts (an operator area and a customer-facing area) without a rewrite.

## Decision
- **React Router** (`react-router-dom`) as the router.
- Routes for AYD-001: `/products` (list), `/products/new` (create),
  `/products/:id/edit` (edit). A root layout hosts the shared shell (navigation,
  query client). Later AYDs add sibling routes (e.g. `/`, `/search`, `/checkout`)
  under the same layout tree.
- nginx (TDR-001) uses an SPA fallback (`try_files ... /index.html`) so deep links
  resolve to the client router.

## Alternatives & trade-offs
- **TanStack Router**: type-safe routing, but newer and heavier to adopt; React Router
  is the ecosystem default and enough for the MVP's route count.
- **No router (conditional rendering)**: breaks deep links, the back button, and
  bookmarkable edit URLs — unacceptable even for the admin UI.
- **Next.js file routing**: rejected with the framework itself in TDR-001.
