import { apiClient } from "./client";
import type { ImportReport } from "./types";

export function importProducts(file: File): Promise<ImportReport> {
  const formData = new FormData();
  formData.append("file", file);
  return apiClient.postForm<ImportReport>("/products/import", formData);
}
