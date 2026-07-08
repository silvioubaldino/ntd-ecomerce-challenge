import { useState } from "react";
import { Link } from "react-router-dom";
import { ErrorBanner } from "../../components/ErrorBanner";
import { PageHeader } from "../../components/PageHeader";
import { Badge } from "../../components/ui/Badge";
import { Button, ButtonLink } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { ConfirmDialog } from "../../components/ui/ConfirmDialog";
import { EmptyState } from "../../components/ui/EmptyState";
import { Skeleton } from "../../components/ui/Skeleton";
import {
  BoxIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  PencilIcon,
  PlusIcon,
  TrashIcon,
  UploadIcon,
} from "../../components/ui/icons";
import type { Product } from "../../api/types";
import { useDeleteProduct, useProducts } from "./hooks";

const headerCell =
  "px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 first:pl-6 last:pr-6";
const bodyCell = "px-4 py-3.5 first:pl-6 last:pr-6";

export function ProductListPage() {
  const [page, setPage] = useState(1);
  const [productToDelete, setProductToDelete] = useState<Product | null>(null);
  const productsQuery = useProducts(page);
  const deleteProduct = useDeleteProduct();

  function confirmDelete() {
    if (!productToDelete) return;
    deleteProduct.mutate(productToDelete.id, {
      onSettled: () => setProductToDelete(null),
    });
  }

  if (productsQuery.isError) {
    return (
      <div className="flex flex-col gap-6">
        <PageHeader title="Products" description="Manage your marketplace catalog." />
        <ErrorBanner
          message="Could not load the catalog."
          onRetry={() => productsQuery.refetch()}
        />
      </div>
    );
  }

  const result = productsQuery.data;
  const totalPages = result
    ? Math.max(1, Math.ceil(result.pagination.total / result.pagination.page_size))
    : 1;

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Products"
        description="Manage your marketplace catalog."
        actions={
          <>
            <ButtonLink to="/products/import" variant="secondary">
              <UploadIcon className="h-4 w-4" />
              Import CSV
            </ButtonLink>
            <ButtonLink to="/products/new">
              <PlusIcon className="h-4 w-4" />
              New Product
            </ButtonLink>
          </>
        }
      />

      {deleteProduct.isError && (
        <ErrorBanner message="Could not delete this product. Please try again." />
      )}

      {!result ? (
        <Card>
          <TableSkeleton />
        </Card>
      ) : result.data.length === 0 ? (
        <Card>
          <EmptyState
            icon={<BoxIcon className="h-7 w-7" />}
            title="No products yet."
            description="Your catalog is empty. Add your first product to start selling."
            action={
              <ButtonLink to="/products/new" variant="secondary">
                <PlusIcon className="h-4 w-4" />
                Add your first product
              </ButtonLink>
            }
          />
        </Card>
      ) : (
        <Card>
          <table className="w-full text-sm">
            <thead className="border-b border-slate-900/5 bg-slate-50/60">
              <tr>
                <th className={headerCell}>Product</th>
                <th className={headerCell}>SKU</th>
                <th className={headerCell}>Category</th>
                <th className={`${headerCell} text-right`}>Price</th>
                <th className={headerCell}>Stock</th>
                <th className={headerCell}>
                  <span className="sr-only">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-900/5">
              {result.data.map((product) => (
                <tr key={product.id} className="transition hover:bg-slate-50/70">
                  <td className={bodyCell}>
                    <div className="flex flex-col">
                      <span className="font-medium text-slate-900">{product.name}</span>
                      {product.description && (
                        <span className="max-w-xs truncate text-xs text-slate-500">
                          {product.description}
                        </span>
                      )}
                    </div>
                  </td>
                  <td className={bodyCell}>
                    <span className="font-mono text-xs text-slate-600">
                      {product.sku}
                    </span>
                  </td>
                  <td className={bodyCell}>
                    <Badge tone="brand">{product.category}</Badge>
                  </td>
                  <td className={`${bodyCell} text-right`}>
                    <span className="mr-0.5 text-slate-400">$</span>
                    <span className="font-medium tabular-nums text-slate-900">
                      {product.price}
                    </span>
                  </td>
                  <td className={bodyCell}>
                    {product.stock > 0 ? (
                      <span className="inline-flex items-center gap-1.5 text-slate-700">
                        <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                        {product.stock} in stock
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1.5 text-slate-500">
                        <span className="h-1.5 w-1.5 rounded-full bg-red-400" />
                        Out of stock
                      </span>
                    )}
                  </td>
                  <td className={bodyCell}>
                    <div className="flex items-center justify-end gap-1">
                      <Link
                        to={`/products/${product.id}/edit`}
                        aria-label="Edit"
                        title="Edit"
                        className="rounded-lg p-2 text-slate-500 transition hover:bg-brand-50 hover:text-brand-700"
                      >
                        <PencilIcon className="h-4 w-4" />
                      </Link>
                      <button
                        type="button"
                        aria-label="Delete"
                        title="Delete"
                        onClick={() => setProductToDelete(product)}
                        className="rounded-lg p-2 text-slate-500 transition hover:bg-red-50 hover:text-red-600"
                      >
                        <TrashIcon className="h-4 w-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          <div className="flex items-center justify-between gap-3 border-t border-slate-900/5 bg-slate-50/60 px-6 py-3.5 text-sm text-slate-600">
            <span>
              Page {result.pagination.page} of {totalPages} ({result.pagination.total}{" "}
              total)
            </span>
            <div className="flex items-center gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                <ChevronLeftIcon className="h-4 w-4" />
                Previous
              </Button>
              <Button
                variant="secondary"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Next
                <ChevronRightIcon className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </Card>
      )}

      <ConfirmDialog
        open={productToDelete !== null}
        title="Delete product?"
        description={
          productToDelete
            ? `"${productToDelete.name}" will be permanently removed from the catalog. This action cannot be undone.`
            : ""
        }
        confirmLabel="Delete product"
        confirming={deleteProduct.isPending}
        onConfirm={confirmDelete}
        onCancel={() => setProductToDelete(null)}
      />
    </div>
  );
}

function TableSkeleton() {
  return (
    <div className="flex flex-col gap-4 p-6" aria-label="Loading products" role="status">
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="flex items-center gap-6">
          <Skeleton className="h-9 w-9 rounded-lg" />
          <div className="flex flex-1 flex-col gap-1.5">
            <Skeleton className="h-3.5 w-1/3" />
            <Skeleton className="h-3 w-1/2" />
          </div>
          <Skeleton className="h-3.5 w-16" />
          <Skeleton className="h-3.5 w-12" />
        </div>
      ))}
    </div>
  );
}
