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

export interface CartItem {
  product_id: string;
  sku: string;
  name: string;
  unit_price: string;
  quantity: number;
  subtotal: string;
}

export interface Cart {
  id: string;
  items: CartItem[];
  total: string;
  created_at: string;
  updated_at: string;
}

export interface ImportSummary {
  total: number;
  imported: number;
  rejected: number;
}

export interface RejectedRow {
  row: number;
  sku: string;
  errors: Record<string, string>;
}

export interface ImportReport {
  summary: ImportSummary;
  rejected: RejectedRow[];
}

export type ApiErrorCode =
  | "validation_error"
  | "sku_already_exists"
  | "product_not_found"
  | "cart_not_found"
  | "item_not_found"
  | "insufficient_stock"
  | "invalid_header"
  | "invalid_file"
  | "file_too_large"
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
