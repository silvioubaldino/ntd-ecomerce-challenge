import { http, HttpResponse } from "msw";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";
import { makeCart, makeCartItem } from "../../test/handlers";

const CART_ID_KEY = "ntd.cart_id";

describe("CartPage", () => {
  beforeEach(() => localStorage.clear());

  it("lists each Cart Item with a running total", async () => {
    // Arrange
    localStorage.setItem(CART_ID_KEY, "cart-1");
    const item = makeCartItem({
      name: "Blue sneakers",
      sku: "SNK-9",
      unit_price: "19.90",
      quantity: 2,
      subtotal: "39.80",
    });
    server.use(
      http.get("/api/carts/:cartId", () =>
        HttpResponse.json(makeCart({ id: "cart-1", items: [item], total: "39.80" })),
      ),
    );

    // Act
    renderWithProviders(<App />, { route: "/cart" });

    // Assert
    expect(await screen.findByText("Blue sneakers")).toBeInTheDocument();
    expect(screen.getByText("SNK-9")).toBeInTheDocument();
    expect(screen.getByText("19.90 each")).toBeInTheDocument();
    expect(screen.getByLabelText("Quantity for Blue sneakers")).toHaveTextContent("2");
    // 39.80 shows as both the line subtotal and the running total.
    expect(screen.getAllByText("39.80")).toHaveLength(2);
  });

  it("increments a line quantity via PUT with the absolute value", async () => {
    // Arrange
    localStorage.setItem(CART_ID_KEY, "cart-1");
    let putBody: unknown;
    const item = makeCartItem({ product_id: "p1", quantity: 2 });
    server.use(
      http.get("/api/carts/:cartId", () =>
        HttpResponse.json(makeCart({ id: "cart-1", items: [item], total: "39.80" })),
      ),
      http.put("/api/carts/:cartId/items/:productId", async ({ request }) => {
        putBody = await request.json();
        return HttpResponse.json(
          makeCart({
            id: "cart-1",
            items: [makeCartItem({ product_id: "p1", quantity: 3, subtotal: "59.70" })],
            total: "59.70",
          }),
        );
      }),
    );

    renderWithProviders(<App />, { route: "/cart" });
    await screen.findByText("Widget");

    // Act
    await userEvent.click(screen.getByRole("button", { name: "Increase quantity" }));

    // Assert
    await waitFor(() => expect(putBody).toEqual({ quantity: 3 }));
    await waitFor(() => expect(screen.getAllByText("59.70")).toHaveLength(2));
    expect(screen.getByLabelText("Quantity for Widget")).toHaveTextContent("3");
  });

  it("disables the decrease control at quantity 1", async () => {
    // Arrange
    localStorage.setItem(CART_ID_KEY, "cart-1");
    server.use(
      http.get("/api/carts/:cartId", () =>
        HttpResponse.json(
          makeCart({ id: "cart-1", items: [makeCartItem({ quantity: 1 })], total: "19.90" }),
        ),
      ),
    );

    // Act
    renderWithProviders(<App />, { route: "/cart" });
    await screen.findByText("Widget");

    // Assert
    expect(screen.getByRole("button", { name: "Decrease quantity" })).toBeDisabled();
  });

  it("removes a line via DELETE and then shows the empty state", async () => {
    // Arrange
    localStorage.setItem(CART_ID_KEY, "cart-1");
    let deletePath: string | undefined;
    server.use(
      http.get("/api/carts/:cartId", () =>
        HttpResponse.json(
          makeCart({
            id: "cart-1",
            items: [makeCartItem({ product_id: "p1", name: "Widget" })],
            total: "19.90",
          }),
        ),
      ),
      http.delete("/api/carts/:cartId/items/:productId", ({ request }) => {
        deletePath = new URL(request.url).pathname;
        return HttpResponse.json(makeCart({ id: "cart-1", items: [], total: "0.00" }));
      }),
    );

    renderWithProviders(<App />, { route: "/cart" });
    await screen.findByText("Widget");

    // Act
    await userEvent.click(screen.getByRole("button", { name: "Remove Widget" }));

    // Assert
    await waitFor(() => expect(deletePath).toBe("/api/carts/cart-1/items/p1"));
    expect(await screen.findByText("Your cart is empty.")).toBeInTheDocument();
  });

  it("shows an empty state with a store link when there is no cart", async () => {
    // Arrange — no cart_id in localStorage

    // Act
    renderWithProviders(<App />, { route: "/cart" });

    // Assert
    expect(await screen.findByText("Your cart is empty.")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Go to the store" })).toHaveAttribute(
      "href",
      "/store",
    );
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("shows insufficient_stock inline and keeps the quantity unchanged", async () => {
    // Arrange
    localStorage.setItem(CART_ID_KEY, "cart-1");
    server.use(
      http.get("/api/carts/:cartId", () =>
        HttpResponse.json(
          makeCart({
            id: "cart-1",
            items: [makeCartItem({ product_id: "p1", quantity: 3 })],
            total: "59.70",
          }),
        ),
      ),
      http.put("/api/carts/:cartId/items/:productId", () =>
        HttpResponse.json(
          {
            error: {
              code: "insufficient_stock",
              message: "insufficient stock",
              details: { product_id: "p1", requested: "4", available: "3" },
            },
          },
          { status: 409 },
        ),
      ),
    );

    renderWithProviders(<App />, { route: "/cart" });
    await screen.findByText("Widget");

    // Act
    await userEvent.click(screen.getByRole("button", { name: "Increase quantity" }));

    // Assert
    expect(
      await screen.findByText("Only 3 in stock (you requested 4)."),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("Quantity for Widget")).toHaveTextContent("3");
  });

  it("shows a checkout entry point when the cart has items", async () => {
    // Arrange
    localStorage.setItem(CART_ID_KEY, "cart-1");
    server.use(
      http.get("/api/carts/:cartId", () =>
        HttpResponse.json(
          makeCart({ id: "cart-1", items: [makeCartItem()], total: "19.90" }),
        ),
      ),
    );

    // Act
    renderWithProviders(<App />, { route: "/cart" });
    await screen.findByText("Widget");

    // Assert
    expect(
      screen.getByRole("button", { name: /proceed to checkout/i }),
    ).toBeInTheDocument();
  });
});
