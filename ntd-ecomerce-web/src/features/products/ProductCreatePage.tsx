import { useNavigate } from "react-router-dom";
import { PageHeader } from "../../components/PageHeader";
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
    <div className="flex flex-col gap-6">
      <PageHeader
        title="New Product"
        description="Add a product to the catalog."
        backTo={{ to: "/products", label: "Back to products" }}
      />
      <ProductForm onSubmit={handleSubmit} submitLabel="Create" />
    </div>
  );
}
