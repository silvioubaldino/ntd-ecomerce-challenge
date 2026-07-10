---
id: SPEC-005
type: spec
status: done
updated: 2026-07-09
parents: [AYD-005@context]
related: [GLO]
---

# SPEC-005: Checkout & Order (web) — what + how

> Implements the web side of AYD-005@context: guest checkout. A customer reviews their
> Cart, fills a minimal contact form (name, email), places the order, and lands on an
> order confirmation view. Consumes the AYD-005 contract exactly (`POST /orders`,
> `GET /orders/{order_id}`, decimals-as-strings, error envelope). Reuses `apiClient`, the
> `cart` feature (Cart query + guest `cart_id`), and the design components from
> SPEC-001/003/004; adds no new endpoint or contract.

## What (goal)
Deliver an end-to-end guest checkout in the web UI: a `/checkout` view that reviews the
current Cart (line items + total) and captures customer contact, a "Place order" action
that turns the Cart into a confirmed Order (`POST /orders`), and an `/orders/:orderId`
confirmation view rendering the immutable Order — with contract errors handled gracefully
and the local cart cleared after a successful checkout.

## Acceptance criteria
```gherkin
Scenario: review the cart at checkout
  Given the cart has one or more Cart Items
  When the customer opens /checkout
  Then each line shows sku, name, unit_price, quantity and subtotal (strings, as-is)
    and the cart total is shown

Scenario: cannot check out an empty or missing cart
  Given the customer has no cart or an empty cart
  When they open /checkout
  Then an empty state with a link to /store is shown and there is no place-order control

Scenario: validate the contact fields
  Given the customer is on /checkout with items in the cart
  When they submit without a name or with an invalid email
  Then per-field validation errors are shown and no request is sent

Scenario: place the order successfully and navigate to the confirmation
  Given a valid cart and valid contact fields
  When the customer places the order
  Then POST /orders is called with { cart_id, customer: { name, email } }
    and the app navigates to /orders/{order_id}

Scenario: insufficient stock is shown and nothing changes
  Given a line whose quantity now exceeds available stock
  When the customer places the order
  Then the 409 insufficient_stock is shown in a banner and no navigation happens

Scenario: cart_empty is handled
  Given the api reports the cart is empty at checkout
  When the customer places the order
  Then the empty-cart state is shown, not a raw error

Scenario: the confirmation page renders the order
  Given an existing order id
  When the customer opens /orders/{order_id}
  Then the order id, a confirmed/payment-approved indicator, the customer name and email,
    each Order Item (sku, name, unit_price, quantity, subtotal) and the total are shown

Scenario: order_not_found is handled
  Given an order id the api does not have
  When the customer opens /orders/{order_id}
  Then a friendly not-found state with a link to /store is shown, not a raw error

Scenario: the cart is cleared after a successful checkout
  Given a successful checkout
  When the order is confirmed
  Then the stored cart_id is cleared and the nav cart badge resets
```

## How (approach)
New `checkout` feature module. A new `api/orders.ts` wraps the AYD-005 endpoints via
`apiClient`. TanStack Query holds the confirmed Order (`["order", order_id]`).
`useCheckout` is a mutation that reads the guest `cart_id` from the existing `cart`
storage (`getStoredCartId`), calls `POST /orders`, and on success seeds the Order into the
cache, clears the stored `cart_id` (`clearStoredCartId`) and removes the cart query so the
nav badge resets — the cart is consumed server-side. A stale `cart_not_found` also resets
the stored cart. `useOrder` reads a single Order, enabled once an id is present. Checkout
reuses the existing `useCart` hook to review lines. Money values are strings from the api
and are rendered verbatim — no float math.

## Steps
1. **Types** (`src/api/types.ts`): add `Customer`, `Payment`, `OrderItem`, `Order`; extend
   `ApiErrorCode` with `order_not_found` and `cart_empty`.
2. **Client** (`src/api/orders.ts`): `createOrder(cartId, customer)` → `POST /orders`;
   `getOrder(orderId)` → `GET /orders/{orderId}`.
3. **Messages** (`src/features/checkout/checkoutMessages.ts`): map `insufficient_stock`
   (uses `details`), `validation_error`, `cart_empty`, `cart_not_found`; fall back to the
   cart message mapping / `error.message`.
