import { http, HttpResponse } from "msw";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter, useNavigate } from "react-router-dom";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";
import { makeProduct } from "../../test/handlers";

function BackButton() {
  const navigate = useNavigate();
  return (
    <button type="button" onClick={() => navigate(-1)}>
      Simulate browser back
    </button>
  );
}

function renderWithBackButton(route: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter
        initialEntries={[route]}
        future={{ v7_startTransition: true, v7_relativeSplatPath: true }}
      >
        <BackButton />
        <App />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("StoreSearchPage", () => {
  it("shows the full catalog when the search term is blank", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const q = new URL(request.url).searchParams.get("q");
        expect(q).toBeNull();
        return HttpResponse.json({
          data: [makeProduct({ name: "Widget" })],
          pagination: { limit: 20, next_cursor: null },
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
          pagination: { limit: 20, next_cursor: null },
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
            pagination: { limit: 20, next_cursor: null },
          });
        }
        return HttpResponse.json({
          data: [makeProduct()],
          pagination: { limit: 20, next_cursor: null },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    await userEvent.type(screen.getByLabelText("Search products"), "zzz-no-match");

    expect(await screen.findByText("No products match your search.")).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("reflects the search term in the URL and pre-fills the input on reload", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const q = new URL(request.url).searchParams.get("q");
        return HttpResponse.json({
          data: q ? [makeProduct({ name: `${q} match` })] : [makeProduct()],
          pagination: { limit: 20, next_cursor: null },
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
          pagination: { limit: 20, next_cursor: null },
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

describe("StoreSearchPage — keyset pagination (SPEC-008)", () => {
  it("loads the first page without a cursor and disables Prev", async () => {
    let requestedCursor: string | null | undefined;
    server.use(
      http.get("/api/products", ({ request }) => {
        requestedCursor = new URL(request.url).searchParams.get("cursor");
        return HttpResponse.json({
          data: [makeProduct({ name: "Widget" })],
          pagination: { limit: 20, next_cursor: "idx:1" },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store" });

    expect(await screen.findByText("Widget")).toBeInTheDocument();
    expect(requestedCursor).toBeNull();
    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();
  });

  it("requests the next page with the returned cursor when Next is clicked", async () => {
    let requestedCursor: string | null | undefined;
    server.use(
      http.get("/api/products", ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor");
        requestedCursor = cursor;
        if (!cursor) {
          return HttpResponse.json({
            data: [makeProduct({ name: "Page 1 item" })],
            pagination: { limit: 20, next_cursor: "idx:1" },
          });
        }
        return HttpResponse.json({
          data: [makeProduct({ name: "Page 2 item" })],
          pagination: { limit: 20, next_cursor: null },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Page 1 item");

    await userEvent.click(screen.getByRole("button", { name: "Next" }));

    expect(await screen.findByText("Page 2 item")).toBeInTheDocument();
    expect(requestedCursor).toBe("idx:1");
  });

  it("disables Next and fires no request when next_cursor is null", async () => {
    let requestCount = 0;
    server.use(
      http.get("/api/products", () => {
        requestCount += 1;
        return HttpResponse.json({
          data: [makeProduct()],
          pagination: { limit: 20, next_cursor: null },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");
    const countAfterLoad = requestCount;

    const nextButton = screen.getByRole("button", { name: "Next" });
    expect(nextButton).toBeDisabled();

    await userEvent.click(nextButton);
    expect(requestCount).toBe(countAfterLoad);
  });

  it("walks Next then Prev through the client-held cursor stack", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor");
        if (!cursor) {
          return HttpResponse.json({
            data: [makeProduct({ name: "Item A" })],
            pagination: { limit: 20, next_cursor: "idx:1" },
          });
        }
        if (cursor === "idx:1") {
          return HttpResponse.json({
            data: [makeProduct({ name: "Item B" })],
            pagination: { limit: 20, next_cursor: "idx:2" },
          });
        }
        return HttpResponse.json({
          data: [makeProduct({ name: "Item C" })],
          pagination: { limit: 20, next_cursor: null },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Item A");

    await userEvent.click(screen.getByRole("button", { name: "Next" }));
    expect(await screen.findByText("Item B")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Next" }));
    expect(await screen.findByText("Item C")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Previous" }));
    expect(await screen.findByText("Item B")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Previous" }));
    expect(await screen.findByText("Item A")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();
  });

  it("resets the cursor and stack when the search term changes on a later page", async () => {
    let lastCursor: string | null = null;
    server.use(
      http.get("/api/products", ({ request }) => {
        const url = new URL(request.url);
        lastCursor = url.searchParams.get("cursor");
        const q = url.searchParams.get("q");
        if (!lastCursor) {
          return HttpResponse.json({
            data: [makeProduct({ name: q ? `${q} item 1` : "item 1" })],
            pagination: { limit: 20, next_cursor: "idx:1" },
          });
        }
        return HttpResponse.json({
          data: [makeProduct({ name: q ? `${q} item 2` : "item 2" })],
          pagination: { limit: 20, next_cursor: null },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("item 1");

    await userEvent.click(screen.getByRole("button", { name: "Next" }));
    await screen.findByText("item 2");

    await userEvent.type(screen.getByLabelText("Search products"), "blue");

    expect(await screen.findByText("blue item 1")).toBeInTheDocument();
    await waitFor(() => expect(lastCursor).toBeNull());
    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();
  });

  it("never renders page numbers or a total count", async () => {
    renderWithProviders(<App />, { route: "/store" });
    await screen.findByText("Widget");

    expect(screen.queryByText(/Page \d+ of \d+/)).not.toBeInTheDocument();
    expect(screen.queryByText(/total\)/)).not.toBeInTheDocument();
  });

  it("returns to the previous cursor when the browser Back button is used", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor");
        if (!cursor) {
          return HttpResponse.json({
            data: [makeProduct({ name: "Item A" })],
            pagination: { limit: 20, next_cursor: "idx:1" },
          });
        }
        return HttpResponse.json({
          data: [makeProduct({ name: "Item B" })],
          pagination: { limit: 20, next_cursor: null },
        });
      }),
    );

    renderWithBackButton("/store");

    await screen.findByText("Item A");
    await userEvent.click(screen.getByRole("button", { name: "Next" }));
    await screen.findByText("Item B");

    await userEvent.click(screen.getByRole("button", { name: "Simulate browser back" }));

    expect(await screen.findByText("Item A")).toBeInTheDocument();
  });

  it("loads the requested page from a ?cursor= deep link and falls back to the first page on Prev", async () => {
    let requestedCursor: string | null | undefined;
    server.use(
      http.get("/api/products", ({ request }) => {
        requestedCursor = new URL(request.url).searchParams.get("cursor");
        if (requestedCursor === "abc123") {
          return HttpResponse.json({
            data: [makeProduct({ name: "Deep link item" })],
            pagination: { limit: 20, next_cursor: null },
          });
        }
        return HttpResponse.json({
          data: [makeProduct({ name: "First page item" })],
          pagination: { limit: 20, next_cursor: "abc123" },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store?cursor=abc123" });

    expect(await screen.findByText("Deep link item")).toBeInTheDocument();
    expect(requestedCursor).toBe("abc123");
    expect(screen.getByRole("button", { name: "Previous" })).toBeEnabled();

    await userEvent.click(screen.getByRole("button", { name: "Previous" }));

    expect(await screen.findByText("First page item")).toBeInTheDocument();
    expect(requestedCursor).toBeNull();
  });

  it("recovers to the first page when the api rejects the cursor as invalid", async () => {
    server.use(
      http.get("/api/products", ({ request }) => {
        const cursor = new URL(request.url).searchParams.get("cursor");
        if (cursor === "stale") {
          return HttpResponse.json(
            {
              error: {
                code: "validation_error",
                message: "invalid cursor",
                details: { cursor: "invalid_cursor" },
              },
            },
            { status: 422 },
          );
        }
        return HttpResponse.json({
          data: [makeProduct({ name: "First page item" })],
          pagination: { limit: 20, next_cursor: null },
        });
      }),
    );

    renderWithProviders(<App />, { route: "/store?cursor=stale" });

    expect(await screen.findByText("First page item")).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
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

  it("keeps every active param across pagination and resets the cursor on a filter change", async () => {
    let lastCursor: string | null = null;
    server.use(
      http.get("/api/products", ({ request }) => {
        const url = new URL(request.url);
        const q = url.searchParams.get("q");
        const category = url.searchParams.get("category");
        const priceMin = url.searchParams.get("price_min");
        lastCursor = url.searchParams.get("cursor");
        expect(q).toBe("boot");
        expect(priceMin).toBe("10");
        if (!lastCursor) {
          return HttpResponse.json({
            data: [makeProduct({ name: "match page 1", category: category ?? "Shoes" })],
            pagination: { limit: 20, next_cursor: "idx:1" },
          });
        }
        return HttpResponse.json({
          data: [makeProduct({ name: "match page 2", category: category ?? "Shoes" })],
          pagination: { limit: 20, next_cursor: null },
        });
      }),
    );

    renderWithProviders(<App />, {
      route: "/store?q=boot&category=Shoes&price_min=10",
    });

    expect(await screen.findByText("match page 1")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Next" }));
    expect(await screen.findByText("match page 2")).toBeInTheDocument();
    expect(lastCursor).toBe("idx:1");

    await userEvent.selectOptions(screen.getByLabelText("Sort by"), "price_asc");
    await waitFor(() => expect(lastCursor).toBeNull());
  });

  it("pre-fills every control and carries all params on the first request when opened via URL", async () => {
    let requestSearch: string | undefined;
    server.use(
      http.get("/api/products", ({ request }) => {
        requestSearch = new URL(request.url).search;
        return HttpResponse.json({
          data: [makeProduct({ name: "Widget" })],
          pagination: { limit: 20, next_cursor: null },
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
          pagination: { limit: 20, next_cursor: null },
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
