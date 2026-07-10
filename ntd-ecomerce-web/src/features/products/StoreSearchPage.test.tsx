import { http, HttpResponse } from "msw";
import { screen, waitFor } from "@testing-library/react";
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

describe("StoreSearchPage — filters and sorting (SPEC-006)", () => {
  it("feeds the category dropdown from GET /products/categories", async () => {
    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    const select = screen.getByLabelText("Category") as HTMLSelectElement;
    const options = Array.from(select.options).map((option) => option.text);
    expect(options).toEqual(["All categories", "Apparel", "Shoes", "Tools"]);
  });

  it("filters by category and reflects it in the URL", async () => {
    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    await userEvent.selectOptions(screen.getByLabelText("Category"), "Apparel");

    expect(await screen.findByText("Alpha Jacket")).toBeInTheDocument();
    expect(screen.queryByText("Widget")).not.toBeInTheDocument();
    expect(screen.queryByText("Beta Sneakers")).not.toBeInTheDocument();

    const { unmount } = renderWithProviders(<App />, {
      route: "/store?category=Apparel",
    });
    expect(screen.getByLabelText("Category")).toHaveValue("Apparel");
    expect(await screen.findByText("Alpha Jacket")).toBeInTheDocument();
    unmount();
  });

  it("filters by price range and reflects both bounds in the request", async () => {
    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    await userEvent.type(screen.getByLabelText("Minimum price"), "50");
    await userEvent.type(screen.getByLabelText("Maximum price"), "100");

    await waitFor(() => expect(screen.queryByText("Widget")).not.toBeInTheDocument());
    expect(screen.getByText("Beta Sneakers")).toBeInTheDocument();
    expect(screen.getByText("Delta Boots")).toBeInTheDocument();
    expect(screen.queryByText("Gamma Hammer")).not.toBeInTheDocument();
    expect(screen.queryByText("Alpha Jacket")).not.toBeInTheDocument();
  });

  it("reorders results when the customer sorts by price ascending", async () => {
    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    await userEvent.selectOptions(screen.getByLabelText("Sort by"), "price_asc");

    await waitFor(() => {
      const names = screen.getAllByRole("listitem").map((item) => item.textContent ?? "");
      const gamma = names.findIndex((text) => text.includes("Gamma Hammer"));
      const widget = names.findIndex((text) => text.includes("Widget"));
      const beta = names.findIndex((text) => text.includes("Beta Sneakers"));
      expect(gamma).toBeGreaterThanOrEqual(0);
      expect(gamma).toBeLessThan(widget);
      expect(widget).toBeLessThan(beta);
    });
  });

  it("keeps every active param across pagination and resets page on a filter change", async () => {
    let lastPage = 0;
    server.use(
      http.get("/api/products", ({ request }) => {
        const url = new URL(request.url);
        const q = url.searchParams.get("q");
        const category = url.searchParams.get("category");
        const priceMin = url.searchParams.get("price_min");
        const page = Number(url.searchParams.get("page") ?? "1");
        lastPage = page;
        expect(q).toBe("boot");
        expect(priceMin).toBe("10");
        return HttpResponse.json({
          data: [makeProduct({ name: `match page ${page}`, category: category ?? "Shoes" })],
          pagination: { page, page_size: 20, total: 40 },
        });
      }),
    );

    renderWithProviders(<App />, {
      route: "/store?q=boot&category=Shoes&price_min=10",
    });

    expect(await screen.findByText("match page 1")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Next" }));
    expect(await screen.findByText("match page 2")).toBeInTheDocument();
    expect(lastPage).toBe(2);

    await userEvent.selectOptions(screen.getByLabelText("Sort by"), "price_asc");
    await waitFor(() => expect(lastPage).toBe(1));
  });

  it("pre-fills every control and carries all params on the first request when opened via URL", async () => {
    let requestSearch: string | undefined;
    server.use(
      http.get("/api/products", ({ request }) => {
        requestSearch = new URL(request.url).search;
        return HttpResponse.json({
          data: [makeProduct({ name: "Widget" })],
          pagination: { page: 1, page_size: 20, total: 1 },
        });
      }),
    );

    renderWithProviders(<App />, {
      route: "/store?q=shirt&category=Apparel&price_min=10&sort=price_desc",
    });

    await screen.findByText("Widget");

    expect(screen.getByLabelText("Search products")).toHaveValue("shirt");
    expect(screen.getByLabelText("Category")).toHaveValue("Apparel");
    expect(screen.getByLabelText("Minimum price")).toHaveValue("10");
    expect(screen.getByLabelText("Sort by")).toHaveValue("price_desc");

    const params = new URLSearchParams(requestSearch);
    expect(params.get("q")).toBe("shirt");
    expect(params.get("category")).toBe("Apparel");
    expect(params.get("price_min")).toBe("10");
    expect(params.get("sort")).toBe("price_desc");
  });

  it("clears category/price/sort on Clear filters and loads the unfiltered list", async () => {
    renderWithProviders(<App />, {
      route: "/store?category=Apparel&price_min=10&sort=price_asc",
    });
    await screen.findByText("Alpha Jacket");

    await userEvent.click(screen.getByRole("button", { name: "Clear filters" }));

    expect(await screen.findByText("Widget")).toBeInTheDocument();
    expect(screen.getByLabelText("Category")).toHaveValue("");
    expect(screen.getByLabelText("Minimum price")).toHaveValue("");
    expect(screen.getByLabelText("Sort by")).toHaveValue("");
  });

  it("blocks min > max client-side with an inline message and fires no request", async () => {
    let requestCount = 0;
    server.use(
      http.get("/api/products", () => {
        requestCount += 1;
        return HttpResponse.json({
          data: [makeProduct()],
          pagination: { page: 1, page_size: 20, total: 1 },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");
    const countBeforeGuard = requestCount;

    await userEvent.type(screen.getByLabelText("Minimum price"), "50");
    await userEvent.type(screen.getByLabelText("Maximum price"), "10");

    expect(
      await screen.findByText("Minimum price cannot exceed maximum price."),
    ).toBeInTheDocument();

    await new Promise((resolve) => setTimeout(resolve, 400));
    expect(requestCount).toBe(countBeforeGuard);
  });

  it("shows a filtered empty state with a clear action when no products match", async () => {
    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    await userEvent.selectOptions(screen.getByLabelText("Category"), "Apparel");
    await userEvent.type(screen.getByLabelText("Minimum price"), "1000");

    expect(await screen.findByText("No products match your filters.")).toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "Clear filters" }).length).toBeGreaterThan(0);
  });

  it("degrades gracefully when the categories lookup fails — the list still renders", async () => {
    server.use(
      http.get("/api/products/categories", () =>
        HttpResponse.json(
          { error: { code: "internal_error", message: "boom" } },
          { status: 500 },
        ),
      ),
    );

    renderWithProviders(<App />, { route: "/store" });

    expect(await screen.findByText("Widget")).toBeInTheDocument();
    const select = screen.getByLabelText("Category") as HTMLSelectElement;
    const options = Array.from(select.options).map((option) => option.text);
    expect(options).toEqual(["All categories"]);
  });
});
