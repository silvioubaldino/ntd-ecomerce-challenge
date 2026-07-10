---
id: SPEC-004
type: spec
status: done
updated: 2026-07-09
parents: [AYD-004@context]
related: [GLO]
---

# SPEC-004: Cart (web) — what + how

> Implements the web side of AYD-004@context: the storefront Cart. A guest customer adds
> Products to a Cart from the Store, views the Cart with per-line quantity edit / remove and
> a running total, and reaches a checkout entry point (checkout itself is AYD-005, out of
> scope). Consumes the AYD-004 contract exactly (`/carts` endpoints, guest `cart_id` in
> `localStorage`, decimals-as-strings, error envelope). Reuses `apiClient` and the design
> components from SPEC-001/003; adds no new endpoint or contract.

## What (goal)
Deliver an end-to-end guest Cart in the web UI: "Add to cart" from the Store, a `/cart` view
listing each Cart Item (SKU, name, unit_price, quantity, subtotal) with quantity edit and
line removal, a running total, and a checkout entry point — with contract errors handled
gracefully.

## Acceptance criteria
```gherkin
Scenario: lazily create the cart on first add
  Given the customer has no cart_id in localStorage
  When they click "Add to cart" on a Product in the Store
  Then the web creates a cart (POST /carts), persists cart_id in localStorage,
    and adds the Product (POST /carts/{id}/items with quantity 1)

Scenario: reuse the existing cart on later adds
  Given the customer already has a cart_id in localStorage
  When they add another Product
  Then no new cart is created and the item is added to the existing cart

Scenario: adding an already-present Product increments its quantity
  Given a Product is already in the cart
  When the customer clicks "Add to cart" again for that Product
  Then POST /carts/{id}/items is called and the cart reflects the incremented quantity

Scenario: view the cart with lines and running total
  Given the cart has one or more Cart Items
  When the customer opens /cart
  Then each line shows sku, name, unit_price, quantity and subtotal (strings, rendered as-is)
    and the cart total from the response is shown

Scenario: edit a line quantity
  Given a Cart Item with quantity 2
  When the customer sets the quantity to 3
  Then PUT /carts/{id}/items/{product_id} is called with quantity 3 and the cart reflects it

Scenario: remove a line
  Given a Cart Item in the cart
  When the customer removes the line
  Then DELETE /carts/{id}/items/{product_id} is called and the line disappears

Scenario: empty cart shows an empty state
  Given the cart has no items (or no cart exists yet)
  When the customer opens /cart
  Then a "Your cart is empty" empty state is shown with a link to the Store, not an error

Scenario: insufficient stock is reported and the cart is unchanged
  Given a Cart Item at the Product's max stock
  When the customer requests more than is available
  Then the 409 insufficient_stock is shown as "Only N in stock (you requested M)."
    and the displayed quantity is unchanged

Scenario: a stale cart resets transparently on add
  Given the local cart_id points to a cart the api no longer has (404 cart_not_found)
  When the customer clicks "Add to cart"
  Then the web creates a fresh cart, persists the new cart_id, and adds the item

Scenario: checkout entry point is present but deferred
  Given the cart has items
  When the customer views the cart
  Then a "Proceed to checkout" control is visible (checkout is AYD-005, not implemented here)
```

## How (approach)
New `cart` feature module. Guest identity lives in `localStorage` under a single key,
exposed through a tiny external store (`useSyncExternalStore`) so the nav badge and pages
stay in sync. A new `api/cart.ts` wraps the AYD-004 endpoints via `apiClient`; TanStack Query
holds the Cart (`["cart", cart_id]`), and mutations write the returned Cart straight into the
cache (the api returns the full Cart on every write, so no refetch). `useAddToCart` creates
the cart lazily on first use and transparently recreates it on `404 cart_not_found`. Money
values are strings from the api and are rendered verbatim — no float math.

## Steps
1. **Types** (`src/api/types.ts`): add `Cart` and `CartItem`; extend `ApiErrorCode` union
   with `cart_not_found`, `item_not_found`, `insufficient_stock`.
2. **Client** (`src/api/cart.ts`): `createCart()`, `getCart(cartId)`,
   `addCartItem(cartId, productId, quantity)`, `updateCartItem(cartId, productId, quantity)`,
   `removeCartItem(cartId, productId)` — all returning `Cart`.
