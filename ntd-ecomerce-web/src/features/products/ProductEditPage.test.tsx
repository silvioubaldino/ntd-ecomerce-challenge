import { http, HttpResponse } from "msw";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";
import { makeProduct } from "../../test/handlers";

describe("ProductEditPage", () => {
  it("sends the full ProductInput on PUT and returns to the list on success", async () => {
    const product = makeProduct({ id: "abc-123" });
    let receivedBody: Record<string, unknown> | undefined;
    server.use(
      http.get("/api/products/:id", () => HttpResponse.json(product)),
      http.put("/api/products/:id", async ({ request }) => {
        receivedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json({ ...product, ...receivedBody });
      }),
    );

    renderWithProviders(<App />, { route: "/products/abc-123/edit" });

    const nameInput = await screen.findByLabelText("Name");
    expect(nameInput).toHaveValue(product.name);

    await userEvent.clear(nameInput);
    await userEvent.type(nameInput, "Renamed widget");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByRole("heading", { name: "Products" })).toBeInTheDocument();
    expect(receivedBody).toMatchObject({
      name: "Renamed widget",
      sku: product.sku,
      description: product.description,
      category: product.category,
      price: product.price,
      stock: product.stock,
      weight_kg: product.weight_kg,
    });
  });

  it("shows a not-found message when the api returns 404", async () => {
    server.use(
      http.get("/api/products/:id", () =>
        HttpResponse.json(
          { error: { code: "product_not_found", message: "product not found" } },
          { status: 404 },
        ),
      ),
    );

    renderWithProviders(<App />, { route: "/products/missing/edit" });

    expect(await screen.findByText("Product not found.")).toBeInTheDocument();
    expect(screen.queryByLabelText("Name")).not.toBeInTheDocument();
  });
});
