---
id: AYD-004
type: design
status: draft
updated: 2026-07-08
parents: [RF-04]
children: []          # SPECs to be generated later: [SPEC-004@api, SPEC-004@web]
related: [GLO, AYD-001]
---

# AYD-004: Cart (multi-item grouping)

> Fourth feature and the first half of the purchase flow (RF-04). It introduces the
> **Cart** — the grouping that lets a customer collect **more than one Product** (with
> quantities) before buying. Checkout (turning a Cart into an Order) is designed
> separately in **AYD-005**. Reuses the **Product** model, the decimal-as-string rule and
> the error envelope from AYD-001.

## Goal
Meets the grouping part of **RF-04**: a customer can add one or more Products, with
quantities, to a Cart through the web UI, backed by the API. Outcome: an end-to-end Cart
flow (web → api → db) that holds the selected Products and always reflects a validated,
priced grouping ready to be checked out (AYD-005).

## Affected parts
| Part | Role in this feature | Generated SPEC |
|-------|---------------------|-------------|
| api | Owns the Cart resource: create cart, add/update/remove items, validate quantity against Product stock (RN-03), compute line subtotals and cart total | SPEC-004@api |
| web | Storefront cart UI: "add to cart" from catalog/search, cart view with quantity edit / remove, running total, "checkout" entry point | SPEC-004@web |

## Contract (source of truth)

Reuses AYD-001 conventions: JSON `snake_case`; decimals (`unit_price`, `subtotal`,
`total`) travel as **strings** (e.g. `"89.99"`); error envelope
`{ "error": { "code", "message", "details"? } }`.

### Guest identity
The MVP has no authentication (REQ). A Cart is identified by a **server-generated
`cart_id` (uuid)** returned on creation; the web client persists it locally
(e.g. `localStorage`) and sends it on every cart/checkout call. There is no cross-device
or cross-session sharing — that is out of scope (REQ MVP scope).

### Resources
```
Cart
{
  id:         string (uuid, server-generated)
  items:      [CartItem]
  total:      string decimal >= 0     // sum of item.subtotal, convenience field
  created_at: string (ISO 8601)
  updated_at: string (ISO 8601)
}

CartItem
{
  product_id: string (uuid)
  sku:        string                  // denormalized from Product for display
  name:       string                  // denormalized from Product for display
  unit_price: string decimal >= 0     // CURRENT Product price (indicative, not binding)
  quantity:   integer >= 1
  subtotal:   string decimal >= 0     // unit_price * quantity
}
```

`unit_price`/`total` in the Cart are **indicative** (they mirror the live Product price
for display). The **binding** price snapshot happens only at checkout, in the Order
(AYD-005, RN-04). If a Product's price changes, the Cart reflects the new price on the
next read.

### Endpoints
```
POST /carts
res 201: Cart            // new, empty cart

GET /carts/{cart_id}
res 200: Cart
errors: [ 404 cart_not_found ]

POST /carts/{cart_id}/items          // add a Product; if already present, INCREMENTS quantity
req:  { product_id: string, quantity: integer >= 1 }
res 200: Cart
errors: [ 404 cart_not_found, 404 product_not_found,
          422 validation_error, 409 insufficient_stock ]

PUT /carts/{cart_id}/items/{product_id}   // SET the absolute quantity for a line
req:  { quantity: integer >= 1 }
res 200: Cart
errors: [ 404 cart_not_found, 404 item_not_found,
          422 validation_error, 409 insufficient_stock ]

DELETE /carts/{cart_id}/items/{product_id}
res 200: Cart            // cart after removal (may be empty)
errors: [ 404 cart_not_found, 404 item_not_found ]
```

- `422 validation_error` carries `details`, e.g. `{ "details": { "quantity": "must_be_positive_integer" } }`.
- `409 insufficient_stock` carries `details` with what was requested vs available, e.g.
  `{ "details": { "product_id": "<uuid>", "requested": 5, "available": 3 } }` (RN-03).
- Resulting `quantity` after `POST` (increment) or `PUT` (set) must satisfy RN-03 against
  the Product's current `stock`; otherwise `409 insufficient_stock` and the cart is
  unchanged.
- Adding a `quantity` of `0` is invalid (`422`); to remove a line use `DELETE`.

## Affected domain model
- **Cart** (new): `id`, timestamps. Guest — no user reference.
- **Cart Item** (new): belongs to a Cart; `(cart_id, product_id)` is unique (a Product
  appears at most once per Cart, its `quantity` accumulates). References **Product**
  (AYD-001) by `product_id`.
- **Product** (AYD-001): unchanged schema; read for price/name/sku and its `stock` is the
  ceiling for RN-03. Stock is **not** reserved by the Cart — it is only checked; the
  actual decrement happens at checkout (AYD-005).

## Flow
```mermaid
sequenceDiagram
    participant customer as Customer (browser)
    participant web
    participant api
    participant db
    customer->>web: "add to cart" (product, qty)
    web->>api: POST /carts (first time only) -> cart_id
    web->>api: POST /carts/{cart_id}/items { product_id, quantity }
    api->>db: load Product; check stock (RN-03)
    api->>db: upsert cart_item (increment qty)
    api-->>web: 200 Cart (items, subtotals, total)
    web-->>customer: render cart + running total
    customer->>web: change qty / remove
    web->>api: PUT|DELETE /carts/{cart_id}/items/{product_id}
    api-->>web: 200 Cart
    customer->>web: proceed to checkout
    note over web,api: checkout continues in AYD-005 (POST /orders)
```

## Out of scope / open questions
- **Out:** checkout / Order creation — that is **AYD-005** (`POST /orders`), which
  consumes the `cart_id`.
- **Out:** stock **reservation** — the Cart only validates against current stock (RN-03);
  it never holds/reserves units. Overselling between add-to-cart and checkout is resolved
  at checkout time (AYD-005 re-checks stock, `409 insufficient_stock`).
- **Out:** cross-device / persistent / shared carts, merging carts, discounts, taxes,
  shipping (REQ MVP scope).
- **Open (cart lifetime):** guest carts may be pruned by a TTL; the retention policy is an
  api-local concern (SPEC-004@api / a TDR@api), not part of the contract. A pruned cart
  reads as `404 cart_not_found`.
- **Open (alternative considered):** a purely client-side cart with checkout sending the
  full item list in one `POST /orders`. Rejected for the MVP so quantity/stock validation
  (RN-03) and pricing live in one authoritative place (the api) and the web stays thin.
