import { http, HttpResponse } from "msw";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";
import { makeProduct } from "../../test/handlers";

describe("ProductListPage", () => {
  it("lists products with pagination and page controls that change the page", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const page = Number(new URL(request.url).searchParams.get("page"));
        return HttpResponse.json({
          data: [makeProduct({ name: `Page ${page} product` })],
          pagination: { page, page_size: 20, total: 45 },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/products" });

    expect(await screen.findByText("Page 1 product")).toBeInTheDocument();
    expect(screen.getByText("19.90")).toBeInTheDocument();
    expect(screen.getByText("Page 1 of 3 (45 total)")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Next" }));

    expect(await screen.findByText("Page 2 product")).toBeInTheDocument();
    expect(screen.getByText("Page 2 of 3 (45 total)")).toBeInTheDocument();
  });

  it("shows an empty-state message instead of a table when the catalog is empty", async () => {
    server.use(
      http.get("/api/products", () =>
        HttpResponse.json({ data: [], pagination: { page: 1, page_size: 20, total: 0 } }),
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
          pagination: { page: 1, page_size: 20, total: products.length },
        }),
      ),
      http.delete("/api/products/:id", ({ params }) => {
        products = products.filter((product) => product.id !== params.id);
        return new HttpResponse(null, { status: 204 });
      }),
    );
    vi.spyOn(window, "confirm").mockReturnValue(true);

    renderWithProviders(<App />, { route: "/products" });

    expect(await screen.findByText("Widget")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "Delete" }));

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
          pagination: { page: 1, page_size: 20, total: 1 },
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
