import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createProduct,
  deleteProduct,
  getProduct,
  listCategories,
  listProducts,
  updateProduct,
} from "../../api/products";
import { importProducts } from "../../api/import";
import type { ProductInput, ProductSort } from "../../api/types";

export interface ProductSearchFilters {
  q: string;
  cursor?: string;
  category?: string;
  priceMin?: string;
  priceMax?: string;
  sort?: ProductSort;
  enabled?: boolean;
}

const productKeys = {
  list: (cursor?: string) => ["products", "list", cursor ?? ""] as const,
  search: (filters: ProductSearchFilters) =>
    [
      "products",
      "search",
      filters.q.trim(),
      filters.cursor ?? "",
      filters.category?.trim() ?? "",
      filters.priceMin?.trim() ?? "",
      filters.priceMax?.trim() ?? "",
      filters.sort ?? "",
    ] as const,
  detail: (id: string) => ["products", "detail", id] as const,
  categories: () => ["products", "categories"] as const,
};

export function useProducts(cursor?: string) {
  return useQuery({
    queryKey: productKeys.list(cursor),
    queryFn: () => listProducts({ cursor }),
  });
}

export function useProductSearch(filters: ProductSearchFilters) {
  return useQuery({
    queryKey: productKeys.search(filters),
    queryFn: () =>
      listProducts({
        cursor: filters.cursor,
        limit: 20,
        q: filters.q,
        category: filters.category,
        priceMin: filters.priceMin,
        priceMax: filters.priceMax,
        sort: filters.sort,
      }),
    placeholderData: (previousData) => previousData,
    enabled: filters.enabled ?? true,
  });
}

export function useCategories() {
  return useQuery({
    queryKey: productKeys.categories(),
    queryFn: () => listCategories(),
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}

export function useProduct(id: string) {
  return useQuery({
    queryKey: productKeys.detail(id),
    queryFn: () => getProduct(id),
  });
}

export function useCreateProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ProductInput) => createProduct(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products", "list"] });
    },
  });
}

export function useUpdateProduct(id: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: ProductInput) => updateProduct(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products", "list"] });
      queryClient.invalidateQueries({ queryKey: productKeys.detail(id) });
    },
  });
}

export function useDeleteProduct() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteProduct(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products", "list"] });
    },
  });
}

export function useImportProducts() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => importProducts(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products", "list"] });
    },
  });
}
