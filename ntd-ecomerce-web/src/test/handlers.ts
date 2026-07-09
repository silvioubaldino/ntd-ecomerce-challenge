import { http, HttpResponse } from "msw";
import type { Cart, CartItem, Product } from "../api/types";

export function makeProduct(overrides: Partial<Product> = {}): Product {
  return {
    id: "11111111-1111-1111-1111-111111111111",
    name: "Widget",
    sku: "WID-001",
    description: "A widget",
    category: "Tools",
    price: "19.90",
    stock: 5,
    weight_kg: "0.50",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

export function makeCartItem(overrides: Partial<CartItem> = {}): CartItem {
  return {
    product_id: "11111111-1111-1111-1111-111111111111",
    sku: "WID-001",
    name: "Widget",
    unit_price: "19.90",
    quantity: 1,
    subtotal: "19.90",
    ...overrides,
  };
}

export function makeCart(overrides: Partial<Cart> = {}): Cart {
  return {
    id: "cart-1",
    items: [],
    total: "0.00",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

export const handlers = [
  http.get("/api/products", ({ request }) => {
    const url = new URL(request.url);
    const q = url.searchParams.get("q");
    const page = Number(url.searchParams.get("page") ?? "1");
    if (q) {
      return HttpResponse.json({
        data: [makeProduct({ name: `${q} match` })],
        pagination: { page, page_size: 20, total: 1 },
      });
    }
    return HttpResponse.json({
      data: [makeProduct()],
      pagination: { page, page_size: 20, total: 1 },
    });
  }),

  http.get("/api/products/:id", ({ params }) => {
    if (params.id === "missing") {
      return HttpResponse.json(
        { error: { code: "product_not_found", message: "product not found" } },
        { status: 404 },
      );
    }
    return HttpResponse.json(makeProduct({ id: params.id as string }));
  }),

  http.post("/api/products", async ({ request }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json(makeProduct(body as Partial<Product>), { status: 201 });
  }),

  http.put("/api/products/:id", async ({ request, params }) => {
    const body = (await request.json()) as Record<string, unknown>;
    return HttpResponse.json(
      makeProduct({ id: params.id as string, ...(body as Partial<Product>) }),
    );
  }),

  http.delete("/api/products/:id", () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.post("/api/products/import", () => {
    return HttpResponse.json({
      summary: { total: 1, imported: 1, rejected: 0 },
      rejected: [],
    });
  }),

  // Cart (AYD-004) — permissive defaults; tests override per scenario.
  http.post("/api/carts", () => HttpResponse.json(makeCart(), { status: 201 })),

  http.get("/api/carts/:cartId", ({ params }) =>
    HttpResponse.json(makeCart({ id: params.cartId as string })),
  ),

  http.post("/api/carts/:cartId/items", async ({ request, params }) => {
    const body = (await request.json()) as { product_id: string; quantity: number };
    const item = makeCartItem({
      product_id: body.product_id,
      quantity: body.quantity,
    });
    return HttpResponse.json(
      makeCart({ id: params.cartId as string, items: [item], total: item.subtotal }),
    );
  }),

  http.put("/api/carts/:cartId/items/:productId", async ({ request, params }) => {
    const body = (await request.json()) as { quantity: number };
    const item = makeCartItem({
      product_id: params.productId as string,
      quantity: body.quantity,
    });
    return HttpResponse.json(
      makeCart({ id: params.cartId as string, items: [item], total: item.subtotal }),
    );
  }),

  http.delete("/api/carts/:cartId/items/:productId", ({ params }) =>
    HttpResponse.json(makeCart({ id: params.cartId as string })),
  ),
];
