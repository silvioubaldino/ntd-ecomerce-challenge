---
id: TDR-002
type: tdr
title: TanStack Query over a typed fetch client for api server-state
status: accepted
updated: 2026-07-08
parents: [AYD-001@context]
related: [TDR-001, TDR-004]
superseded_by: null
---

# TDR-002: TanStack Query over a typed fetch client for api server-state

## Context
Every screen in the web part is backed by the api's REST resources (paginated lists,
single Product reads, create/update/delete mutations). This is **server state**:
cached, revalidated after writes, with loading/error phases. AYD-001@context also
fixes a specific wire shape the client must respect — decimals (`price`, `weight_kg`)
as **strings**, `snake_case` fields, the `{ error: { code, message, details? } }`
envelope, and `{ data, pagination }` list responses.

## Decision
- **Typed `fetch` client** (`src/api/`): thin wrapper around the browser `fetch`, all
  calls to relative `/api/...` URLs (TDR-001). It exposes typed functions
  (`listProducts`, `getProduct`, `createProduct`, `updateProduct`, `deleteProduct`)
  and parses the AYD-001 error envelope into a typed `ApiError` carrying `code` and
  `details` so the UI can map field errors and status codes.
- **TypeScript types mirror the AYD-001 contract**: `Product`, `ProductInput`,
  `Pagination`, `ApiError`. Decimal fields are typed as `string` and never coerced to
  `number` for storage/transport.
- **TanStack Query (React Query)** for cache, request state, and invalidation:
  `useQuery` for lists/reads (query keys include pagination), `useMutation` for writes
  that invalidate the product list/detail keys on success.

## Alternatives & trade-offs
- **axios**: fine, but `fetch` + a small wrapper is enough for a same-origin JSON api
  and keeps the dependency surface minimal.
- **Redux Toolkit / RTK Query**: RTK Query overlaps TanStack Query; a global Redux
  store is overkill for a stateless, guest-only MVP with no meaningful client state.
- **Plain fetch in components (no cache layer)**: would hand-roll loading/error/
  refetch-after-mutation on every screen — exactly what TanStack Query removes.
