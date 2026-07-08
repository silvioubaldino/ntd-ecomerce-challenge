export interface Product {
  id: string;
  name: string;
  sku: string;
  description: string;
  category: string;
  price: string;
  stock: number;
  weight_kg: string;
  created_at: string;
  updated_at: string;
}

export type ProductInput = Omit<
  Product,
  "id" | "created_at" | "updated_at"
>;

export interface Pagination {
  page: number;
  page_size: number;
  total: number;
}

export interface ProductList {
  data: Product[];
  pagination: Pagination;
}

export type ApiErrorCode =
  | "validation_error"
  | "sku_already_exists"
  | "product_not_found"
  | "internal_error"
  | string;

export interface ApiErrorBody {
  code: ApiErrorCode;
  message: string;
  details?: Record<string, string>;
}

export class ApiError extends Error {
  code: ApiErrorCode;
  status: number;
  details?: Record<string, string>;

  constructor(status: number, body: ApiErrorBody) {
    super(body.message);
    this.name = "ApiError";
    this.status = status;
    this.code = body.code;
    this.details = body.details;
  }
}
