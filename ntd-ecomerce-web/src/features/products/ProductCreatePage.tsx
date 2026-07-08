import { useNavigate } from "react-router-dom";
import { useCreateProduct } from "./hooks";
import { ProductForm } from "./ProductForm";
import type { ProductFormValues } from "./schema";

export function ProductCreatePage() {
  const navigate = useNavigate();
  const createProduct = useCreateProduct();

  async function handleSubmit(values: ProductFormValues) {
    await createProduct.mutateAsync(values);
    navigate("/products");
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-xl font-semibold text-gray-900">New Product</h1>
      <ProductForm onSubmit={handleSubmit} submitLabel="Create" />
    </div>
  );
}
