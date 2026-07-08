import { useNavigate, useParams } from "react-router-dom";
import { ApiError } from "../../api/types";
import { ErrorBanner } from "../../components/ErrorBanner";
import { useProduct, useUpdateProduct } from "./hooks";
import { ProductForm } from "./ProductForm";
import type { ProductFormValues } from "./schema";

export function ProductEditPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const productQuery = useProduct(id!);
  const updateProduct = useUpdateProduct(id!);

  async function handleSubmit(values: ProductFormValues) {
    await updateProduct.mutateAsync(values);
    navigate("/products");
  }

  if (productQuery.isLoading) {
    return <p className="text-sm text-gray-500">Loading…</p>;
  }

  if (productQuery.isError) {
    const error = productQuery.error;
    if (error instanceof ApiError && error.code === "product_not_found") {
      return <p className="text-sm text-gray-700">Product not found.</p>;
    }
    return (
      <ErrorBanner
        message="Could not load this product."
        onRetry={() => productQuery.refetch()}
      />
    );
  }

  const product = productQuery.data!;

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-xl font-semibold text-gray-900">Edit Product</h1>
      <ProductForm
        defaultValues={{
          name: product.name,
          sku: product.sku,
          description: product.description,
          category: product.category,
          price: product.price,
          stock: product.stock,
          weight_kg: product.weight_kg,
        }}
        onSubmit={handleSubmit}
        submitLabel="Save"
      />
    </div>
  );
}
