---
id: SPEC-004
type: spec
status: done
updated: 2026-07-09
parents: [AYD-004@context]
related: [GLO]
---

# SPEC-004: Cart (api) — what + how

> Implements the AYD-004@context contract: a guest Cart resource that groups Products
> with quantities, validates quantity against the Product's current stock (RN-03), and
> returns line subtotals + a cart total. Reuses the AYD-001 Product model, the
> decimal-as-string rule and the `{ "error": { code, message, details? } }` envelope.
> Stock is **checked, not reserved** (the decrement happens at checkout, AYD-005).

## What (goal)
Deliver the api side of RF-04's grouping half: `POST /carts`, `GET /carts/{cart_id}`,
`POST /carts/{cart_id}/items` (increment), `PUT /carts/{cart_id}/items/{product_id}`
(set absolute) and `DELETE /carts/{cart_id}/items/{product_id}`, always returning the
full priced Cart. `unit_price`/`subtotal`/`total` are indicative (mirror the live
Product price) and travel as decimal strings.

## Acceptance criteria
```gherkin
Scenario: create an empty cart
  When POST /carts is called
  Then it responds 201 with a new cart_id, items: [] and total: "0"

Scenario: get an existing cart
  Given a cart with items exists
  When GET /carts/{cart_id} is called
  Then it responds 200 with items (each with sku/name/unit_price/quantity/subtotal) and total

Scenario: get a missing cart
  When GET /carts/{cart_id} is called with an unknown cart_id
  Then it responds 404 with error code "cart_not_found"

Scenario: add a product to a cart
  Given a cart exists and a product with stock 10
  When POST /carts/{cart_id}/items { product_id, quantity: 2 } is called
  Then it responds 200 and the cart contains that product with quantity 2

Scenario: adding the same product increments the quantity
  Given a cart already holds product P with quantity 2 and P has stock 10
  When POST /carts/{cart_id}/items { product_id: P, quantity: 3 } is called
  Then it responds 200 and the line quantity for P is 5

Scenario: add rejects quantity below 1
  Given a cart exists
  When POST /carts/{cart_id}/items { product_id, quantity: 0 } is called
  Then it responds 422 with code "validation_error" and details { quantity: "must_be_positive_integer" }

Scenario: add to a missing cart
  When POST /carts/{cart_id}/items is called with an unknown cart_id
  Then it responds 404 with error code "cart_not_found"

Scenario: add a missing product
  Given a cart exists
  When POST /carts/{cart_id}/items is called with an unknown product_id
  Then it responds 404 with error code "product_not_found"

Scenario: add beyond available stock
  Given a cart exists and product P has stock 3
  When POST /carts/{cart_id}/items { product_id: P, quantity: 5 } is called
  Then it responds 409 with code "insufficient_stock" and details { product_id, requested: "5", available: "3" }
  And the cart is unchanged

Scenario: incrementing beyond available stock
  Given a cart holds product P with quantity 2 and P has stock 3
  When POST /carts/{cart_id}/items { product_id: P, quantity: 2 } is called
  Then it responds 409 "insufficient_stock" with details requested: "4", available: "3"
  And the cart is unchanged

Scenario: set the absolute quantity of a line
  Given a cart holds product P with quantity 2 and P has stock 10
  When PUT /carts/{cart_id}/items/{P} { quantity: 5 } is called
  Then it responds 200 and the line quantity for P is 5

Scenario: set rejects quantity below 1
  Given a cart holds product P
  When PUT /carts/{cart_id}/items/{P} { quantity: 0 } is called
  Then it responds 422 "validation_error" (to remove a line, use DELETE)

Scenario: set on a line that is not in the cart
  Given a cart exists without product P
  When PUT /carts/{cart_id}/items/{P} { quantity: 1 } is called
  Then it responds 404 with error code "item_not_found"

Scenario: set beyond available stock
  Given a cart holds product P and P has stock 3
  When PUT /carts/{cart_id}/items/{P} { quantity: 5 } is called
  Then it responds 409 "insufficient_stock" with details requested: "5", available: "3"

Scenario: remove a line
  Given a cart holds product P
  When DELETE /carts/{cart_id}/items/{P} is called
  Then it responds 200 with the cart after removal (may be empty)

Scenario: remove a line that is not in the cart
  Given a cart exists without product P
  When DELETE /carts/{cart_id}/items/{P} is called
  Then it responds 404 with error code "item_not_found"

Scenario: remove from a missing cart
  When DELETE /carts/{cart_id}/items/{P} is called with an unknown cart_id
  Then it responds 404 with error code "cart_not_found"

Scenario: cart total reflects the live product price
  Given a cart holds product P (price "10.00") with quantity 3
  When GET /carts/{cart_id} is called
  Then the line subtotal is "30.00" and total is the sum of subtotals
```

## How (approach)
Layered like the Product resource (domain → usecase → repository → api, wired in
`bootstrap`). Two new tables (`carts`, `cart_items`) with `(cart_id, product_id)` as the
composite primary key so a Product appears at most once per Cart. The repository persists
only `product_id` + `quantity` per line; the usecase **enriches** each line from the live
Product (`sku`, `name`, `unit_price = Product.price`), computes `subtotal = unit_price *
quantity` and `total = Σ subtotal`. `POST /items` reads the current line quantity and
persists `resulting = current + requested`; `PUT` persists the absolute value — both
validate `resulting <= Product.stock` (RN-03) before any write, so a rejected op leaves
the cart unchanged. Stock is never reserved.

