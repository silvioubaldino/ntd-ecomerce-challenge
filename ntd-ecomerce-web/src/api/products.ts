import { apiClient } from "./client";
import type { CategoryList, Product, ProductInput, ProductList, ProductSort } from "./types";

export interface ListProductsOptions {
  page: number;
  pageSize?: number;
  q?: string;
  category?: string;
  priceMin?: string;
  priceMax?: string;
  sort?: ProductSort;
}

export function listProducts({
  page,
  pageSize = 20,
  q,
  category,
  priceMin,
  priceMax,
  sort,
}: ListProductsOptions): Promise<ProductList> {
  const params = new URLSearchParams();
  params.set("page", String(page));
  params.set("page_size", String(pageSize));

  const trimmedQ = q?.trim();
  if (trimmedQ) params.set("q", trimmedQ);

  const trimmedCategory = category?.trim();
  if (trimmedCategory) params.set("category", trimmedCategory);

  const trimmedPriceMin = priceMin?.trim();
  if (trimmedPriceMin) params.set("price_min", trimmedPriceMin);

  const trimmedPriceMax = priceMax?.trim();
  if (trimmedPriceMax) params.set("price_max", trimmedPriceMax);

  if (sort) params.set("sort", sort);

  return apiClient.get<ProductList>(`/products?${params.toString()}`);
}

export function listCategories(): Promise<CategoryList> {
  return apiClient.get<CategoryList>("/products/categories");
}

export function getProduct(id: string): Promise<Product> {
  return apiClient.get<Product>(`/products/${id}`);
}

export function createProduct(input: ProductInput): Promise<Product> {
  return apiClient.post<Product>("/products", input);
}

export function updateProduct(id: string, input: ProductInput): Promise<Product> {
  return apiClient.put<Product>(`/products/${id}`, input);
}

export function deleteProduct(id: string): Promise<void> {
  return apiClient.delete<void>(`/products/${id}`);
}
