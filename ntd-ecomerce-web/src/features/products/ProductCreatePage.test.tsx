import { http, HttpResponse } from "msw";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";
import { makeProduct } from "../../test/handlers";

async function fillValidForm() {
  await userEvent.type(screen.getByLabelText("Name"), "Gadget");
  await userEvent.type(screen.getByLabelText("SKU"), "GDG-001");
  await userEvent.type(screen.getByLabelText("Category"), "Tools");
  await userEvent.clear(screen.getByLabelText("Price"));
  await userEvent.type(screen.getByLabelText("Price"), "10.00");
  await userEvent.clear(screen.getByLabelText("Weight (kg)"));
  await userEvent.type(screen.getByLabelText("Weight (kg)"), "0.10");
}

describe("ProductCreatePage", () => {
  it("sends ProductInput with decimals as strings and returns to the list on success", async () => {
    let receivedBody: Record<string, unknown> | undefined;
    server.use(
      http.post("/api/products", async ({ request }) => {
        receivedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(makeProduct(receivedBody), { status: 201 });
      }),
    );

    renderWithProviders(<App />, { route: "/products/new" });
    await fillValidForm();
    await userEvent.click(screen.getByRole("button", { name: "Create" }));

    expect(await screen.findByRole("heading", { name: "Products" })).toBeInTheDocument();
    expect(receivedBody).toMatchObject({
      name: "Gadget",
      sku: "GDG-001",
      category: "Tools",
      price: "10.00",
      weight_kg: "0.10",
    });
    expect(typeof receivedBody?.price).toBe("string");
  });

  it("blocks submit and shows field errors on invalid input without calling the api", async () => {
    let called = false;
    server.use(
      http.post("/api/products", () => {
        called = true;
        return HttpResponse.json(makeProduct(), { status: 201 });
      }),
    );

    renderWithProviders(<App />, { route: "/products/new" });
    await userEvent.clear(screen.getByLabelText("Price"));
    await userEvent.type(screen.getByLabelText("Price"), "-5");
    await userEvent.click(screen.getByRole("button", { name: "Create" }));

    expect(await screen.findAllByText(/required|non-negative/i)).not.toHaveLength(0);
    expect(called).toBe(false);
  });

  it("maps 422 validation_error details onto field errors", async () => {
    server.use(
      http.post("/api/products", () =>
        HttpResponse.json(
          {
            error: {
              code: "validation_error",
              message: "validation failed",
              details: { sku: "too_long" },
            },
          },
          { status: 422 },
        ),
      ),
    );

    renderWithProviders(<App />, { route: "/products/new" });
    await fillValidForm();
    await userEvent.click(screen.getByRole("button", { name: "Create" }));

    expect(await screen.findByText("too long")).toBeInTheDocument();
  });

  it("maps 409 sku_already_exists to the sku field", async () => {
    server.use(
      http.post("/api/products", () =>
        HttpResponse.json(
          { error: { code: "sku_already_exists", message: "sku already exists" } },
          { status: 409 },
        ),
      ),
    );

    renderWithProviders(<App />, { route: "/products/new" });
    await fillValidForm();
    await userEvent.click(screen.getByRole("button", { name: "Create" }));

    expect(await screen.findByText("SKU already exists")).toBeInTheDocument();
  });
});