Notes on the contract's `details`: the shared error envelope's `details` is
`map[string]string`, so `insufficient_stock` reports `requested`/`available` as decimal
strings of integers (e.g. `"5"`, `"3"`) and `product_id` as the uuid string. `requested`
is the **resulting** line quantity that would exceed stock (for `POST`, current +
increment; for `PUT`, the set value).

## Steps
1. `internal/domain/cart.go`: `Cart` (`id`, `items`, `total`, timestamps) and `CartItem`
   (`product_id`, `sku`, `name`, `unit_price`, `quantity`, `subtotal`); `AddItemInput`
   (`product_id`, `quantity`) and `SetItemInput` (`quantity`) with `Validate()`
   returning `{quantity: "must_be_positive_integer"}` when `quantity < 1`; `Recalculate()`
   filling each `subtotal` and the cart `total`.
2. `internal/usecase/usecase_errors.go`: add `ErrInvalidCartInput`, `ErrInsufficientStock`,
   `ErrItemNotFound`.
3. `internal/domain/errors.go`: add `WrapConflictDetails(err, code, message, details)`
   (conflict kind + details), used for `insufficient_stock`.
4. `internal/usecase/cart_usecase.go`: `Cart` usecase over a `CartRepository`
   (`Create`, `FindByID`, `SaveItem` upsert-absolute, `RemoveItem`) and a `ProductReader`
   (`FindByID`). Implements `Create`, `Get`, `AddItem`, `SetItem`, `RemoveItem` with the
   ordering: validate input → load cart (`cart_not_found`) → (item/product checks) →
   stock check (`insufficient_stock`) → persist → reload+enrich.
5. `internal/infrastructure/repository/cart_repository.go`: `cartModel`/`cartItemModel`,
   `Create`, `FindByID` (loads items ordered by `created_at asc`), `SaveItem`
   (`ON CONFLICT (cart_id, product_id) DO UPDATE quantity`), `RemoveItem`
   (`item_not_found` when no row); item writes bump `carts.updated_at` in a transaction.
6. `internal/infrastructure/api/cart_api.go`: `CartUsecase` interface + handlers/routes;
   invalid path uuid → `422 validation_error`, malformed body → `422`.
7. `migrations/000002_create_carts.{up,down}.sql`: create/drop `carts` and `cart_items`
   (FK to `products`, `ON DELETE CASCADE` from cart, `CHECK (quantity >= 1)`).
8. `internal/bootstrap/cart/setup.go` + `registry.GetCartRepository` + register in
   `bootstrap/setup.go`.
9. **Verify:** `make build`, `make linter`, `make test`. Add tests per acceptance
   criterion (AAA). Update `docs/changelog.md`; mark this spec `done`.

## Affected files
- `ntd-ecomerce-api/internal/domain/cart.go` (+ `cart_test.go`)
- `ntd-ecomerce-api/internal/domain/errors.go`
- `ntd-ecomerce-api/internal/usecase/usecase_errors.go`
- `ntd-ecomerce-api/internal/usecase/cart_usecase.go` (+ `cart_usecase_test.go`, `cart_mock_test.go`)
- `ntd-ecomerce-api/internal/infrastructure/repository/cart_repository.go` (+ `cart_repository_test.go`)
- `ntd-ecomerce-api/internal/infrastructure/api/cart_api.go` (+ `cart_api_test.go`, `cart_mock_test.go`)
- `ntd-ecomerce-api/migrations/000002_create_carts.{up,down}.sql`
- `ntd-ecomerce-api/internal/bootstrap/cart/setup.go`, `internal/bootstrap/registry/registry.go`, `internal/bootstrap/setup.go`
- `ntd-ecomerce-api/docs/changelog.md`

## Tests
- **Acceptance:** one handler-level test per Gherkin scenario (mocked `CartUsecase`) for
  status + error code/details; usecase tests (mocked `CartRepository` + `ProductReader`)
  for increment/set/remove, stock checks (RN-03), and the 404/422/409 branches.
- **Unit:** `domain` — `Validate()` (`quantity < 1`), `Recalculate()` subtotal/total.
  `repository` — `FindByID` (cart + items, `cart_not_found`), `SaveItem` upsert,
  `RemoveItem` (`item_not_found`) with sqlmock. Table-driven, AAA, testify.

## Checklist
- [x] `POST /carts` → 201 empty cart; `GET` → 200 / 404 cart_not_found
- [x] `POST /items` increments; `PUT` sets absolute; `DELETE` removes (200, may be empty)
- [x] RN-03 enforced pre-write: `409 insufficient_stock` (+ product_id/requested/available), cart unchanged
- [x] `422 validation_error` (+ details) for quantity < 1 / malformed body; `404` cart/product/item_not_found
- [x] Decimals (`unit_price`, `subtotal`, `total`) as strings; `(cart_id, product_id)` unique
- [x] Build, linter, tests green; changelog updated; spec `done`
