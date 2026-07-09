import { ApiError } from "../../api/types";
import { checkoutErrorMessage } from "./checkoutMessages";

describe("checkoutErrorMessage", () => {
  it("maps cart_empty to a friendly message", () => {
    const error = new ApiError(422, { code: "cart_empty", message: "cart is empty" });
    expect(checkoutErrorMessage(error)).toBe("Your cart is empty.");
  });

  it("maps cart_not_found to a friendly message", () => {
    const error = new ApiError(404, { code: "cart_not_found", message: "no cart" });
    expect(checkoutErrorMessage(error)).toMatch(/no longer available/i);
  });

  it("reuses the cart mapping for insufficient_stock", () => {
    const error = new ApiError(409, {
      code: "insufficient_stock",
      message: "insufficient stock",
      details: { product_id: "p1", requested: "5", available: "3" },
    });
    expect(checkoutErrorMessage(error)).toBe("Only 3 in stock (you requested 5).");
  });

  it("falls back to the api message for unknown codes", () => {
    const error = new ApiError(500, { code: "internal_error", message: "boom" });
    expect(checkoutErrorMessage(error)).toBe("boom");
  });

  it("handles non-ApiError values", () => {
    expect(checkoutErrorMessage(new Error("nope"))).toMatch(/something went wrong/i);
  });
});
