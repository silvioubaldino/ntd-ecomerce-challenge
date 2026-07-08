import { useState } from "react";
import { Link } from "react-router-dom";
import { ErrorBanner } from "../../components/ErrorBanner";
import { useDeleteProduct, useProducts } from "./hooks";

export function ProductListPage() {
  const [page, setPage] = useState(1);
  const productsQuery = useProducts(page);
  const deleteProduct = useDeleteProduct();

  if (productsQuery.isLoading) {
    return <p className="text-sm text-gray-500">Loading…</p>;
  }

  if (productsQuery.isError) {
    return (
      <ErrorBanner
        message="Could not load the catalog."
        onRetry={() => productsQuery.refetch()}
      />
    );
  }

  const { data, pagination } = productsQuery.data!;
  const totalPages = Math.max(1, Math.ceil(pagination.total / pagination.page_size));

  function handleDelete(id: string) {
    if (!window.confirm("Delete this product?")) return;
    deleteProduct.mutate(id);
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-xl font-semibold text-gray-900">Products</h1>

      {deleteProduct.isError && (
        <ErrorBanner message="Could not delete this product. Please try again." />
      )}

      {data.length === 0 ? (
        <p className="text-sm text-gray-500">No products yet.</p>
      ) : (
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b text-gray-500">
              <th className="py-2">Name</th>
              <th className="py-2">SKU</th>
              <th className="py-2">Category</th>
              <th className="py-2">Price</th>
              <th className="py-2">Stock</th>
              <th className="py-2" />
            </tr>
          </thead>
          <tbody>
            {data.map((product) => (
              <tr key={product.id} className="border-b">
                <td className="py-2">{product.name}</td>
                <td className="py-2">{product.sku}</td>
                <td className="py-2">{product.category}</td>
                <td className="py-2">{product.price}</td>
                <td className="py-2">{product.stock}</td>
                <td className="flex gap-3 py-2">
                  <Link
                    to={`/products/${product.id}/edit`}
                    className="text-blue-600 hover:underline"
                  >
                    Edit
                  </Link>
                  <button
                    type="button"
                    onClick={() => handleDelete(product.id)}
                    className="text-red-600 hover:underline"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <div className="flex items-center gap-3 text-sm">
        <button
          type="button"
          disabled={page <= 1}
          onClick={() => setPage((p) => p - 1)}
          className="rounded border px-3 py-1 disabled:opacity-50"
        >
          Previous
        </button>
        <span>
          Page {pagination.page} of {totalPages} ({pagination.total} total)
        </span>
        <button
          type="button"
          disabled={page >= totalPages}
          onClick={() => setPage((p) => p + 1)}
          className="rounded border px-3 py-1 disabled:opacity-50"
        >
          Next
        </button>
      </div>
    </div>
  );
}
