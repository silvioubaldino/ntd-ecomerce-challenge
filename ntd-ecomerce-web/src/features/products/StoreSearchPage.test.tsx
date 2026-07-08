import { http, HttpResponse } from "msw";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";
import { makeProduct } from "../../test/handlers";

describe("StoreSearchPage", () => {
  it("shows the full catalog when the search term is blank", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const q = new URL(request.url).searchParams.get("q");
        expect(q).toBeNull();
        return HttpResponse.json({
          data: [makeProduct({ name: "Widget" })],
          pagination: { page: 1, page_size: 20, total: 1 },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store" });

    expect(await screen.findByText("Widget")).toBeInTheDocument();
  });

  it("shows matching products when the customer types a search term", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const q = new URL(request.url).searchParams.get("q");
        return HttpResponse.json({
          data: q ? [makeProduct({ name: `${q} sneakers` })] : [makeProduct()],
          pagination: { page: 1, page_size: 20, total: q ? 1 : 1 },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    await userEvent.type(screen.getByLabelText("Search products"), "blue");

    expect(await screen.findByText("blue sneakers")).toBeInTheDocument();
  });

  it("shows an empty state, not an error, when there are zero matches", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const q = new URL(request.url).searchParams.get("q");
        if (q) {
          return HttpResponse.json({
            data: [],
            pagination: { page: 1, page_size: 20, total: 0 },
          });
        }
        return HttpResponse.json({
          data: [makeProduct()],
          pagination: { page: 1, page_size: 20, total: 1 },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    await userEvent.type(screen.getByLabelText("Search products"), "zzz-no-match");

    expect(await screen.findByText("No products match your search.")).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("walks pagination of the filtered set, keeping q and using the filtered total", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const url = new URL(request.url);
        const q = url.searchParams.get("q");
        const page = Number(url.searchParams.get("page") ?? "1");
        expect(q).toBe("blue");
        return HttpResponse.json({
          data: [makeProduct({ name: `blue page ${page}` })],
          pagination: { page, page_size: 20, total: 45 },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store?q=blue" });

    expect(await screen.findByText("blue page 1")).toBeInTheDocument();
    expect(screen.getByText("Page 1 of 3 (45 total)")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Next" }));

    expect(await screen.findByText("blue page 2")).toBeInTheDocument();
    expect(screen.getByText("Page 2 of 3 (45 total)")).toBeInTheDocument();
  });

  it("reflects the search term in the URL and pre-fills the input on reload", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const q = new URL(request.url).searchParams.get("q");
        return HttpResponse.json({
          data: q ? [makeProduct({ name: `${q} match` })] : [makeProduct()],
          pagination: { page: 1, page_size: 20, total: 1 },
        });
      }),
    );

    const { unmount } = renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    await userEvent.type(screen.getByLabelText("Search products"), "green");
    await screen.findByText("green match");
    unmount();

    // Simulate reloading the page directly at /store?q=green (the URL updated
    // by the debounced search above is now the entry point).
    renderWithProviders(<App />, { route: "/store?q=green" });
    expect(screen.getByLabelText("Search products")).toHaveValue("green");
    expect(await screen.findByText("green match")).toBeInTheDocument();
  });

  it("restores the full list and removes q from the URL when the search is cleared", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const q = new URL(request.url).searchParams.get("q");
        return HttpResponse.json({
          data: q ? [makeProduct({ name: `${q} match` })] : [makeProduct({ name: "Widget" })],
          pagination: { page: 1, page_size: 20, total: 1 },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store?q=green" });
    expect(await screen.findByText("green match")).toBeInTheDocument();

    const input = screen.getByLabelText("Search products");
    await userEvent.clear(input);

    expect(await screen.findByText("Widget")).toBeInTheDocument();
    expect(input).toHaveValue("");
  });
});
