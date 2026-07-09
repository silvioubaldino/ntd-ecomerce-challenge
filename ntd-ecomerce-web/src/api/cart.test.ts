import { http, HttpResponse } from "msw";
import { server } from "../test/server";
import {
  addCartItem,
  createCart,
  getCart,
  removeCartItem,
  updateCartItem,
} from "./cart";

describe("cart api client", () => {
  it("POSTs an empty body to /carts to create a cart", async () => {
    // Arrange
    let method: string | undefined;
    server.use(
      http.post("/api/carts", ({ request }) => {
        method = request.method;
        return HttpResponse.json({ id: "c1" }, { status: 201 });
      }),
    );

    // Act
    const cart = await createCart();

    // Assert
    expect(method).toBe("POST");
    expect(cart.id).toBe("c1");
  });

  it("GETs /carts/{id}", async () => {
    // Arrange
    server.use(http.get("/api/carts/:id", ({ params }) => HttpResponse.json({ id: params.id })));

    // Act
    const cart = await getCart("abc");

    // Assert
    expect(cart.id).toBe("abc");
  });

  it("POSTs product_id and quantity when adding an item", async () => {
    // Arrange
    let body: unknown;
    let path: string | undefined;
    server.use(
      http.post("/api/carts/:id/items", async ({ request }) => {
        body = await request.json();
        path = new URL(request.url).pathname;
        return HttpResponse.json({ id: "c1" });
      }),
    );

    // Act
    await addCartItem("c1", "p1", 2);

    // Assert
    expect(path).toBe("/api/carts/c1/items");
    expect(body).toEqual({ product_id: "p1", quantity: 2 });
  });

  it("PUTs the absolute quantity for a line", async () => {
    // Arrange
    let body: unknown;
    let method: string | undefined;
    server.use(
      http.put("/api/carts/:id/items/:productId", async ({ request }) => {
        method = request.method;
        body = await request.json();
        return HttpResponse.json({ id: "c1" });
      }),
    );

    // Act
    await updateCartItem("c1", "p1", 5);

    // Assert
    expect(method).toBe("PUT");
    expect(body).toEqual({ quantity: 5 });
  });

  it("DELETEs a line by product id", async () => {
    // Arrange
    let method: string | undefined;
    let path: string | undefined;
    server.use(
      http.delete("/api/carts/:id/items/:productId", ({ request }) => {
        method = request.method;
        path = new URL(request.url).pathname;
        return HttpResponse.json({ id: "c1" });
      }),
    );

    // Act
    await removeCartItem("c1", "p1");

    // Assert
    expect(method).toBe("DELETE");
    expect(path).toBe("/api/carts/c1/items/p1");
  });
});
