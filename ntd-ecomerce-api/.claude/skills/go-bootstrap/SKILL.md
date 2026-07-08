---
name: go-bootstrap
description: Wire a new clean-arch feature into the app — DI and route registration under internal/bootstrap/, internal/bootstrap/registry/, and cmd/api/main.go. Use when adding a feature's Setup() wiring or a repository accessor to the registry. Enforces the conventions below.
---

# Go Bootstrap & Dependency Injection

Authoritative description of the wiring layer that ties handler + usecase + repository
together for a feature. Read this before adding a new clean-arch feature end-to-end or
touching `cmd/api/main.go`. Architecture rationale: TDR-004; stack: TDR-001/002/003.

## When to use

- Adding `internal/bootstrap/{feature}/setup.go` for a new feature.
- Adding a new repository accessor to `internal/bootstrap/registry/registry.go`.
- Touching `cmd/api/main.go` (config, logger, DB, migrations, gin setup).

## The pieces and how they fit

```
cmd/api/main.go
  run()
    load config from env            -> DB DSN, port
    configureLogger()               -> slog
    openDB()                        -> *gorm.DB
    runMigrations()                 -> golang-migrate over embed.FS (TDR-003), before serving
    setupGin(logger)                -> *gin.Engine (recovery, logging, CORS, /healthz)
    bootstrap.SetupComponents(r, db)
      -> {feature}.Setup(r, reg) per feature:
           reg.Get{X}Repository() -> usecase.New{Entity}(...) -> api.New{Entity}Handlers(r, &svc)
    start http server + graceful shutdown
```

There is **no authentication** in the MVP (guest-only scope, see requirements.md);
all routes are public. If auth ever lands, it becomes an AYD first — don't add
middleware tiers here preemptively.

## Non-negotiable rules

**`cmd/api/main.go`**
- Keep `run()` as orchestration only: config, logger, DB, migrations, `setupGin`,
  `bootstrap.SetupComponents`, serve. Pull any standalone concern into its own
  top-level function instead of growing `run()`.
- Migrations run **before** the server starts listening; a migration failure aborts boot.

**`internal/bootstrap/setup.go`**
- One `SetupComponents(r *gin.Engine, db *gorm.DB)` that creates a single
  `registry.NewRegistry(db)` and calls each feature's `{feature}.Setup(r, reg)`.
  Append new features at the end; imports stay goimports-sorted.

**`internal/bootstrap/registry/registry.go`**
- One unexported pointer field + one `Get{X}Repository()` per repository, lazy
  nil-check-then-construct:
  ```go
  func (r *Registry) GetProductRepository() *repository.ProductRepository {
      if r.productRepository == nil {
          r.productRepository = repository.NewProductRepository(r.db)
      }
      return r.productRepository
  }
  ```
- `Registry` is process-lifetime — never stash per-request state on it; request data
  flows through `ctx`.
- Composite dependencies call the other `Get*` methods, never reach into fields.
- Stateless external clients (e.g. the fake payment gateway) are **not**
  registry-backed — construct them inside the feature's `Setup()`.

**`internal/bootstrap/{feature}/setup.go`**
- `package {feature}`, one exported `Setup(r *gin.Engine, registry *registry.Registry)`.
- Body shape: pull dependencies from the registry (or construct a gateway inline),
  build usecase(s) — innermost first when one depends on another — then call the
  handler constructor. If the usecase uses pointer receivers, pass `&service`.

## Adding a brand-new feature, end to end

1. `internal/domain/{feature}.go` — entity + validation (no GORM/JSON-framework deps).
2. `internal/infrastructure/repository/{feature}_repository.go` + model + `FromDomain`/`ToDomain`.
3. `internal/usecase/{feature}_usecase.go` — see the `go-usecases` skill.
4. `internal/infrastructure/api/{feature}_api.go` — see the `go-api-handlers` skill.
5. Registry field + `Get{Feature}Repository()`.
6. `internal/bootstrap/{feature}/setup.go` + the call in `SetupComponents`.

## Anti-patterns (reject these)

| Don't | Do |
|---|---|
| Wire a feature inline in `cmd/api/main.go` | `internal/bootstrap/{feature}/setup.go`, called from `SetupComponents` |
| Registry field without the lazy getter | Always pair field + `Get{X}Repository()` |
| `db.AutoMigrate(...)` anywhere | Schema changes are golang-migrate SQL files (TDR-003) |
| Memoize an external gateway on the Registry | Construct it in the feature's `Setup()` |
| Store request-scoped data on the Registry | Pass it through `ctx` |

## Run & verify

```bash
go build ./...
make linter
docker compose up   # confirm boot, migrations, and that the new route responds
```
