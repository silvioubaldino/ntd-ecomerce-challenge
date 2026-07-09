import { ApiError } from "../../api/types";
import { cartErrorMessage } from "../cart/cartMessages";

// Maps AYD-005 checkout error codes to customer-facing copy. Money/quantities from
// `details` are rendered verbatim (no float math). Reuses the cart mapping for the codes
// shared with AYD-004 (insufficient_stock, validation_error); adds the checkout-specific
// ones; falls back to the api message.
export function checkoutErrorMessage(error: unknown): string {
  if (!(error instanceof ApiError)) {
    return "Something went wrong. Please try again.";
  }

  switch (error.code) {
    case "cart_empty":
      return "Your cart is empty.";
    case "cart_not_found":
      return "Your cart is no longer available. Add items again to check out.";
    default:
      return cartErrorMessage(error);
  }
}
