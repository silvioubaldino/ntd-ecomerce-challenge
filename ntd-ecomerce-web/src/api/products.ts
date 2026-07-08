import { apiClient } from "./client";
import type { Product, ProductInput, ProductList } from "./types";

export function listProducts(
  page: number,
  pageSize = 20,
  q?: string,
): Promise<ProductList> {
  const trimmedQ = q?.trim();
  const query = `page=${page}&page_size=${pageSize}${
    trimmedQ ? `&q=${encodeURIComponent(trimmedQ)}` : ""
  }`;
  return apiClient.get<ProductList>(`/products?${query}`);
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
