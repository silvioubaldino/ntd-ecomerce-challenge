import { useNavigate, useParams } from "react-router-dom";
import { ApiError } from "../../api/types";
import { ErrorBanner } from "../../components/ErrorBanner";
import { PageHeader } from "../../components/PageHeader";
import { ButtonLink } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { EmptyState } from "../../components/ui/EmptyState";
import { Skeleton } from "../../components/ui/Skeleton";
import { BoxIcon } from "../../components/ui/icons";
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

  const header = (
    <PageHeader
      title="Edit Product"
      description="Update this product's details."
      backTo={{ to: "/products", label: "Back to products" }}
    />
  );

  if (productQuery.isLoading) {
    return (
      <div className="flex flex-col gap-6">
        {header}
        <Card className="max-w-2xl p-6 sm:p-8" aria-label="Loading product" role="status">
          <div className="flex flex-col gap-5">
            <Skeleton className="h-4 w-32" />
            <div className="grid gap-5 sm:grid-cols-2">
              <Skeleton className="h-10" />
              <Skeleton className="h-10" />
            </div>
            <Skeleton className="h-20" />
            <Skeleton className="h-10" />
          </div>
        </Card>
      </div>
    );
  }

  if (productQuery.isError) {
    const error = productQuery.error;
    if (error instanceof ApiError && error.code === "product_not_found") {
      return (
        <div className="flex flex-col gap-6">
          {header}
          <Card className="max-w-2xl">
            <EmptyState
              icon={<BoxIcon className="h-7 w-7" />}
              title="Product not found."
              description="It may have been removed from the catalog."
              action={
                <ButtonLink to="/products" variant="secondary">
                  Back to products
                </ButtonLink>
              }
            />
          </Card>
        </div>
      );
    }
    return (
      <div className="flex flex-col gap-6">
        {header}
        <ErrorBanner
          message="Could not load this product."
          onRetry={() => productQuery.refetch()}
        />
      </div>
    );
  }

  const product = productQuery.data!;

  return (
    <div className="flex flex-col gap-6">
      {header}
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
