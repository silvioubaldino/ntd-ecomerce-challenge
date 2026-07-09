import { http, HttpResponse } from "msw";
import { server } from "../test/server";
import { createOrder, getOrder } from "./orders";

describe("orders api client", () => {
  it("POSTs cart_id and customer to /orders", async () => {
    // Arrange
    let method: string | undefined;
    let path: string | undefined;
    let body: unknown;
    server.use(
      http.post("/api/orders", async ({ request }) => {
        method = request.method;
        path = new URL(request.url).pathname;
        body = await request.json();
        return HttpResponse.json({ id: "o1" }, { status: 201 });
      }),
    );

    // Act
    const order = await createOrder("cart-1", { name: "Ada", email: "ada@example.com" });

    // Assert
    expect(method).toBe("POST");
    expect(path).toBe("/api/orders");
    expect(body).toEqual({
      cart_id: "cart-1",
      customer: { name: "Ada", email: "ada@example.com" },
    });
    expect(order.id).toBe("o1");
  });

  it("GETs /orders/{id}", async () => {
    // Arrange
    let path: string | undefined;
    server.use(
      http.get("/api/orders/:orderId", ({ request, params }) => {
        path = new URL(request.url).pathname;
        return HttpResponse.json({ id: params.orderId });
      }),
    );

    // Act
    const order = await getOrder("o1");

    // Assert
    expect(path).toBe("/api/orders/o1");
    expect(order.id).toBe("o1");
  });
});
