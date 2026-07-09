import { http, HttpResponse } from "msw";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";
import { makeCart, makeCartItem, makeOrder } from "../../test/handlers";

const CART_ID_KEY = "ntd.cart_id";

function seedCart() {
  localStorage.setItem(CART_ID_KEY, "cart-1");
  server.use(
    http.get("/api/carts/:cartId", () =>
      HttpResponse.json(
        makeCart({
          id: "cart-1",
          items: [
            makeCartItem({
              product_id: "p1",
              name: "Blue sneakers",
              sku: "SNK-9",
              unit_price: "19.90",
              quantity: 2,
              subtotal: "39.80",
            }),
          ],
          total: "39.80",
        }),
      ),
    ),
  );
}

async function fillContact(name: string, email: string) {
  await userEvent.type(screen.getByLabelText("Full name"), name);
  await userEvent.type(screen.getByLabelText("Email"), email);
}

describe("CheckoutPage", () => {
  beforeEach(() => localStorage.clear());

  it("reviews the cart with line items and the total", async () => {
    // Arrange
    seedCart();

    // Act
    renderWithProviders(<App />, { route: "/checkout" });

    // Assert
    expect(await screen.findByText("Blue sneakers")).toBeInTheDocument();
    expect(screen.getByText("SNK-9")).toBeInTheDocument();
    expect(screen.getByText("19.90 × 2")).toBeInTheDocument();
    // 39.80 shows as both the line subtotal and the order total.
    expect(screen.getAllByText("39.80")).toHaveLength(2);
  });

  it("shows an empty state with a store link when there is no cart", async () => {
    // Arrange — no cart_id in localStorage

    // Act
    renderWithProviders(<App />, { route: "/checkout" });

    // Assert
    expect(await screen.findByText("Your cart is empty.")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Go to the store" })).toHaveAttribute(
      "href",
      "/store",
    );
    expect(screen.queryByLabelText("Full name")).not.toBeInTheDocument();
  });

  it("validates the contact fields before sending a request", async () => {
    // Arrange
    seedCart();
    let posted = false;
    server.use(
      http.post("/api/orders", () => {
        posted = true;
        return HttpResponse.json(makeOrder(), { status: 201 });
      }),
    );
    renderWithProviders(<App />, { route: "/checkout" });
    await screen.findByText("Blue sneakers");

    // Act — submit empty
    await userEvent.click(screen.getByRole("button", { name: "Place order" }));

    // Assert
    expect(await screen.findByText("Please enter your name.")).toBeInTheDocument();
    expect(screen.getByText("Please enter your email.")).toBeInTheDocument();

    // Act — invalid email
    await userEvent.type(screen.getByLabelText("Full name"), "Ada");
    await userEvent.type(screen.getByLabelText("Email"), "not-an-email");
    await userEvent.click(screen.getByRole("button", { name: "Place order" }));

    // Assert
    expect(
      await screen.findByText("Please enter a valid email address."),
    ).toBeInTheDocument();
    expect(posted).toBe(false);
  });

  it("places the order and navigates to the confirmation, clearing the cart", async () => {
    // Arrange
    seedCart();
    let body: unknown;
    server.use(
      http.post("/api/orders", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json(
          makeOrder({ id: "order-9", customer: { name: "Ada", email: "ada@example.com" } }),
          { status: 201 },
        );
      }),
    );
    renderWithProviders(<App />, { route: "/checkout" });
    await screen.findByText("Blue sneakers");

    // Act
    await fillContact("Ada", "ada@example.com");
    await userEvent.click(screen.getByRole("button", { name: "Place order" }));

    // Assert
    expect(await screen.findByText("Thank you for your order!")).toBeInTheDocument();
    expect(body).toEqual({
      cart_id: "cart-1",
      customer: { name: "Ada", email: "ada@example.com" },
    });
    expect(localStorage.getItem(CART_ID_KEY)).toBeNull();
  });

  it("shows insufficient_stock in a banner and does not navigate", async () => {
    // Arrange
    seedCart();
    server.use(
      http.post("/api/orders", () =>
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
    renderWithProviders(<App />, { route: "/checkout" });
    await screen.findByText("Blue sneakers");

    // Act
    await fillContact("Ada", "ada@example.com");
    await userEvent.click(screen.getByRole("button", { name: "Place order" }));

    // Assert
    expect(
      await screen.findByText("Only 3 in stock (you requested 4)."),
    ).toBeInTheDocument();
    expect(screen.queryByText("Thank you for your order!")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Place order" })).toBeInTheDocument();
  });

  it("handles cart_empty at submit with the empty state", async () => {
    // Arrange
    seedCart();
    server.use(
      http.post("/api/orders", () =>
        HttpResponse.json(
          { error: { code: "cart_empty", message: "cart is empty" } },
          { status: 422 },
        ),
      ),
    );
    renderWithProviders(<App />, { route: "/checkout" });
    await screen.findByText("Blue sneakers");

    // Act
    await fillContact("Ada", "ada@example.com");
    await userEvent.click(screen.getByRole("button", { name: "Place order" }));

    // Assert
    expect(await screen.findByText("Your cart is empty.")).toBeInTheDocument();
  });

  it("shows per-field messages from a 422 validation_error", async () => {
    // Arrange
    seedCart();
    server.use(
      http.post("/api/orders", () =>
        HttpResponse.json(
          {
            error: {
              code: "validation_error",
              message: "invalid customer",
              details: { email: "must be a valid email" },
            },
          },
          { status: 422 },
        ),
      ),
    );
    renderWithProviders(<App />, { route: "/checkout" });
    await screen.findByText("Blue sneakers");

    // Act
    await fillContact("Ada", "ada@example.com");
    await userEvent.click(screen.getByRole("button", { name: "Place order" }));

    // Assert
    expect(await screen.findByText("must be a valid email")).toBeInTheDocument();
    expect(screen.queryByText("Thank you for your order!")).not.toBeInTheDocument();
  });
});
