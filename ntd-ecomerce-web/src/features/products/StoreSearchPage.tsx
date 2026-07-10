import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { ErrorBanner } from "../../components/ErrorBanner";
import { PageHeader } from "../../components/PageHeader";
import { Badge } from "../../components/ui/Badge";
import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { EmptyState } from "../../components/ui/EmptyState";
import { Skeleton } from "../../components/ui/Skeleton";
import {
  BoxIcon,
  CartIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
} from "../../components/ui/icons";
import type { Product } from "../../api/types";
import { cartErrorMessage } from "../cart/cartMessages";
import { useAddToCart } from "../cart/hooks";
import { useProductSearch } from "./hooks";

const SEARCH_DEBOUNCE_MS = 300;

export function StoreSearchPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const q = searchParams.get("q") ?? "";
  const page = Number(searchParams.get("page") ?? "1") || 1;

  const [inputValue, setInputValue] = useState(q);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    setInputValue(q);
  }, [q]);

  useEffect(() => {
    const trimmed = inputValue.trim();
    if (trimmed === q.trim()) return;

    debounceRef.current = setTimeout(() => {
      applySearch(inputValue);
    }, SEARCH_DEBOUNCE_MS);

    return () => clearTimeout(debounceRef.current);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [inputValue]);

  function applySearch(value: string) {
    clearTimeout(debounceRef.current);
    const trimmed = value.trim();
    const next = new URLSearchParams(searchParams);
    if (trimmed) {
      next.set("q", trimmed);
    } else {
      next.delete("q");
    }
    next.delete("page");
    setSearchParams(next);
  }

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(searchParams);
    next.set("page", String(nextPage));
    setSearchParams(next);
  }

  const searchQuery = useProductSearch(q, page);

  if (searchQuery.isError) {
    return (
      <div className="flex flex-col gap-6">
        <PageHeader title="Store" description="Search the product catalog." />
        <SearchInput
          value={inputValue}
          onChange={setInputValue}
          onSubmit={() => applySearch(inputValue)}
        />
        <ErrorBanner
          message="Could not load products."
          onRetry={() => searchQuery.refetch()}
        />
      </div>
    );
  }

  const result = searchQuery.data;
  const totalPages = result
    ? Math.max(1, Math.ceil(result.pagination.total / result.pagination.page_size))
    : 1;

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Store" description="Search the product catalog." />

      <SearchInput
        value={inputValue}
        onChange={setInputValue}
        onSubmit={() => applySearch(inputValue)}
      />

      {!result ? (
        <Card>
          <ResultsSkeleton />
        </Card>
      ) : result.data.length === 0 ? (
        <Card>
          <EmptyState
            icon={<BoxIcon className="h-7 w-7" />}
            title="No products match your search."
            description={
              q ? `No results for "${q}". Try a different term.` : "The catalog is empty."
            }
          />
        </Card>
      ) : (
        <Card>
          <ul className="divide-y divide-slate-900/5">
            {result.data.map((product) => (
              <ProductRow key={product.id} product={product} />
            ))}
          </ul>

          <div className="flex items-center justify-between gap-3 border-t border-slate-900/5 bg-slate-50/60 px-6 py-3.5 text-sm text-slate-600">
            <span>
              Page {result.pagination.page} of {totalPages} ({result.pagination.total}{" "}
              total)
            </span>
            <div className="flex items-center gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1}
                onClick={() => goToPage(page - 1)}
              >
                <ChevronLeftIcon className="h-4 w-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => goToPage(page + 1)}
              >
                Next
                <ChevronRightIcon className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
}

function ProductRow({ product }: { product: Product }) {
  const addToCart = useAddToCart();
  const outOfStock = product.stock <= 0;

  function add() {
    addToCart.mutate({ productId: product.id, quantity: 1 });
  }

  return (
    <li className="flex flex-col gap-2 px-6 py-4">
      <div className="flex items-center justify-between gap-6">
        <div className="flex flex-col gap-1">
          <span className="font-medium text-slate-900">{product.name}</span>
          {product.description && (
            <span className="max-w-xl truncate text-xs text-slate-500">
              {product.description}
            </span>
          )}
          <div className="flex items-center gap-2">
            <Badge tone="brand">{product.category}</Badge>
            <span className="font-mono text-xs text-slate-500">{product.sku}</span>
          </div>
        </div>

        <div className="flex items-center gap-5">
          <div className="flex flex-col items-end gap-1">
            <span className="font-medium tabular-nums text-slate-900">
              <span className="mr-0.5 text-slate-400">$</span>
              {product.price}
            </span>
            {outOfStock ? (
              <span className="inline-flex items-center gap-1.5 text-xs text-slate-500">
                <span className="h-1.5 w-1.5 rounded-full bg-red-400" />
                Out of stock
              </span>
            ) : (
              <span className="inline-flex items-center gap-1.5 text-xs text-slate-700">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                {product.stock} in stock
              </span>
            )}
          </div>

          <Button
            size="sm"
            disabled={outOfStock || addToCart.isPending}
            onClick={add}
            aria-label={`Add ${product.name} to cart`}
          >
            <CartIcon className="h-4 w-4" />
            {outOfStock ? "Unavailable" : "Add to cart"}
          </Button>
        </div>
      </div>

      {addToCart.isError && (
        <p role="alert" className="text-right text-sm text-red-600">
          {cartErrorMessage(addToCart.error)}
        </p>
      )}
    </li>
  );
}

interface SearchInputProps {
  value: string;
  onChange: (value: string) => void;
  onSubmit: () => void;
}

function SearchInput({ value, onChange, onSubmit }: SearchInputProps) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor="store-search" className="text-sm font-medium text-slate-700">
        Search products
      </label>
      <input
        id="store-search"
        type="search"
        placeholder="Search by name, SKU, description, or category…"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter") onSubmit();
        }}
        className="w-full max-w-lg rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm text-slate-900 shadow-sm outline-none transition placeholder:text-slate-400 focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20"
      />
    </div>
  );
}

function ResultsSkeleton() {
  return (
    <div className="flex flex-col gap-4 p-6" aria-label="Loading products" role="status">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex items-center gap-6">
          <Skeleton className="h-9 w-9 rounded-lg" />
          <div className="flex flex-1 flex-col gap-1.5">
            <Skeleton className="h-3.5 w-1/3" />
            <Skeleton className="h-3 w-1/2" />
          </div>
          <Skeleton className="h-3.5 w-16" />
        </div>
      ))}
    </div>
  );
}
