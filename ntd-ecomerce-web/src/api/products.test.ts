import { http, HttpResponse } from "msw";
import { server } from "../test/server";
import { listCategories, listProducts } from "./products";

describe("listProducts", () => {
  it("sends only page and page_size when no filters are given", async () => {
    let query: string | undefined;
    server.use(
      http.get("/api/products", ({ request }) => {
        query = new URL(request.url).search;
        return HttpResponse.json({ data: [], pagination: { page: 1, page_size: 20, total: 0 } });
      }),
    );

    await listProducts({ page: 1 });

    const params = new URLSearchParams(query);
    expect(params.get("page")).toBe("1");
    expect(params.get("page_size")).toBe("20");
    expect(params.has("q")).toBe(false);
    expect(params.has("category")).toBe(false);
    expect(params.has("price_min")).toBe(false);
    expect(params.has("price_max")).toBe(false);
    expect(params.has("sort")).toBe(false);
  });

  it("omits blank/whitespace-only filter values", async () => {
    let query: string | undefined;
    server.use(
      http.get("/api/products", ({ request }) => {
        query = new URL(request.url).search;
        return HttpResponse.json({ data: [], pagination: { page: 1, page_size: 20, total: 0 } });
      }),
    );

    await listProducts({ page: 1, q: "  ", category: "  ", priceMin: "", priceMax: "  " });

    const params = new URLSearchParams(query);
    expect(params.has("q")).toBe(false);
    expect(params.has("category")).toBe(false);
    expect(params.has("price_min")).toBe(false);
    expect(params.has("price_max")).toBe(false);
  });

  it("sends trimmed q, category, price_min, price_max and sort verbatim", async () => {
    let query: string | undefined;
    server.use(
      http.get("/api/products", ({ request }) => {
        query = new URL(request.url).search;
        return HttpResponse.json({ data: [], pagination: { page: 1, page_size: 20, total: 0 } });
      }),
    );

    await listProducts({
      page: 2,
      pageSize: 10,
      q: "  shirt  ",
      category: " Apparel ",
      priceMin: "10",
      priceMax: "50",
      sort: "price_asc",
    });

    const params = new URLSearchParams(query);
    expect(params.get("page")).toBe("2");
    expect(params.get("page_size")).toBe("10");
    expect(params.get("q")).toBe("shirt");
    expect(params.get("category")).toBe("Apparel");
    expect(params.get("price_min")).toBe("10");
    expect(params.get("price_max")).toBe("50");
    expect(params.get("sort")).toBe("price_asc");
  });
});

describe("listCategories", () => {
  it("GETs /products/categories", async () => {
    let path: string | undefined;
    server.use(
      http.get("/api/products/categories", ({ request }) => {
        path = new URL(request.url).pathname;
        return HttpResponse.json({ data: ["Apparel", "Shoes"] });
      }),
    );

    const result = await listCategories();

    expect(path).toBe("/api/products/categories");
    expect(result.data).toEqual(["Apparel", "Shoes"]);
  });
});
