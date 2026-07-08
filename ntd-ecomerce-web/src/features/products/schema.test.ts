import { productInputSchema } from "./schema";

const valid = {
  name: "Widget",
  sku: "WID-001",
  description: "",
  category: "Tools",
  price: "19.90",
  stock: 5,
  weight_kg: "0.50",
};

describe("productInputSchema", () => {
  it("accepts a valid ProductInput", () => {
    expect(productInputSchema.safeParse(valid).success).toBe(true);
  });

  it.each(["name", "sku", "category"])("rejects a missing %s", (field) => {
    const result = productInputSchema.safeParse({ ...valid, [field]: "" });
    expect(result.success).toBe(false);
  });

  it("rejects a name over 255 characters", () => {
    const result = productInputSchema.safeParse({ ...valid, name: "a".repeat(256) });
    expect(result.success).toBe(false);
  });

  it.each(["-1", "-0.5", "abc", ""])("rejects a non-numeric or negative price %s", (price) => {
    const result = productInputSchema.safeParse({ ...valid, price });
    expect(result.success).toBe(false);
  });

  it.each(["0", "19.9", "100.00"])("accepts a valid decimal string price %s", (price) => {
    const result = productInputSchema.safeParse({ ...valid, price });
    expect(result.success).toBe(true);
  });

  it("rejects a negative stock", () => {
    const result = productInputSchema.safeParse({ ...valid, stock: -1 });
    expect(result.success).toBe(false);
  });

  it("rejects a non-integer stock", () => {
    const result = productInputSchema.safeParse({ ...valid, stock: 1.5 });
    expect(result.success).toBe(false);
  });
});
