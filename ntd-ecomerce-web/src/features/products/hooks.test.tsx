import type { ReactNode } from "react";
import { http, HttpResponse } from "msw";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { server } from "../../test/server";
import { useProductSearch, type ProductSearchFilters } from "./hooks";

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

describe("useProductSearch", () => {
  it("issues a separate request per distinct filter combination (varies the query key)", async () => {
    const seenQueries: string[] = [];
    server.use(
      http.get("/api/products", ({ request }) => {
        seenQueries.push(new URL(request.url).search);
        return HttpResponse.json({ data: [], pagination: { limit: 20, next_cursor: null } });
      }),
    );

    const wrapper = createWrapper();
    const { result, rerender } = renderHook(
      (filters: ProductSearchFilters) => useProductSearch(filters),
      { wrapper, initialProps: { q: "" } as ProductSearchFilters },
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(seenQueries).toHaveLength(1);

    rerender({ q: "" });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(seenQueries).toHaveLength(1);

    rerender({ q: "", category: "Apparel" });
    await waitFor(() => expect(seenQueries).toHaveLength(2));

    rerender({ q: "", category: "Apparel", priceMin: "10" });
    await waitFor(() => expect(seenQueries).toHaveLength(3));

    rerender({ q: "", category: "Apparel", priceMin: "10", priceMax: "50" });
    await waitFor(() => expect(seenQueries).toHaveLength(4));

    rerender({
      q: "",
      category: "Apparel",
      priceMin: "10",
      priceMax: "50",
      sort: "price_asc",
    });
    await waitFor(() => expect(seenQueries).toHaveLength(5));

    rerender({
      q: "",
      category: "Apparel",
      priceMin: "10",
      priceMax: "50",
      sort: "price_asc",
      cursor: "cursor:20",
    });
    await waitFor(() => expect(seenQueries).toHaveLength(6));
  });

  it("does not fire a request while disabled (e.g. an invalid price range)", async () => {
    let requestCount = 0;
    server.use(
      http.get("/api/products", () => {
        requestCount += 1;
        return HttpResponse.json({ data: [], pagination: { limit: 20, next_cursor: null } });
      }),
    );

    const wrapper = createWrapper();
    renderHook(() => useProductSearch({ q: "", enabled: false }), { wrapper });

    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(requestCount).toBe(0);
  });
});
