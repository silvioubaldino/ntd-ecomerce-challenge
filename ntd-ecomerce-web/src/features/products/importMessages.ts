const fieldReasonLabels: Record<string, string> = {
  required: "Required",
  malformed_sku: "Malformed SKU",
  duplicate_sku: "Duplicate SKU",
  must_be_non_negative_decimal: "Must be a non-negative decimal",
  must_be_non_negative_integer: "Must be a non-negative integer",
  unsafe_content: "Unsafe content",
};

export function fieldReasonLabel(code: string): string {
  return fieldReasonLabels[code] ?? code;
}

const requestErrorLabels: Record<string, string> = {
  invalid_header: "The CSV header is missing or doesn't match the expected columns.",
  invalid_file: "The file is missing, empty, or not a valid CSV.",
  file_too_large: "The file exceeds the maximum allowed size.",
};

export function requestErrorMessage(code: string, fallback: string): string {
  return requestErrorLabels[code] ?? fallback;
}
