# 🛒 NTD E-commerce Challenge

Solution to the NTD e-commerce code challenge: a guest-checkout storefront with
Product catalog management, CSV bulk import, search, and a cart-based purchase flow.
Payment is simulated — no real payment provider is integrated.

**Monorepo**, three parts + a local database:

```
ntd-ecomerce-context/   shared docs (single source of truth)
ntd-ecomerce-api/       backend — Go + Gin + GORM + Postgres, owns the contracts
ntd-ecomerce-web/       frontend — React + TypeScript SPA, consumes the api
db (local)              Postgres 17, run via Docker
```



## Features implemented

- **Product CRUD** — create, view, update, delete, through both the API and the web UI.
- **CSV bulk import** — upload a CSV to create Products in bulk, with per-row validation
(see [CSV import](#csv-import) below).
- **Product search** — keyword search over the catalog, API-backed, from the web UI.
- **Cart & purchase** — a guest Cart holds Products with quantities; checkout creates
an Order with a simulated payment. Stock is validated and decremented atomically at
checkout.
- Runs fully as a Docker container stack (`docker compose up`).



## How to run locally



### Option A — Docker (recommended)

Prerequisites: Docker + Docker Compose.

```bash
git clone https://github.com/silvioubaldino/ntd-ecomerce-challenge.git
cd ntd-ecomerce-challenge
docker compose up --build
```

This starts three services:


| Service | URL                                            | Notes                                                        |
| ------- | ---------------------------------------------- | ------------------------------------------------------------ |
| `web`   | [http://localhost:5173](http://localhost:5173) | React SPA, served by nginx, proxies `/api/*` to the api      |
| `api`   | [http://localhost:8080](http://localhost:8080) | Go REST API; health check at `GET /healthz`                  |
| `db`    | localhost:5432                                 | Postgres 17; schema migrations run automatically on api boot |


No manual migration or seed step is required — the api applies its SQL migrations at
startup before it starts serving requests.

### Option B — Run each part locally (without Docker)

**API** (needs a local Postgres — you can also just run `docker compose up db`):

```bash
cd ntd-ecomerce-api
cp .env.example .env         # adjust DATABASE_URL if not using the default local Postgres
export $(cat .env | xargs)   # the api reads plain env vars, no .env autoloading
go run ./cmd/api
```

**Web** (dev server, proxies `/api` to `http://localhost:8080`):

```bash
cd ntd-ecomerce-web
npm install
npm run dev
# open http://localhost:5173
```

Tests:

```bash
cd ntd-ecomerce-api && make test     # go test -race -cover ./...
cd ntd-ecomerce-web  && npm test     # vitest
```



## CSV import

The reference sample CSV ([`NTD Code Challenge E-Commerce.csv`](./NTD%20Code%20Challenge%20E-Commerce.csv),
kept at the project root) was **downloaded on 2026-07-08**. Use it via
`/products/import` in the web UI, or `POST /products/import` directly, to try the
bulk-import flow end to end.

Expected columns: `name, sku, description, category, price, stock, weight_kg`. Each row
is validated independently on import — rows with missing required fields, a
non-numeric `price`/`stock`/`weight_kg`, negative `stock`, a malformed or duplicate
`sku`, or unsafe content are rejected and reported rather than silently imported; valid
rows in the same file are still imported. The sample file intentionally contains a mix
of valid and invalid rows to exercise this validation.

## Approach

The build followed a spec-driven flow rather than jumping straight to code:

1. **Requirements + glossary** first (`ntd-ecomerce-context/docs/requirements.md`), so
  every feature and every piece of code uses the same vocabulary (Product, Cart, Cart
   Item, Order, Order Item, SKU).
2. **Design (AYD) per feature**, written *before* implementation, fixing the API
  contract (endpoints, payloads, error shapes) and the domain model shared between
   backend and frontend. The backend owns the contract; the frontend consumes it.
3. **Spec per part** (api/web) translating the design into concrete acceptance
  criteria and steps.
4. **Implementation**, with local technical decisions recorded (see below) whenever a
  choice wasn't obvious from the design alone.

This made it possible to use AI assistance deliberately at each stage — drafting
requirements and contracts, generating boilerplate, writing tests — while keeping
human review and decision-making at the design/technical-decision layer, rather than
letting the AI make undocumented architectural choices. Full design docs, specs, and
decision records are kept in the repo under each part's `docs/` folder for anyone who
wants the detailed trail.

## Key decisions and alternatives considered



### Backend (`ntd-ecomerce-api`)

- **Go 1.25 + Gin.** Chosen over `chi`/`net/http` (more boilerplate for binding,
middleware, and route groups) and `echo` (no material advantage over Gin) — this is
the operator's strongest backend stack, which lowered delivery risk across four
features (CRUD, CSV import, search, purchase).
- **GORM (Postgres) +** `shopspring/decimal` for data access, with the **schema owned
by SQL migrations, not** `AutoMigrate`**.** Money and weight are stored as `NUMERIC` and
travel over JSON as **strings**, never floats, to avoid precision loss. Considered and
rejected: `pgx` + `sqlc` (more codegen ceremony than the MVP needed), `ent` (steep
learning curve for this scope), and `AutoMigrate` (hides schema evolution and can't
express constraints precisely).
- **golang-migrate, embedded and run at API boot.** Versioned SQL migrations ship
inside the binary and apply automatically before the server starts serving traffic,
so `docker compose up` alone yields a ready schema — no extra container or manual
step. Considered and rejected: Postgres `docker-entrypoint-initdb.d` init scripts (no
versioning, only run on an empty volume) and a separate migrate step/container (an
extra moving part not justified for a single-API stack).
- **Layered architecture with a dedicated** `domain` **package** (`domain → usecase → infrastructure`, dependencies declared as narrow interfaces at the point of use),
chosen over package-per-feature (cross-feature calls, e.g. purchase → product stock,
created import cycles) and full hexagonal/ports-and-adapters (more ceremony than a
four-feature MVP needs).



### Frontend (`ntd-ecomerce-web`)

- **Vite + React 18 + TypeScript SPA**, built and served by **nginx** in Docker, with a
**same-origin reverse proxy to** `/api/`* (nginx in prod, Vite dev-server proxy in
dev) instead of enabling CORS on the api — keeps the api's public contract untouched
and cross-origin concerns entirely on the frontend side. Considered and rejected:
Next.js (its server-side features would sit idle since the Go api already owns the
contract, and it adds a Node runtime to the Docker footprint) and Create React App
(effectively unmaintained).
- **TanStack Query** over a small typed `fetch` client for all server state (lists,
reads, mutations with cache invalidation), instead of Redux Toolkit/RTK Query (a
global store is overkill for a stateless, guest-only MVP) or raw `fetch` in
components (would hand-roll loading/error/refetch on every screen).
- **React Router** for navigation, chosen over TanStack Router (heavier to adopt for
the MVP's route count) and no router at all (breaks deep links, back button, and
bookmarkable edit URLs).
- **React Hook Form + Zod** for forms, mirroring the API's validation rules client-side
for fast feedback while still treating the API's `422`/`409` responses as the source
of truth — never a client-only validation replacement.
- **Vitest + React Testing Library + MSW**, mocking only at the network boundary (the
api), so components, the typed client, and form logic stay exercised end-to-end in
tests. Full E2E (Playwright/Cypress) was considered but judged heavier than needed to
cover the MVP's acceptance criteria.
- **Small in-house design system on Tailwind tokens**, no component library — chosen
over MUI/Chakra (heavier bundles, generic look, a second styling paradigm) and
shadcn/ui (closest fit, but pulls in Radix and a codegen workflow for a surface small
enough to hand-roll).



## Notes on AI use

AI assistance was used throughout (requirements drafting, design docs, spec
authoring, and implementation), guided and reviewed at each step rather than applied
as a single end-to-end generation. Per the challenge instructions, code does not carry
explanatory comments.

## Repository

[https://github.com/silvioubaldino/ntd-ecomerce-challenge](https://github.com/silvioubaldino/ntd-ecomerce-challenge)