import { ApiError } from "../../api/types";

export function cartErrorMessage(error: unknown): string {
  if (!(error instanceof ApiError)) {
    return "Something went wrong. Please try again.";
  }

  switch (error.code) {
    case "insufficient_stock": {
      const available = error.details?.available;
      const requested = error.details?.requested;
      if (available !== undefined && requested !== undefined) {
        return `Only ${available} in stock (you requested ${requested}).`;
      }
      return "Not enough stock for that quantity.";
    }
    case "validation_error":
      return "That quantity is not valid. Enter a whole number of 1 or more.";
    case "product_not_found":
      return "This product is no longer available.";
    case "item_not_found":
      return "This item is no longer in your cart.";
    default:
      return error.message || "Something went wrong. Please try again.";
  }
}
