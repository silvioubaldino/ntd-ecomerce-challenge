import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { ErrorBanner } from "../../components/ErrorBanner";
import { PageHeader } from "../../components/PageHeader";
import { Button, ButtonLink } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { EmptyState } from "../../components/ui/EmptyState";
import { Skeleton } from "../../components/ui/Skeleton";
import { CartIcon } from "../../components/ui/icons";
import { ApiError } from "../../api/types";
import { useCart, useCartId } from "../cart/hooks";
import { checkoutErrorMessage } from "./checkoutMessages";
import { useCheckout } from "./hooks";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

interface FieldErrors {
  name?: string;
  email?: string;
}

export function CheckoutPage() {
  const cartId = useCartId();
  const cartQuery = useCart();
  const checkout = useCheckout();
  const navigate = useNavigate();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [topError, setTopError] = useState<string | null>(null);
  // Set when the api reports the cart is empty/gone at submit time.
  const [consumed, setConsumed] = useState(false);

  const header = (
    <PageHeader
      title="Checkout"
      description="Review your order and enter your contact details."
      backTo={{ to: "/cart", label: "Back to cart" }}
    />
  );

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
          <div className="flex flex-col gap-4 p-6" aria-label="Loading cart" role="status">
            <Skeleton className="h-4 w-1/3" />
            <Skeleton className="h-4 w-1/4" />
            <Skeleton className="h-4 w-1/2" />
          </div>
        </Card>
      </div>
    );
  }

  const cart = cartQuery.data;
  const isEmpty = consumed || cartId === null || !cart || cart.items.length === 0;

  if (isEmpty) {
    return (
      <div className="flex flex-col gap-6">
        {header}
        <Card>
          <EmptyState
            icon={<CartIcon className="h-7 w-7" />}
            title="Your cart is empty."
            description="Add products to your cart before checking out."
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

  function validate(): FieldErrors {
    const errors: FieldErrors = {};
    if (name.trim() === "") {
      errors.name = "Please enter your name.";
    }
    if (email.trim() === "") {
      errors.email = "Please enter your email.";
    } else if (!EMAIL_RE.test(email.trim())) {
      errors.email = "Please enter a valid email address.";
    }
    return errors;
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setTopError(null);

    const errors = validate();
    setFieldErrors(errors);
    if (Object.keys(errors).length > 0) {
      return;
    }

    checkout.mutate(
      { name: name.trim(), email: email.trim() },
      {
        onSuccess: (order) => navigate(`/orders/${order.id}`),
        onError: (error) => {
          if (error instanceof ApiError) {
            if (error.code === "cart_empty" || error.code === "cart_not_found") {
              setConsumed(true);
              return;
            }
            if (error.code === "validation_error" && error.details) {
              const details = error.details;
              setFieldErrors({
                name: details.name,
                email: details.email,
              });
              if (!details.name && !details.email) {
                setTopError(checkoutErrorMessage(error));
              }
              return;
            }
          }
          setTopError(checkoutErrorMessage(error));
        },
      },
    );
  }

  const pending = checkout.isPending;

  return (
    <div className="flex flex-col gap-6">
      {header}

      {topError && <ErrorBanner message={topError} />}

      <Card>
        <ul className="divide-y divide-slate-900/5">
          {cart.items.map((item) => (
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
            {cart.total}
          </span>
        </div>
      </Card>

      <Card>
        <form className="flex flex-col gap-5 p-6" onSubmit={handleSubmit} noValidate>
          <div className="flex flex-col gap-1.5">
            <label htmlFor="checkout-name" className="text-sm font-medium text-slate-700">
              Full name
            </label>
            <input
              id="checkout-name"
              name="name"
              type="text"
              value={name}
              onChange={(event) => setName(event.target.value)}
              aria-invalid={fieldErrors.name ? true : undefined}
              aria-describedby={fieldErrors.name ? "checkout-name-error" : undefined}
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 shadow-sm focus-visible:border-brand-500 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-brand-600/20"
            />
            {fieldErrors.name && (
              <p id="checkout-name-error" role="alert" className="text-sm text-red-600">
                {fieldErrors.name}
              </p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <label htmlFor="checkout-email" className="text-sm font-medium text-slate-700">
              Email
            </label>
            <input
              id="checkout-email"
              name="email"
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              aria-invalid={fieldErrors.email ? true : undefined}
              aria-describedby={fieldErrors.email ? "checkout-email-error" : undefined}
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 shadow-sm focus-visible:border-brand-500 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-brand-600/20"
            />
            {fieldErrors.email && (
              <p id="checkout-email-error" role="alert" className="text-sm text-red-600">
                {fieldErrors.email}
              </p>
            )}
          </div>

          <p className="text-xs text-slate-500">
            Payment is simulated for this demo — no real charge is made.
          </p>

          <div className="flex items-center justify-end">
            <Button type="submit" disabled={pending}>
              {pending ? "Placing order…" : "Place order"}
            </Button>
          </div>
        </form>
      </Card>
    </div>
  );
}
