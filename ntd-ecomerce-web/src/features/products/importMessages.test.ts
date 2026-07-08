import { fieldReasonLabel, requestErrorMessage } from "./importMessages";

describe("fieldReasonLabel", () => {
  it.each([
    ["required", "Required"],
    ["malformed_sku", "Malformed SKU"],
    ["duplicate_sku", "Duplicate SKU"],
    ["must_be_non_negative_decimal", "Must be a non-negative decimal"],
    ["must_be_non_negative_integer", "Must be a non-negative integer"],
    ["unsafe_content", "Unsafe content"],
  ])("maps %s to a readable label", (code, label) => {
    expect(fieldReasonLabel(code)).toBe(label);
  });

  it("falls back to the raw code for an unknown reason", () => {
    expect(fieldReasonLabel("some_new_code")).toBe("some_new_code");
  });
});

describe("requestErrorMessage", () => {
  it.each([
    ["invalid_header", /header/i],
    ["invalid_file", /file/i],
    ["file_too_large", /size/i],
  ])("maps %s to a readable message", (code, pattern) => {
    expect(requestErrorMessage(code, "fallback")).toMatch(pattern);
  });

  it("falls back to the given message for an unknown code", () => {
    expect(requestErrorMessage("internal_error", "Unexpected server error.")).toBe(
      "Unexpected server error.",
    );
  });
});
