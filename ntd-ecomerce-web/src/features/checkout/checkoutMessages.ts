import { ApiError } from "../../api/types";
import { cartErrorMessage } from "../cart/cartMessages";

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
