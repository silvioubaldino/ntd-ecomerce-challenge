---
id: SPEC-005
type: spec
status: done
updated: 2026-07-09
parents: [AYD-005@context]
related: [GLO]
---

# SPEC-005: Order & Checkout (api) — what + how

> Implements the AYD-005@context contract: checkout turns a guest Cart (AYD-004) into a
> confirmed, immutable Order running a **simulated** payment. It re-checks stock (RN-03),
> snapshots prices/sku/name and decrements stock atomically (RN-04), and consumes the
> Cart. Reuses the AYD-001 Product model, the decimal-as-string rule and the
> `{ "error": { code, message, details? } }` envelope. `product_id` on an Order Item is a
> **soft reference** (no FK), so Products stay hard-deletable.

## What (goal)
Deliver the api side of RF-04's checkout half: `POST /orders` (create an Order from a
Cart) and `GET /orders/{order_id}` (read an Order). The Order carries the customer
contact, the priced items (each snapshotting `sku`/`name`/`unit_price`), the immutable
`total`, and an always-approved simulated `payment`. `unit_price`/`subtotal`/`total`
travel as decimal strings.

## Acceptance criteria
```gherkin
Scenario: create an order from a cart (happy path)
  Given a cart holds product P (price "10.00", stock 10) with quantity 3
  When POST /orders { cart_id, customer: { name, email } } is called
  Then it responds 201 with an Order (status "confirmed")
  And the order has one item snapshotting P's sku/name/unit_price "10.00", quantity 3, subtotal "30.00"
  And total is "30.00" and payment is { method: "simulated", status: "approved" }

Scenario: checkout a missing cart
  When POST /orders is called with an unknown cart_id
  Then it responds 404 with error code "cart_not_found"

Scenario: checkout an empty cart
  Given a cart exists with no items
  When POST /orders is called with that cart_id
  Then it responds 422 with error code "cart_empty"

Scenario: checkout with an invalid customer
  Given a cart holds a product
  When POST /orders is called with a blank name and a malformed email
  Then it responds 422 with code "validation_error" and details { name: "required", email: "invalid" }

Scenario: checkout when stock is insufficient (nothing committed)
  Given a cart holds product P with quantity 5 and P now has stock 3
  When POST /orders is called with that cart_id
  Then it responds 409 with code "insufficient_stock" and details keyed by product_id
  And no order is created, stock is unchanged and the cart still exists

Scenario: checkout decrements stock and consumes the cart
  Given a cart holds product P (stock 10) with quantity 3
  When POST /orders succeeds
  Then P's stock becomes 7
  And a later GET /carts/{cart_id} responds 404 "cart_not_found" (the cart cannot be checked out twice)

Scenario: get an existing order
  Given an order exists
  When GET /orders/{order_id} is called
  Then it responds 200 with the order, its items and totals

Scenario: get a missing order
  When GET /orders/{order_id} is called with an unknown order_id
  Then it responds 404 with error code "order_not_found"

Scenario: an order is immutable against later product changes
  Given an order snapshotted product P's price "10.00"
  When P's price or record later changes or is deleted
  Then the order's item unit_price and total stay "10.00" (snapshots hold no live reference)
```

## How (approach)
Layered like the Cart resource (domain → usecase → repository → api, wired in
`bootstrap`). The whole of checkout runs in a **single DB transaction** owned by the
repository (same spirit as `cart_repository` owning its transactions), because it must
load the cart, lock+recheck product stock, snapshot, decrement, insert the order and
delete the cart atomically with row locks. The usecase stays thin: it validates the
customer (`422 validation_error`) then delegates to `repo.Checkout(ctx, cartID, customer)`.

Transaction steps (RN-04): load the cart (`404 cart_not_found`); require items
(`422 cart_empty`); load the referenced products locking their rows
`FOR UPDATE` (gorm `clause.Locking{Strength: "UPDATE"}`); re-check each line's quantity
against the current stock (RN-03) collecting any shortfall into `insufficient_stock`
`details` (`409`, rollback, nothing committed); snapshot `sku`/`name`/`unit_price` from
each product and compute `subtotal = unit_price * quantity` and `total = Σ subtotal`
(RN-04); simulate payment (always approved); decrement each product's stock; insert the
order + items with `status "confirmed"`; delete the cart (cart_items cascade) so it can
never be checked out twice.

Notes on `details`: the shared envelope's `details` is `map[string]string`, so
`insufficient_stock` keys each offending line by its `product_id` string mapping to
`"requested=<q>, available=<stock>"`. `validation_error` reports `{"name":"required"}`
and/or `{"email":"invalid"}` per field, matching the cart/product validation detail style.