3. **Guest identity** (`src/features/cart/cartStorage.ts`): `getStoredCartId`,
   `setStoredCartId`, `clearStoredCartId`, `subscribeCartId` (localStorage key `ntd.cart_id`
   + in-memory listeners).
4. **Hooks** (`src/features/cart/hooks.ts`): `useCartId` (external store),
   `useCart` (query, disabled when no cart_id; clears storage on `cart_not_found`),
   `useAddToCart` (lazy create + recreate-on-404 + retry), `useUpdateCartItem`,
   `useRemoveCartItem`. Mutations `setQueryData` with the returned Cart.
5. **Error messages** (`src/features/cart/cartMessages.ts`): map `insufficient_stock`
   (uses `details.requested`/`available`), `validation_error`, `product_not_found` to
   friendly copy; fall back to `error.message`.
6. **Add to cart** (`src/features/products/StoreSearchPage.tsx`): per-Product "Add to cart"
   button (disabled + "Out of stock" when `stock === 0`); on click, `useAddToCart`; surface
   errors inline per row.
7. **Cart view** (`src/features/cart/CartPage.tsx` + route `/cart` in `App.tsx`): list lines,
   quantity stepper (+/- calling PUT; minus disabled at 1, remove via trash → DELETE),
   subtotals, running total, "Proceed to checkout" button (stub, disabled/no-op with a note),
   empty state with a link to `/store`, loading Skeleton, error banner.
8. **Nav** (`src/components/Layout.tsx`): add a "Cart" link with a live item-count badge from
   the cached Cart.
9. **Tests**: MSW handlers for all `/carts` endpoints; acceptance tests (below) + unit tests
   for the client, storage, and message mapping.
10. **Verify**: `npm run build`, `npm run lint`, `npm test`. Changelog line; mark this spec
    `status: done`.

## Affected files
- `ntd-ecomerce-web/src/api/types.ts` (add `Cart`, `CartItem`, error codes)
- `ntd-ecomerce-web/src/api/cart.ts` (new)
- `ntd-ecomerce-web/src/api/cart.test.ts` (new)
- `ntd-ecomerce-web/src/features/cart/cartStorage.ts` (new)
- `ntd-ecomerce-web/src/features/cart/cartMessages.ts` (new)
- `ntd-ecomerce-web/src/features/cart/cartMessages.test.ts` (new)
- `ntd-ecomerce-web/src/features/cart/hooks.ts` (new)
- `ntd-ecomerce-web/src/features/cart/CartPage.tsx` (new)
- `ntd-ecomerce-web/src/features/cart/CartPage.test.tsx` (new)
- `ntd-ecomerce-web/src/features/products/StoreSearchPage.tsx` (add "Add to cart")
- `ntd-ecomerce-web/src/features/products/AddToCart.test.tsx` (new)
- `ntd-ecomerce-web/src/components/Layout.tsx` (Cart nav link + count badge)
- `ntd-ecomerce-web/src/components/ui/icons.tsx` (cart icon)
- `ntd-ecomerce-web/src/App.tsx` (add `/cart` route)
- `ntd-ecomerce-web/src/test/handlers.ts` (MSW `/carts` handlers + `makeCart`)
- `ntd-ecomerce-web/docs/changelog.md`

## Tests
- **Acceptance:** one test per Gherkin scenario (RTL + MSW, mocking only the api boundary):
  lazy create on first add; reuse existing cart; increment on repeat add; cart view renders
  lines + total; edit quantity (PUT); remove line (DELETE); empty-cart empty state;
  409 insufficient_stock shown and quantity unchanged; stale-cart 404 recreate; checkout
  entry point present.
- **Unit:** `api/cart.ts` builds the right method/path/body per endpoint; `cartStorage`
  get/set/clear + subscribe notifies; `cartMessages` maps each code (esp. insufficient_stock
  with requested/available) and falls back to `error.message`.

## Checklist
- [x] `cart_id` created lazily, persisted in localStorage, sent on every cart call
- [x] Add to cart from the Store; repeat add increments
- [x] Cart view lists sku/name/unit_price/quantity/subtotal with running total (strings as-is)
- [x] Quantity edit (PUT) and line remove (DELETE) work and update the view
- [x] 409 insufficient_stock, 422 validation_error, 404 cart_not_found handled gracefully
- [x] Checkout entry point present (order creation deferred to AYD-005)
- [x] All acceptance scenarios covered by passing tests; lint clean; build green
- [x] Changelog line added; spec `done`
