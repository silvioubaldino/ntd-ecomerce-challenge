import { ApiError } from "../../api/types";
import { cartErrorMessage } from "./cartMessages";

describe("cartErrorMessage", () => {
  it("reports requested vs available for insufficient_stock", () => {
    // Arrange
    const error = new ApiError(409, {
      code: "insufficient_stock",
      message: "insufficient stock",
      details: { product_id: "p1", requested: "5", available: "3" },
    });

    // Act
    const message = cartErrorMessage(error);

    // Assert
    expect(message).toBe("Only 3 in stock (you requested 5).");
  });

  it("maps validation_error to a quantity hint", () => {
    const error = new ApiError(422, {
      code: "validation_error",
      message: "invalid",
      details: { quantity: "must_be_positive_integer" },
    });
    expect(cartErrorMessage(error)).toMatch(/quantity is not valid/i);
  });

  it("maps product_not_found and item_not_found to friendly copy", () => {
    expect(
      cartErrorMessage(new ApiError(404, { code: "product_not_found", message: "x" })),
    ).toMatch(/no longer available/i);
    expect(
      cartErrorMessage(new ApiError(404, { code: "item_not_found", message: "x" })),
    ).toMatch(/no longer in your cart/i);
  });

  it("falls back to the api message for unknown codes", () => {
    const error = new ApiError(500, { code: "internal_error", message: "boom" });
    expect(cartErrorMessage(error)).toBe("boom");
  });

  it("handles non-ApiError values", () => {
    expect(cartErrorMessage(new Error("nope"))).toMatch(/something went wrong/i);
  });
});
