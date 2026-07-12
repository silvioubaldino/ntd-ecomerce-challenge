import { http, HttpResponse } from "msw";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";
import { makeProduct } from "../../test/handlers";

describe("ProductListPage", () => {
  it("lists products with Prev/Next controls driven by cursors", async () => {
    let requestedCursor: string | null | undefined;
    server.use(
      http.get("/api/products", ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor");
        requestedCursor = cursor;
        if (!cursor) {
          return HttpResponse.json({
            data: [makeProduct({ name: "Page 1 product" })],
            pagination: { limit: 20, next_cursor: "idx:1" },
          });
        }
        return HttpResponse.json({
          data: [makeProduct({ name: "Page 2 product" })],
          pagination: { limit: 20, next_cursor: null },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/products" });

    expect(await screen.findByText("Page 1 product")).toBeInTheDocument();
    expect(screen.getByText("19.90")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();

    await userEvent.click(screen.getByRole("button", { name: "Next" }));

    expect(await screen.findByText("Page 2 product")).toBeInTheDocument();
    expect(requestedCursor).toBe("idx:1");
    expect(screen.getByRole("button", { name: "Next" })).toBeDisabled();

    await userEvent.click(screen.getByRole("button", { name: "Previous" }));

    expect(await screen.findByText("Page 1 product")).toBeInTheDocument();
    expect(requestedCursor).toBeNull();
  });

  it("never renders page numbers or a total count", async () => {
    renderWithProviders(<App />, { route: "/products" });
    await screen.findByText("Widget");

    expect(screen.queryByText(/Page \d+ of \d+/)).not.toBeInTheDocument();
    expect(screen.queryByText(/total\)/)).not.toBeInTheDocument();
  });

  it("shows an empty-state message instead of a table when the catalog is empty", async () => {
    server.use(
      http.get("/api/products", () =>
        HttpResponse.json({ data: [], pagination: { limit: 20, next_cursor: null } }),
      ),
    );

    renderWithProviders(<App />, { route: "/products" });

    expect(await screen.findByText("No products yet.")).toBeInTheDocument();
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
  });

  it("removes a product from the list after a confirmed delete", async () => {
    let products = [makeProduct()];
    server.use(
      http.get("/api/products", () =>
        HttpResponse.json({
          data: products,
          pagination: { limit: 20, next_cursor: null },
        }),
      ),
      http.delete("/api/products/:id", ({ params }) => {
        products = products.filter((product) => product.id !== params.id);
        return new HttpResponse(null, { status: 204 });
      }),
    );
    renderWithProviders(<App />, { route: "/products" });

    expect(await screen.findByText("Widget")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));

    const dialog = await screen.findByRole("dialog");
    await userEvent.click(within(dialog).getByRole("button", { name: "Delete product" }));

    await waitFor(() => expect(screen.queryByText("Widget")).not.toBeInTheDocument());
    expect(await screen.findByText("No products yet.")).toBeInTheDocument();
  });

  it("shows a recoverable error state when the request fails", async () => {
    let failed = true;
    server.use(
      http.get("/api/products", () => {
        if (failed) return new HttpResponse(null, { status: 500 });
        return HttpResponse.json({
          data: [makeProduct()],
          pagination: { limit: 20, next_cursor: null },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/products" });

    const banner = await screen.findByRole("alert");
    expect(within(banner).getByText(/could not load the catalog/i)).toBeInTheDocument();

    failed = false;
    await userEvent.click(within(banner).getByRole("button", { name: "Retry" }));

    expect(await screen.findByText("Widget")).toBeInTheDocument();
  });
});