4. **Hooks** (`src/features/checkout/hooks.ts`): `useCheckout` (mutation: create order,
   seed `["order", id]`, clear cart id, remove cart query; reset on `cart_not_found`);
   `useOrder(orderId)` (query `["order", orderId]`, enabled when present).
5. **Checkout page** (`src/features/checkout/CheckoutPage.tsx` + route `/checkout`): cart
   review, contact form with client-side validation and per-field errors, "Place order" →
   navigate to `/orders/{id}` on success; empty state when the cart is empty/missing;
   `ErrorBanner` for top-level errors; per-field messages from `error.details` on 422.
6. **Confirmation page** (`src/features/checkout/OrderConfirmationPage.tsx` + route
   `/orders/:orderId`): loading Skeleton; friendly not-found on `order_not_found`; on
   success the order id, confirmed/payment indicator, customer, Order Items and total, plus
   a "Continue shopping" link to `/store`.
7. **Cart entry point** (`src/features/cart/CartPage.tsx`): replace the stubbed checkout
   button with a real `ButtonLink to="/checkout"`.
8. **Routes** (`src/App.tsx`): add `/checkout` and `/orders/:orderId`.
9. **Tests**: MSW handlers for `/orders` (+ `makeOrder`); acceptance tests (below) + unit
   tests for the client and message mapping.
10. **Verify**: `npm run build`, `npm run lint`, `npm test`. Changelog line; mark this spec
    `status: done`.

## Affected files
- `ntd-ecomerce-web/src/api/types.ts` (add `Customer`, `Payment`, `OrderItem`, `Order`, error codes)
- `ntd-ecomerce-web/src/api/orders.ts` (new)
- `ntd-ecomerce-web/src/api/orders.test.ts` (new)
- `ntd-ecomerce-web/src/features/checkout/checkoutMessages.ts` (new)
- `ntd-ecomerce-web/src/features/checkout/checkoutMessages.test.ts` (new)
- `ntd-ecomerce-web/src/features/checkout/hooks.ts` (new)
- `ntd-ecomerce-web/src/features/checkout/CheckoutPage.tsx` (new)
- `ntd-ecomerce-web/src/features/checkout/CheckoutPage.test.tsx` (new)
- `ntd-ecomerce-web/src/features/checkout/OrderConfirmationPage.tsx` (new)
- `ntd-ecomerce-web/src/features/checkout/OrderConfirmationPage.test.tsx` (new)
- `ntd-ecomerce-web/src/features/cart/CartPage.tsx` (real checkout entry point)
- `ntd-ecomerce-web/src/components/ui/icons.tsx` (check icon)
- `ntd-ecomerce-web/src/App.tsx` (add `/checkout` and `/orders/:orderId` routes)
- `ntd-ecomerce-web/src/test/handlers.ts` (MSW `/orders` handlers + `makeOrder`)
- `ntd-ecomerce-web/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario (RTL + MSW, mocking only the api boundary):
  cart review at checkout; empty/missing cart empty state; per-field contact validation;
  place-order success → navigate to confirmation with the right body; 409 insufficient_stock
  shown, no navigation; 422 cart_empty handled; confirmation renders the order; 404
  order_not_found handled; cart cleared after success (nav badge resets).
- **Unit:** `api/orders.ts` builds the right method/path/body per endpoint; `checkoutMessages`
  maps each code (esp. insufficient_stock and cart_empty) and falls back to `error.message`.

## Checklist
- [x] `/checkout` reviews the cart (sku/name/unit_price/quantity/subtotal + total, strings as-is)
- [x] Contact form validates name + email client-side with per-field errors
- [x] Place order → `POST /orders { cart_id, customer }` → navigate to `/orders/{id}`
- [x] 409 insufficient_stock, 422 validation_error, 422 cart_empty, 404 cart_not_found handled
- [x] `/orders/:orderId` renders the confirmed Order; order_not_found shows a friendly state
- [x] Stored cart_id cleared after a successful checkout; nav badge resets
- [x] All acceptance scenarios covered by passing tests; lint clean; build green
- [x] Changelog line added; spec `done`
