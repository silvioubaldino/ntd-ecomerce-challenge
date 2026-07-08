import { http, HttpResponse } from "msw";
import { server } from "../test/server";
import { apiClient } from "./client";
import { ApiError } from "./types";

describe("apiClient error envelope parsing", () => {
  it("parses a 422 validation_error into a typed ApiError with details", async () => {
    server.use(
      http.post("/api/products", () =>
        HttpResponse.json(
          {
            error: {
              code: "validation_error",
              message: "validation failed",
              details: { price: "must_be_non_negative_decimal" },
            },
          },
          { status: 422 },
        ),
      ),
    );

    const error = (await apiClient.post("/products", {}).catch((e) => e)) as ApiError;
    expect(error).toBeInstanceOf(ApiError);
    expect(error.status).toBe(422);
    expect(error.code).toBe("validation_error");
    expect(error.details).toEqual({ price: "must_be_non_negative_decimal" });
  });

  it("parses a 409 sku_already_exists into a typed ApiError", async () => {
    server.use(
      http.post("/api/products", () =>
        HttpResponse.json(
          { error: { code: "sku_already_exists", message: "sku already exists" } },
          { status: 409 },
        ),
      ),
    );

    const error = (await apiClient.post("/products", {}).catch((e) => e)) as ApiError;
    expect(error.status).toBe(409);
    expect(error.code).toBe("sku_already_exists");
  });

  it("parses a 404 product_not_found into a typed ApiError", async () => {
    server.use(
      http.get("/api/products/:id", () =>
        HttpResponse.json(
          { error: { code: "product_not_found", message: "product not found" } },
          { status: 404 },
        ),
      ),
    );

    const error = (await apiClient.get("/products/missing").catch((e) => e)) as ApiError;
    expect(error.status).toBe(404);
    expect(error.code).toBe("product_not_found");
  });
});

describe("apiClient.postForm", () => {
  it("sends a multipart body without a JSON Content-Type header", async () => {
    let receivedContentType: string | null = null;
    server.use(
      http.post("/api/products/import", ({ request }) => {
        receivedContentType = request.headers.get("content-type");
        return HttpResponse.json({ summary: { total: 0, imported: 0, rejected: 0 }, rejected: [] });
      }),
    );

    const formData = new FormData();
    formData.append("file", new File(["name,sku"], "products.csv", { type: "text/csv" }));
    await apiClient.postForm("/products/import", formData);

    expect(receivedContentType).toMatch(/^multipart\/form-data/);
  });

  it("parses the error envelope for a 422 invalid_header response", async () => {
    server.use(
      http.post("/api/products/import", () =>
        HttpResponse.json(
          { error: { code: "invalid_header", message: "header mismatch" } },
          { status: 422 },
        ),
      ),
    );

    const formData = new FormData();
    formData.append("file", new File(["bad"], "products.csv"));
    const error = (await apiClient
      .postForm("/products/import", formData)
      .catch((e) => e)) as ApiError;

    expect(error).toBeInstanceOf(ApiError);
    expect(error.status).toBe(422);
    expect(error.code).toBe("invalid_header");
  });
});
