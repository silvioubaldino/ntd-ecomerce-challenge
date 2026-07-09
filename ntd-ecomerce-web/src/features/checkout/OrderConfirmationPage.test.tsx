import { http, HttpResponse } from "msw";
import { screen } from "@testing-library/react";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";
import { makeOrder, makeOrderItem } from "../../test/handlers";

describe("OrderConfirmationPage", () => {
  it("renders the confirmed order with items and total", async () => {
    // Arrange
    server.use(
      http.get("/api/orders/:orderId", () =>
        HttpResponse.json(
          makeOrder({
            id: "order-42",
            customer: { name: "Ada Lovelace", email: "ada@example.com" },
            items: [
              makeOrderItem({
                sku: "SNK-9",
                name: "Blue sneakers",
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

    // Act
    renderWithProviders(<App />, { route: "/orders/order-42" });

    // Assert
    expect(await screen.findByText("Thank you for your order!")).toBeInTheDocument();
    expect(screen.getByText("order-42")).toBeInTheDocument();
    expect(screen.getByText("Ada Lovelace")).toBeInTheDocument();
    expect(screen.getByText("ada@example.com")).toBeInTheDocument();
    expect(screen.getByText("Blue sneakers")).toBeInTheDocument();
    expect(screen.getByText("SNK-9")).toBeInTheDocument();
    expect(screen.getByText("19.90 × 2")).toBeInTheDocument();
    // 39.80 shows as both the line subtotal and the order total.
    expect(screen.getAllByText("39.80")).toHaveLength(2);
    expect(screen.getByRole("link", { name: "Continue shopping" })).toHaveAttribute(
      "href",
      "/store",
    );
  });

  it("shows a friendly not-found state on order_not_found", async () => {
    // Arrange — the default handler returns 404 for the "missing" id

    // Act
    renderWithProviders(<App />, { route: "/orders/missing" });

    // Assert
    expect(await screen.findByText("Order not found.")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Continue shopping" })).toHaveAttribute(
      "href",
      "/store",
    );
    expect(screen.queryByText("Thank you for your order!")).not.toBeInTheDocument();
  });
});
