---
id: TDR-001
type: tdr
title: Vite + React + TypeScript SPA, served by nginx with an /api reverse proxy
status: accepted
updated: 2026-07-08
parents: [AYD-001@context]
related: [TDR-002, TDR-003, TDR-004, TDR-005]
superseded_by: null
---

# TDR-001: Vite + React + TypeScript SPA, served by nginx with an /api reverse proxy

## Context
The web part must deliver the UIs for the MVP (Product CRUD now — AYD-001@context —
then search RF-03 and purchase RF-04) as a **pure consumer** of the Go api, which
**owns the contracts** (see `architecture.md`). It also has to run as a Docker
container (RNF-01). Two forces shape the decision: (1) how the SPA is built and
served, and (2) how the browser reaches the api without a cross-origin problem — the
api currently ships **no CORS middleware** and we prefer not to change the contract
just to unblock the frontend.

## Decision
- **Vite + React 18 + TypeScript** as a **single-page application** (client-side
  only). Module/app name `ntd-ecomerce-web`.
- **Build & serve in Docker**: multi-stage image — `vite build` produces static
  assets, served by **nginx**. Added to the root `docker-compose.yml` as the `web`
  service (RNF-01).
- **Same-origin via reverse proxy** (no CORS, contract untouched):
  - **Production:** nginx serves the SPA and proxies `/api/*` to the api container.
  - **Dev:** the Vite dev server proxies `/api/*` to `http://localhost:8080`.
  - The frontend always calls **relative** `/api/...` URLs — the api base is never
    hardcoded and there is no cross-origin request.

## Alternatives & trade-offs
- **Next.js**: SSR/SSG, file-based routing, API routes and server components. But the
  topology already assigns contract ownership to the Go api, so Next's server-side
  features would sit idle or duplicate the backend, while adding a Node runtime to the
  Docker footprint. SEO/SSR gains are marginal for the MVP's (simulated) storefront.
- **Create React App**: effectively unmaintained; slower dev loop than Vite.
- **Enabling CORS on the api instead of proxying**: would push a frontend concern into
  the api's public surface (allowed origins as config) for no user-facing value; the
  reverse proxy keeps it a web-local concern and the api same-origin.
