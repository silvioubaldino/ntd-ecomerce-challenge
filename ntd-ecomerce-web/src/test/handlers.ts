import { http, HttpResponse } from "msw";
import type { Cart, CartItem, Order, OrderItem, Product } from "../api/types";

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

export function makeOrderItem(overrides: Partial<OrderItem> = {}): OrderItem {
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

export function makeOrder(overrides: Partial<Order> = {}): Order {
  return {
    id: "order-1",
    status: "confirmed",
    customer: { name: "Ada Lovelace", email: "ada@example.com" },
    items: [makeOrderItem()],
    total: "19.90",
    payment: { method: "simulated", status: "approved" },
    created_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

// Fixture catalog for SPEC-006 filter/sort scenarios — the default GET /products
// handler filters and sorts this set by the received query params. Tests that need
// a specific shape (e.g. exact pagination totals) still override the handler with
// `server.use`, same as before.
const FIXTURE_PRODUCTS: Product[] = [
  makeProduct({
    id: "aaaaaaaa-0000-0000-0000-000000000001",
    name: "Alpha Jacket",
    sku: "APP-001",
    description: "A warm jacket",
    category: "Apparel",
    price: "45.00",
    created_at: "2026-01-01T00:00:00Z",
  }),
  makeProduct({
    id: "aaaaaaaa-0000-0000-0000-000000000002",
    name: "Beta Sneakers",
    sku: "SHO-001",
    description: "Running sneakers",
    category: "Shoes",
    price: "80.00",
    created_at: "2026-01-02T00:00:00Z",
  }),
  makeProduct({
    id: "aaaaaaaa-0000-0000-0000-000000000003",
    name: "Gamma Hammer",
    sku: "TOL-001",
    description: "A sturdy hammer",
    category: "Tools",
    price: "15.00",
    created_at: "2026-01-03T00:00:00Z",
  }),
  makeProduct({
    id: "aaaaaaaa-0000-0000-0000-000000000004",
    name: "Delta Boots",
    sku: "SHO-002",
    description: "Hiking boots",
    category: "Shoes",
    price: "60.00",
    created_at: "2026-01-04T00:00:00Z",
  }),
  makeProduct({
    id: "aaaaaaaa-0000-0000-0000-000000000005",
    name: "Widget",
    sku: "WID-001",
    description: "A widget",
    category: "Tools",
    price: "19.90",
    created_at: "2026-01-05T00:00:00Z",
  }),
];

function filterAndSortProducts(url: URL): Product[] {
  const q = url.searchParams.get("q")?.trim().toLowerCase();
  const category = url.searchParams.get("category")?.trim().toLowerCase();
  const priceMin = url.searchParams.get("price_min");
  const priceMax = url.searchParams.get("price_max");
  const sort = url.searchParams.get("sort");

  let items = [...FIXTURE_PRODUCTS];

  if (q) {
    items = items.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.sku.toLowerCase().includes(q) ||
        p.description.toLowerCase().includes(q),
    );
  }
  if (category) {
    items = items.filter((p) => p.category.toLowerCase() === category);
  }
  if (priceMin) {
    items = items.filter((p) => Number(p.price) >= Number(priceMin));
  }
  if (priceMax) {
    items = items.filter((p) => Number(p.price) <= Number(priceMax));
  }

  switch (sort) {
    case "price_asc":
      items.sort((a, b) => Number(a.price) - Number(b.price));
      break;
    case "price_desc":
      items.sort((a, b) => Number(b.price) - Number(a.price));
      break;
    case "name_asc":
      items.sort((a, b) => a.name.localeCompare(b.name));
      break;
    case "name_desc":
      items.sort((a, b) => b.name.localeCompare(a.name));
      break;
    case "newest":
      items.sort((a, b) => b.created_at.localeCompare(a.created_at));
      break;
    default:
      break;
  }

  return items;
}

export const handlers = [
  http.get("/api/products", ({ request }) => {
    const url = new URL(request.url);
    const page = Number(url.searchParams.get("page") ?? "1");
    const pageSize = Number(url.searchParams.get("page_size") ?? "20");

    const items = filterAndSortProducts(url);
    const total = items.length;
    const start = (page - 1) * pageSize;
    const pageItems = items.slice(start, start + pageSize);

    return HttpResponse.json({
      data: pageItems,
      pagination: { page, page_size: pageSize, total },
    });
  }),

  http.get("/api/products/categories", () => {
    const categories = Array.from(new Set(FIXTURE_PRODUCTS.map((p) => p.category))).sort();
    return HttpResponse.json({ data: categories });
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

  // Orders (AYD-005) — permissive defaults; tests override per scenario.
  http.post("/api/orders", async ({ request }) => {
    const body = (await request.json()) as {
      cart_id: string;
      customer: { name: string; email: string };
    };
    return HttpResponse.json(
      makeOrder({ customer: body.customer }),
      { status: 201 },
    );
  }),

  http.get("/api/orders/:orderId", ({ params }) => {
    if (params.orderId === "missing") {
      return HttpResponse.json(
        { error: { code: "order_not_found", message: "order not found" } },
        { status: 404 },
      );
    }
    return HttpResponse.json(makeOrder({ id: params.orderId as string }));
  }),
];