Domain model: `order_items.product_id` is a **soft reference** — no FK to `products`
(Products stay hard-deletable, the item keeps its own snapshot). `orders.id` /
`order_items.order_id` use a real FK with `ON DELETE CASCADE`.

## Steps
1. `internal/domain/order.go`: `Order` (`id`, `status`, `customer`, `items`, `total`,
   `payment`, `created_at`), `OrderItem` (`product_id`, `sku`, `name`, `unit_price`,
   `quantity`, `subtotal`), `Customer` (`name`, `email`), `Payment` (`method`, `status`);
   `CheckoutInput` (`cart_id`, `customer`); the `"confirmed"` / simulated-approved
   constants; `Customer.Validate()` returning `{name: "required"}` / `{email: "invalid"}`;
   `ApprovedPayment()` constructor; `Recalculate()` filling subtotals + total.
2. `internal/usecase/usecase_errors.go`: add `ErrInvalidCustomerInput`, `ErrCartEmpty`,
   `ErrOrderNotFound` (reuse existing `ErrInsufficientStock`).
3. `internal/usecase/order_usecase.go`: `Order` usecase over an `OrderRepository`
   (`Checkout`, `FindByID`). `Checkout` validates the customer then delegates the atomic
   work; `Get` → `FindByID`.
4. `internal/infrastructure/repository/order_repository.go`: `OrderRepository` with
   `Checkout` (the full `db.Transaction(...)`, locking products `FOR UPDATE`) and
   `FindByID` (`order_not_found`). Reuses the package `productModel`; reads
   `carts`/`cart_items` via lightweight row structs.
5. `internal/infrastructure/api/order_api.go`: `OrderUsecase` interface + handlers/routes
   `POST /orders`, `GET /orders/:order_id`; malformed body → `422`; invalid path uuid →
   `parseUUIDParam` (reused from `cart_api.go`).
6. `migrations/000003_create_orders.{up,down}.sql`: create/drop `orders` and `order_items`
   (order→items FK `ON DELETE CASCADE`; `product_id` soft ref with NO FK;
   `CHECK (quantity >= 1)`).
7. `internal/bootstrap/order/setup.go` + `registry.GetOrderRepository` + register in
   `bootstrap/setup.go`.
8. **Verify:** `make build`, `make linter`, `make test`. Add tests per acceptance
   criterion (AAA). Update `docs/changelog.md`; mark this spec `done`.

## Affected files
- `ntd-ecomerce-api/internal/domain/order.go` (+ `order_test.go`)
- `ntd-ecomerce-api/internal/usecase/usecase_errors.go`
- `ntd-ecomerce-api/internal/usecase/order_usecase.go` (+ `order_usecase_test.go`, `order_mock_test.go`)
- `ntd-ecomerce-api/internal/infrastructure/repository/order_repository.go` (+ `order_repository_test.go`)
- `ntd-ecomerce-api/internal/infrastructure/api/order_api.go` (+ `order_api_test.go`, `order_mock_test.go`)
- `ntd-ecomerce-api/migrations/000003_create_orders.{up,down}.sql`
- `ntd-ecomerce-api/internal/bootstrap/order/setup.go`, `internal/bootstrap/registry/registry.go`, `internal/bootstrap/setup.go`
- `ntd-ecomerce-api/docs/changelog.md`

## Tests
- **Acceptance:** one handler-level test per Gherkin scenario (mocked `OrderUsecase`) for
  status + error code/details; usecase tests (mocked `OrderRepository`) for the
  customer-validation branch and the delegated 404/422/409 branches.
- **Unit:** `domain` — `Customer.Validate()` (name/email), `Recalculate()` subtotal/total,
  `ApprovedPayment()`. `repository` — sqlmock for `Checkout` happy-path commit,
  `cart_not_found`, `cart_empty`, `insufficient_stock` rollback, and `FindByID` +
  `order_not_found`. Table-driven, AAA, testify.

## Checklist
- [x] `POST /orders` → 201 confirmed Order; `GET /orders/{id}` → 200 / 404 order_not_found
- [x] `404 cart_not_found` / `422 cart_empty` / `422 validation_error` (+ details) branches
- [x] RN-03 re-checked under row lock: `409 insufficient_stock` (+ details), nothing committed
- [x] RN-04 atomic: snapshot unit_price/sku/name, immutable total, stock decremented, cart consumed
- [x] `product_id` soft reference (no FK); order→items FK `ON DELETE CASCADE`
- [x] Decimals (`unit_price`, `subtotal`, `total`) as strings; payment always simulated/approved
- [x] Build, linter, tests green; changelog updated; spec `done`
