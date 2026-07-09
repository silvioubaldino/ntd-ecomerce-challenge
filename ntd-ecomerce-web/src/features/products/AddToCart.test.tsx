import { http, HttpResponse } from "msw";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";
import { makeCart, makeCartItem, makeProduct } from "../../test/handlers";

const CART_ID_KEY = "ntd.cart_id";

function stubProducts(product = makeProduct()) {
  server.use(
    http.get("/api/products", () =>
      HttpResponse.json({
        data: [product],
        pagination: { page: 1, page_size: 20, total: 1 },
      }),
    ),
  );
}

describe("Add to cart from the Store", () => {
  beforeEach(() => localStorage.clear());

  it("creates the cart lazily and adds the product on the first add", async () => {
    // Arrange
    stubProducts();
    let cartsCreated = 0;
    let itemPath: string | undefined;
    let itemBody: unknown;
    server.use(
      http.post("/api/carts", () => {
        cartsCreated += 1;
        return HttpResponse.json(makeCart({ id: "new-cart" }), { status: 201 });
      }),
      http.post("/api/carts/:cartId/items", async ({ request }) => {
        itemPath = new URL(request.url).pathname;
        itemBody = await request.json();
        return HttpResponse.json(
          makeCart({ id: "new-cart", items: [makeCartItem()], total: "19.90" }),
        );
      }),
      http.get("/api/carts/:cartId", () =>
        HttpResponse.json(makeCart({ id: "new-cart", items: [makeCartItem()], total: "19.90" })),
      ),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    // Act
    await userEvent.click(screen.getByRole("button", { name: "Add Widget to cart" }));

    // Assert
    await waitFor(() => expect(localStorage.getItem(CART_ID_KEY)).toBe("new-cart"));
    expect(cartsCreated).toBe(1);
    expect(itemPath).toBe("/api/carts/new-cart/items");
    expect(itemBody).toEqual({ product_id: makeProduct().id, quantity: 1 });
  });

  it("reuses the existing cart on later adds (no new POST /carts)", async () => {
    // Arrange
    localStorage.setItem(CART_ID_KEY, "existing");
    stubProducts();
    let cartsCreated = 0;
    let itemPath: string | undefined;
    server.use(
      http.post("/api/carts", () => {
        cartsCreated += 1;
        return HttpResponse.json(makeCart({ id: "should-not-happen" }), { status: 201 });
      }),
      http.post("/api/carts/:cartId/items", async ({ request }) => {
        itemPath = new URL(request.url).pathname;
        return HttpResponse.json(
          makeCart({ id: "existing", items: [makeCartItem()], total: "19.90" }),
        );
      }),
      http.get("/api/carts/:cartId", () => HttpResponse.json(makeCart({ id: "existing" }))),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    // Act
    await userEvent.click(screen.getByRole("button", { name: "Add Widget to cart" }));

    // Assert
    await waitFor(() => expect(itemPath).toBe("/api/carts/existing/items"));
    expect(cartsCreated).toBe(0);
  });

  it("recreates a stale cart on 404 and retries the add", async () => {
    // Arrange
    localStorage.setItem(CART_ID_KEY, "stale");
    stubProducts();
    server.use(
      http.post("/api/carts", () =>
        HttpResponse.json(makeCart({ id: "fresh" }), { status: 201 }),
      ),
      http.post("/api/carts/:cartId/items", ({ params }) => {
        if (params.cartId === "stale") {
          return HttpResponse.json(
            { error: { code: "cart_not_found", message: "gone" } },
            { status: 404 },
          );
        }
        return HttpResponse.json(
          makeCart({ id: "fresh", items: [makeCartItem()], total: "19.90" }),
        );
      }),
      http.get("/api/carts/:cartId", ({ params }) =>
        HttpResponse.json(makeCart({ id: params.cartId as string })),
      ),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    // Act
    await userEvent.click(screen.getByRole("button", { name: "Add Widget to cart" }));

    // Assert
    await waitFor(() => expect(localStorage.getItem(CART_ID_KEY)).toBe("fresh"));
  });

  it("shows an insufficient_stock message when the add is rejected", async () => {
    // Arrange
    localStorage.setItem(CART_ID_KEY, "existing");
    stubProducts();
    server.use(
      http.post("/api/carts/:cartId/items", () =>
        HttpResponse.json(
          {
            error: {
              code: "insufficient_stock",
              message: "insufficient stock",
              details: { product_id: "p1", requested: "1", available: "0" },
            },
          },
          { status: 409 },
        ),
      ),
      http.get("/api/carts/:cartId", () => HttpResponse.json(makeCart({ id: "existing" }))),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    // Act
    await userEvent.click(screen.getByRole("button", { name: "Add Widget to cart" }));

    // Assert
    expect(
      await screen.findByText("Only 0 in stock (you requested 1)."),
    ).toBeInTheDocument();
  });

  it("disables adding an out-of-stock product", async () => {
    // Arrange
    stubProducts(makeProduct({ stock: 0 }));

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    // Assert
    expect(screen.getByRole("button", { name: "Add Widget to cart" })).toBeDisabled();
    expect(screen.getByText("Out of stock")).toBeInTheDocument();
  });
});
