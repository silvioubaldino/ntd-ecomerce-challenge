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
import { ApiError, type Product, type ProductSort } from "../../api/types";
import { cartErrorMessage } from "../cart/cartMessages";
import { useAddToCart } from "../cart/hooks";
import { useCategories, useProductSearch } from "./hooks";
import { useCursorStack } from "./useCursorStack";

const SEARCH_DEBOUNCE_MS = 300;

const SORT_OPTIONS: Array<{ value: ProductSort | ""; label: string }> = [
  { value: "", label: "Default" },
  { value: "price_asc", label: "Price: low to high" },
  { value: "price_desc", label: "Price: high to low" },
  { value: "name_asc", label: "Name: A to Z" },
  { value: "name_desc", label: "Name: Z to A" },
  { value: "newest", label: "Newest" },
];

function parsePrice(value: string): number | null {
  const trimmed = value.trim();
  if (!trimmed) return null;
  const parsed = Number(trimmed);
  return Number.isNaN(parsed) ? null : parsed;
}

export function StoreSearchPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const q = searchParams.get("q") ?? "";
  const cursor = searchParams.get("cursor") ?? undefined;
  const category = searchParams.get("category") ?? "";
  const priceMinParam = searchParams.get("price_min") ?? "";
  const priceMaxParam = searchParams.get("price_max") ?? "";
  const sortParam = (searchParams.get("sort") ?? "") as ProductSort | "";

  const [inputValue, setInputValue] = useState(q);
  const debounceRef = useRef<ReturnType<typeof setTimeout>>();

  const [priceMinInput, setPriceMinInput] = useState(priceMinParam);
  const [priceMaxInput, setPriceMaxInput] = useState(priceMaxParam);
  const priceDebounceRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    setInputValue(q);
  }, [q]);

  useEffect(() => {
    setPriceMinInput(priceMinParam);
  }, [priceMinParam]);

  useEffect(() => {
    setPriceMaxInput(priceMaxParam);
  }, [priceMaxParam]);

  useEffect(() => {
    const trimmed = inputValue.trim();
    if (trimmed === q.trim()) return;

    debounceRef.current = setTimeout(() => {
      applySearch(inputValue);
    }, SEARCH_DEBOUNCE_MS);

    return () => clearTimeout(debounceRef.current);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [inputValue]);

  const priceRangeInvalid = (() => {
    const min = parsePrice(priceMinInput);
    const max = parsePrice(priceMaxInput);
    return min !== null && max !== null && min > max;
  })();

  useEffect(() => {
    if (priceRangeInvalid) return;
    if (
      priceMinInput.trim() === priceMinParam.trim() &&
      priceMaxInput.trim() === priceMaxParam.trim()
    ) {
      return;
    }

    priceDebounceRef.current = setTimeout(() => {
      applyFilters({ price_min: priceMinInput, price_max: priceMaxInput });
    }, SEARCH_DEBOUNCE_MS);

    return () => clearTimeout(priceDebounceRef.current);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [priceMinInput, priceMaxInput, priceRangeInvalid]);

  function setCursor(nextCursor: string | undefined) {
    const next = new URLSearchParams(searchParams);
    if (nextCursor) {
      next.set("cursor", nextCursor);
    } else {
      next.delete("cursor");
    }
    setSearchParams(next);
  }

  const cursorStack = useCursorStack(cursor, setCursor);

  function applyFilters(updates: Record<string, string | undefined>) {
    const next = new URLSearchParams(searchParams);
    for (const [key, value] of Object.entries(updates)) {
      const trimmed = value?.trim();
      if (trimmed) {
        next.set(key, trimmed);
      } else {
        next.delete(key);
      }
    }
    next.delete("cursor");
    cursorStack.reset();
    setSearchParams(next);
  }

  function applySearch(value: string) {
    clearTimeout(debounceRef.current);
    applyFilters({ q: value });
  }

  function clearFilters() {
    const next = new URLSearchParams(searchParams);
    next.delete("category");
    next.delete("price_min");
    next.delete("price_max");
    next.delete("sort");
    next.delete("cursor");
    cursorStack.reset();
    setSearchParams(next);
  }

  const hasActiveFilters = Boolean(category || priceMinParam || priceMaxParam || sortParam);

  const urlPriceRangeInvalid = (() => {
    const min = parsePrice(priceMinParam);
    const max = parsePrice(priceMaxParam);
    return min !== null && max !== null && min > max;
  })();

  const categoriesQuery = useCategories();
  const categories = categoriesQuery.data?.data ?? [];

  const searchQuery = useProductSearch({
    q,
    cursor,
    category,
    priceMin: priceMinParam,
    priceMax: priceMaxParam,
    sort: sortParam || undefined,
    enabled: !urlPriceRangeInvalid,
  });

  const invalidCursorError =
    searchQuery.error instanceof ApiError &&
    searchQuery.error.code === "validation_error" &&
    Boolean(searchQuery.error.details?.cursor);

  useEffect(() => {
    if (invalidCursorError && cursor) {
      cursorStack.reset();
      setCursor(undefined);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [invalidCursorError, cursor]);

  if (searchQuery.isError && !invalidCursorError) {
    return (
      <div className="flex flex-col gap-6">
        <PageHeader title="Store" description="Search the product catalog." />
        <SearchInput
          value={inputValue}
          onChange={setInputValue}
          onSubmit={() => applySearch(inputValue)}
        />
        <FilterBar
          category={category}
          categories={categories}
          priceMinInput={priceMinInput}
          priceMaxInput={priceMaxInput}
          onPriceMinChange={setPriceMinInput}
          onPriceMaxChange={setPriceMaxInput}
          priceRangeInvalid={priceRangeInvalid}
          sort={sortParam}
          onCategoryChange={(value) => applyFilters({ category: value })}
          onSortChange={(value) => applyFilters({ sort: value })}
          hasActiveFilters={hasActiveFilters}
          onClear={clearFilters}
        />
        <ErrorBanner
          message="Could not load products."
          onRetry={() => searchQuery.refetch()}
        />
      </div>
    );
  }

  const result = searchQuery.data;

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Store" description="Search the product catalog." />

      <SearchInput
        value={inputValue}
        onChange={setInputValue}
        onSubmit={() => applySearch(inputValue)}
      />

      <FilterBar
        category={category}
        categories={categories}
        priceMinInput={priceMinInput}
        priceMaxInput={priceMaxInput}
        onPriceMinChange={setPriceMinInput}
        onPriceMaxChange={setPriceMaxInput}
        priceRangeInvalid={priceRangeInvalid}
        sort={sortParam}
        onCategoryChange={(value) => applyFilters({ category: value })}
        onSortChange={(value) => applyFilters({ sort: value })}
        hasActiveFilters={hasActiveFilters}
        onClear={clearFilters}
      />

      {!result ? (
        <Card>
          <ResultsSkeleton />
        </Card>
      ) : result.data.length === 0 ? (
        <Card>
          <EmptyState
            icon={<BoxIcon className="h-7 w-7" />}
            title={
              hasActiveFilters
                ? "No products match your filters."
                : "No products match your search."
            }
            description={
              hasActiveFilters
                ? "Try widening your category, price range, or search term."
                : q
                  ? `No results for "${q}". Try a different term.`
                  : "The catalog is empty."
            }
            action={
              hasActiveFilters ? (
                <Button variant="secondary" size="sm" onClick={clearFilters}>
                  Clear filters
                </Button>
              ) : undefined
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

          <div className="flex items-center justify-end gap-2 border-t border-slate-900/5 bg-slate-50/60 px-6 py-3.5 text-sm text-slate-600">
            <Button
              variant="secondary"
              size="sm"
              disabled={!cursorStack.canGoPrev}
              onClick={() => cursorStack.goPrev()}
            >
              <ChevronLeftIcon className="h-4 w-4" />
              Previous
            </Button>
            <Button
              variant="secondary"
              size="sm"
              disabled={!result.pagination.next_cursor}
              onClick={() =>
                result.pagination.next_cursor && cursorStack.goNext(result.pagination.next_cursor)
              }
            >
              Next
              <ChevronRightIcon className="h-4 w-4" />
            </Button>
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
        placeholder="Search by name, SKU, or description…"
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

interface FilterBarProps {
  category: string;
  categories: string[];
  priceMinInput: string;
  priceMaxInput: string;
  onPriceMinChange: (value: string) => void;
  onPriceMaxChange: (value: string) => void;
  priceRangeInvalid: boolean;
  sort: ProductSort | "";
  onCategoryChange: (value: string) => void;
  onSortChange: (value: string) => void;
  hasActiveFilters: boolean;
  onClear: () => void;
}

function FilterBar({
  category,
  categories,
  priceMinInput,
  priceMaxInput,
  onPriceMinChange,
  onPriceMaxChange,
  priceRangeInvalid,
  sort,
  onCategoryChange,
  onSortChange,
  hasActiveFilters,
  onClear,
}: FilterBarProps) {
  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-end gap-4">
        <div className="flex flex-col gap-1.5">
          <label htmlFor="store-category" className="text-sm font-medium text-slate-700">
            Category
          </label>
          <select
            id="store-category"
            value={category}
            onChange={(event) => onCategoryChange(event.target.value)}
            className="input w-48"
          >
            <option value="">All categories</option>
            {categories.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1.5">
          <span className="text-sm font-medium text-slate-700">Price range</span>
          <div className="flex items-center gap-2">
            <input
              aria-label="Minimum price"
              inputMode="decimal"
              placeholder="Min"
              value={priceMinInput}
              onChange={(event) => onPriceMinChange(event.target.value)}
              className={`input w-24 ${priceRangeInvalid ? "input-invalid" : ""}`}
            />
            <span className="text-slate-400">–</span>
            <input
              aria-label="Maximum price"
              inputMode="decimal"
              placeholder="Max"
              value={priceMaxInput}
              onChange={(event) => onPriceMaxChange(event.target.value)}
              className={`input w-24 ${priceRangeInvalid ? "input-invalid" : ""}`}
            />
          </div>
        </div>

        <div className="flex flex-col gap-1.5">
          <label htmlFor="store-sort" className="text-sm font-medium text-slate-700">
            Sort by
          </label>
          <select
            id="store-sort"
            value={sort}
            onChange={(event) => onSortChange(event.target.value)}
            className="input w-48"
          >
            {SORT_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        {hasActiveFilters && (
          <Button variant="secondary" size="sm" onClick={onClear}>
            Clear filters
          </Button>
        )}
      </div>

      {priceRangeInvalid && (
        <p role="alert" className="text-sm text-red-600">
          Minimum price cannot exceed maximum price.
        </p>
      )}
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
