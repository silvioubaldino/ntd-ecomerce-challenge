import { useParams } from "react-router-dom";
import { PageHeader } from "../../components/PageHeader";
import { ButtonLink } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { EmptyState } from "../../components/ui/EmptyState";
import { Skeleton } from "../../components/ui/Skeleton";
import { BoxIcon, CheckCircleIcon } from "../../components/ui/icons";
import { ApiError } from "../../api/types";
import { useOrder } from "./hooks";

export function OrderConfirmationPage() {
  const { orderId } = useParams<{ orderId: string }>();
  const orderQuery = useOrder(orderId);

  const header = <PageHeader title="Order confirmation" />;

  if (orderQuery.isLoading) {
    return (
      <div className="flex flex-col gap-6">
        {header}
        <Card>
          <div className="flex flex-col gap-4 p-6" aria-label="Loading order" role="status">
            <Skeleton className="h-5 w-1/2" />
            <Skeleton className="h-4 w-1/3" />
            <Skeleton className="h-4 w-2/3" />
          </div>
        </Card>
      </div>
    );
  }

  const notFound =
    orderQuery.error instanceof ApiError && orderQuery.error.code === "order_not_found";

  if (notFound || !orderQuery.data) {
    return (
      <div className="flex flex-col gap-6">
        {header}
        <Card>
          <EmptyState
            icon={<BoxIcon className="h-7 w-7" />}
            title="Order not found."
            description="We couldn't find that order. It may have been removed or the link is incorrect."
            action={
              <ButtonLink to="/store" variant="secondary">
                Continue shopping
              </ButtonLink>
            }
          />
        </Card>
      </div>
    );
  }

  const order = orderQuery.data;

  return (
    <div className="flex flex-col gap-6">
      {header}

      <Card>
        <div className="flex flex-col gap-4 p-6">
          <div className="flex items-start gap-3">
            <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-emerald-50 text-emerald-600 ring-1 ring-inset ring-emerald-600/10">
              <CheckCircleIcon className="h-6 w-6" />
            </span>
            <div className="flex flex-col gap-1">
              <p className="text-lg font-semibold text-slate-900">Thank you for your order!</p>
              <p className="text-sm text-slate-600">
                Your order is <span className="font-medium text-emerald-700">confirmed</span> and
                payment was <span className="font-medium text-emerald-700">approved</span>{" "}
                (simulated).
              </p>
            </div>
          </div>

          <dl className="grid gap-x-6 gap-y-2 text-sm sm:grid-cols-2">
            <div className="flex flex-col">
              <dt className="text-slate-500">Order ID</dt>
              <dd className="font-mono text-slate-900">{order.id}</dd>
            </div>
            <div className="flex flex-col">
              <dt className="text-slate-500">Customer</dt>
              <dd className="text-slate-900">{order.customer.name}</dd>
              <dd className="text-slate-500">{order.customer.email}</dd>
            </div>
          </dl>
        </div>
      </Card>

      <Card>
        <ul className="divide-y divide-slate-900/5">
          {order.items.map((item) => (
            <li
              key={item.product_id}
              className="flex items-center justify-between gap-6 px-6 py-4"
            >
              <div className="flex flex-col gap-0.5">
                <span className="font-medium text-slate-900">{item.name}</span>
                <span className="font-mono text-xs text-slate-500">{item.sku}</span>
                <span className="text-xs text-slate-500">
                  <span className="text-slate-400">$</span>
                  {item.unit_price} × {item.quantity}
                </span>
              </div>
              <span className="w-20 text-right font-medium tabular-nums text-slate-900">
                <span className="mr-0.5 text-slate-400">$</span>
                {item.subtotal}
              </span>
            </li>
          ))}
        </ul>

        <div className="flex items-center justify-between gap-3 border-t border-slate-900/5 bg-slate-50/60 px-6 py-4">
          <span className="text-sm text-slate-600">Total</span>
          <span className="text-lg font-semibold tabular-nums text-slate-900">
            <span className="mr-0.5 text-slate-400">$</span>
            {order.total}
          </span>
        </div>
      </Card>

      <div className="flex items-center justify-end">
        <ButtonLink to="/store" variant="secondary">
          Continue shopping
        </ButtonLink>
      </div>
    </div>
  );
}
