import { http, HttpResponse } from "msw";
import type { Product } from "../api/types";

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

export const handlers = [
  http.get("/api/products", () => {
    return HttpResponse.json({
      data: [makeProduct()],
      pagination: { page: 1, page_size: 20, total: 1 },
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
];
