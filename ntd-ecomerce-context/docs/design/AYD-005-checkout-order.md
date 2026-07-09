---
id: AYD-005
type: design
status: draft
updated: 2026-07-09
parents: [RF-04]
children: [SPEC-005@api, SPEC-005@web]
related: [GLO, AYD-001, AYD-004]
---

# AYD-005: Checkout & Order

> Fifth feature and the second half of the purchase flow (RF-04). It turns a **Cart**
> (AYD-004) into a confirmed **Order** with **one or more Products**, running a
> **simulated** payment. Reuses the Product model, the decimal-as-string rule and the
> error envelope from AYD-001, and consumes the `cart_id` from AYD-004. This AYD resolves
> the open question in AYD-001 about how Orders reference Products.

## Goal
Meets the purchase part of **RF-04**: a customer completes a purchase of the Products in
their Cart through the web UI, backed by the API, resulting in an Order; payment is
simulated (no real provider). Outcome: an end-to-end checkout flow (web → api → db) that
snapshots prices, decrements stock atomically (RN-04), and produces a persistent,
immutable Order.

## Affected parts
| Part | Role in this feature | Generated SPEC |
|-------|---------------------|-------------|
| api | Owns checkout: validate the Cart, re-check stock (RN-03), snapshot prices + decrement stock atomically (RN-04), simulate payment, create and expose the Order | SPEC-005@api |
| web | Checkout UI: minimal customer contact form, order review, "place order", and an order confirmation view | SPEC-005@web |

## Contract (source of truth)

Reuses AYD-001 conventions: JSON `snake_case`; decimals (`unit_price`, `subtotal`,
`total`) as **strings**; error envelope `{ "error": { "code", "message", "details"? } }`.

### Resources
```
Order
{
  id:         string (uuid, server-generated)
  status:     string enum { "confirmed" }        // single status for the MVP
  customer:   { name: string, email: string }    // minimal guest contact
  items:      [OrderItem]                         // >= 1
  total:      string decimal >= 0                 // sum of item.subtotal, immutable
  payment:    { method: "simulated", status: "approved" }
  created_at: string (ISO 8601)
}

OrderItem
{
  product_id: string (uuid)          // soft reference to Product (no hard FK cascade)
  sku:        string                 // snapshot at purchase time
  name:       string                 // snapshot at purchase time
  unit_price: string decimal >= 0    // snapshot of Product price at purchase time (RN-04)
  quantity:   integer >= 1
  subtotal:   string decimal >= 0    // unit_price * quantity
}
```

An Order is a **historical record**: `sku`, `name`, `unit_price` are snapshots, so later
editing or deleting the referenced Product never changes a past Order.

### Endpoints
```
POST /orders                          // checkout: create an Order from a Cart
req:  { cart_id: string, customer: { name: string, email: string } }
res 201: Order
errors: [ 404 cart_not_found,
          422 cart_empty,
          422 validation_error,       // missing/invalid customer fields
          409 insufficient_stock ]    // stock changed since add-to-cart

GET /orders/{order_id}
res 200: Order
errors: [ 404 order_not_found ]
```

Checkout (`POST /orders`) semantics — all steps in a **single DB transaction** (RN-04):
1. Load the Cart by `cart_id` (`404 cart_not_found`); it must be non-empty (`422 cart_empty`).
2. Validate `customer.name` and `customer.email` (`422 validation_error`, `details` per field).
3. Re-check every line's `quantity` against the Product's current `stock` (RN-03);
   any shortfall → `409 insufficient_stock` with `details` listing the offending lines,
   and **nothing is committed**.
4. Snapshot each line's `unit_price`/`sku`/`name` from the current Product (RN-04) and
   compute `subtotal` and Order `total`.
5. **Simulate payment** → always `{ method: "simulated", status: "approved" }` (no real
   provider; there is no decline path in the MVP).
6. Decrement each Product's `stock` by the purchased quantity and create the Order +
   OrderItems (`status: "confirmed"`).
7. Consume the Cart: after a successful checkout the Cart is emptied/closed, so a later
   `GET /carts/{cart_id}` reads as `404 cart_not_found` (AYD-004) — it cannot be checked
   out twice.

## Affected domain model
- **Order** (new): `id`, `status`, `customer` (name, email), `total`, `payment`,
  `created_at`.
- **Order Item** (new): belongs to an Order; snapshots Product fields; references
  **Product** by `product_id` as a **soft reference** (no cascading FK) so Product
  lifecycle is independent of order history.
- **Product** (AYD-001): `stock` is decremented on checkout (RN-04); no schema change.
  **Resolves AYD-001's open question** (Orders referencing Products): Products remain
  hard-deletable because OrderItems keep their own snapshot and hold no restricting FK.

## Flow
```mermaid
sequenceDiagram
    participant customer as Customer (browser)
    participant web
    participant api
    participant db
    customer->>web: fill contact + review cart + place order
    web->>api: POST /orders { cart_id, customer }
    api->>db: load Cart + items (404 if missing / 422 if empty)
    api->>db: re-check stock per line (RN-03)
    alt any line short
        api-->>web: 409 insufficient_stock (details), nothing committed
    else all available
        api->>api: snapshot unit_price/sku/name; compute totals (RN-04)
        api->>api: simulate payment -> approved
        api->>db: BEGIN; decrement stock; insert order+items; close cart; COMMIT
        api-->>web: 201 Order (confirmed)
        web-->>customer: order confirmation (id, items, total)
    end
```

## Out of scope / open questions
- **Out:** real payment provider, payment decline/retry — payment is always simulated and
  approved (RF-04); the `payment` block exists so a real provider can slot in later
  without a contract change.
- **Out:** order listing/history, cancellation, refunds, status transitions beyond
  `confirmed` — single terminal status for the MVP.
- **Out:** shipping address/cost, taxes, discounts/coupons (REQ MVP scope). `customer` is
  minimal contact only (name, email).
- **Out:** authentication — checkout is guest-only; the Order is not linked to a user
  account (REQ).
- **Open (stock model):** stock is decremented at checkout, not reserved at add-to-cart
  (see AYD-004); concurrent checkouts are serialized by the transaction + stock check, so
  the last one over the limit gets `409 insufficient_stock`. Row-locking strategy is an
  api-local concern (SPEC-005@api / a TDR@api), not a contract change.
