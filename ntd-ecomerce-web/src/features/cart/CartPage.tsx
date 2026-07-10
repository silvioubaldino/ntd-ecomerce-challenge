import { useState } from "react";
import { ErrorBanner } from "../../components/ErrorBanner";
import { PageHeader } from "../../components/PageHeader";
import { Button, ButtonLink } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { EmptyState } from "../../components/ui/EmptyState";
import { Skeleton } from "../../components/ui/Skeleton";
import {
  CartIcon,
  ChevronRightIcon,
  MinusIcon,
  PlusIcon,
  TrashIcon,
} from "../../components/ui/icons";
import type { CartItem } from "../../api/types";
import { cartErrorMessage } from "./cartMessages";
import { useCart, useCartId, useRemoveCartItem, useUpdateCartItem } from "./hooks";

export function CartPage() {
  const cartId = useCartId();
  const cartQuery = useCart();
  const updateItem = useUpdateCartItem();
  const removeItem = useRemoveCartItem();

  const [lineError, setLineError] = useState<{ productId: string; message: string } | null>(
    null,
  );

  function changeQuantity(item: CartItem, quantity: number) {
    setLineError(null);
    updateItem.mutate(
      { productId: item.product_id, quantity },
      {
        onError: (error) =>
          setLineError({ productId: item.product_id, message: cartErrorMessage(error) }),
      },
    );
  }

  function remove(item: CartItem) {
    setLineError(null);
    removeItem.mutate(
      { productId: item.product_id },
      {
        onError: (error) =>
          setLineError({ productId: item.product_id, message: cartErrorMessage(error) }),
      },
    );
  }

  const header = <PageHeader title="Your cart" description="Review your items before checkout." />;

  if (cartQuery.isError) {
    return (
      <div className="flex flex-col gap-6">
        {header}
        <ErrorBanner message="Could not load your cart." onRetry={() => cartQuery.refetch()} />
      </div>
    );
  }

  if (cartId !== null && cartQuery.isLoading) {
    return (
      <div className="flex flex-col gap-6">
        {header}
        <Card>
          <CartSkeleton />
        </Card>
      </div>
    );
  }

  const cart = cartQuery.data;
  const isEmpty = cartId === null || !cart || cart.items.length === 0;

  if (isEmpty) {
    return (
      <div className="flex flex-col gap-6">
        {header}
        <Card>
          <EmptyState
            icon={<CartIcon className="h-7 w-7" />}
            title="Your cart is empty."
            description="Browse the store and add products to start your order."
            action={
              <ButtonLink to="/store" variant="secondary">
                Go to the store
              </ButtonLink>
            }
          />
        </Card>
      </div>
    );
  }

  const busy = updateItem.isPending || removeItem.isPending;

  return (
    <div className="flex flex-col gap-6">
      {header}

      <Card>
        <ul className="divide-y divide-slate-900/5">
          {cart.items.map((item) => (
            <li key={item.product_id} className="flex flex-col gap-2 px-6 py-4">
              <div className="flex items-center justify-between gap-6">
                <div className="flex flex-col gap-0.5">
                  <span className="font-medium text-slate-900">{item.name}</span>
                  <span className="font-mono text-xs text-slate-500">{item.sku}</span>
                  <span className="text-xs text-slate-500">
                    <span className="text-slate-400">$</span>
                    {item.unit_price} each
                  </span>
                </div>

                <div className="flex items-center gap-4">
                  <div className="flex items-center gap-1.5">
                    <Button
                      variant="secondary"
                      size="sm"
                      aria-label="Decrease quantity"
                      disabled={busy || item.quantity <= 1}
                      onClick={() => changeQuantity(item, item.quantity - 1)}
                    >
                      <MinusIcon className="h-4 w-4" />
                    </Button>
                    <span
                      aria-label={`Quantity for ${item.name}`}
                      className="w-8 text-center tabular-nums text-slate-900"
                    >
                      {item.quantity}
                    </span>
                    <Button
                      variant="secondary"
                      size="sm"
                      aria-label="Increase quantity"
                      disabled={busy}
                      onClick={() => changeQuantity(item, item.quantity + 1)}
                    >
                      <PlusIcon className="h-4 w-4" />
                    </Button>
                  </div>

                  <span className="w-20 text-right font-medium tabular-nums text-slate-900">
                    <span className="mr-0.5 text-slate-400">$</span>
                    {item.subtotal}
                  </span>

                  <button
                    type="button"
                    aria-label={`Remove ${item.name}`}
                    title="Remove"
                    disabled={busy}
                    onClick={() => remove(item)}
                    className="rounded-lg p-2 text-slate-500 transition hover:bg-red-50 hover:text-red-600 disabled:pointer-events-none disabled:opacity-50"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                </div>
              </div>

              {lineError?.productId === item.product_id && (
                <p role="alert" className="text-sm text-red-600">
                  {lineError.message}
                </p>
              )}
            </li>
          ))}
        </ul>

        <div className="flex items-center justify-between gap-3 border-t border-slate-900/5 bg-slate-50/60 px-6 py-4">
          <span className="text-sm text-slate-600">Total</span>
          <span className="text-lg font-semibold tabular-nums text-slate-900">
            <span className="mr-0.5 text-slate-400">$</span>
            {cart.total}
          </span>
        </div>
      </Card>

      <div className="flex items-center justify-end">
        <ButtonLink to="/checkout">
          Proceed to checkout
          <ChevronRightIcon className="h-4 w-4" />
        </ButtonLink>
      </div>
    </div>
  );
}

function CartSkeleton() {
  return (
    <div className="flex flex-col gap-4 p-6" aria-label="Loading cart" role="status">
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className="flex items-center justify-between gap-6">
          <div className="flex flex-1 flex-col gap-1.5">
            <Skeleton className="h-3.5 w-1/3" />
            <Skeleton className="h-3 w-1/4" />
          </div>
          <Skeleton className="h-8 w-28" />
          <Skeleton className="h-3.5 w-14" />
        </div>
      ))}
    </div>
  );
}
