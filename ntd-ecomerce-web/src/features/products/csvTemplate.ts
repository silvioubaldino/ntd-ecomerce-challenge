export const CSV_TEMPLATE_HEADER =
  "name,sku,description,category,price,stock,weight_kg";

export function downloadTemplate() {
  const blob = new Blob([`${CSV_TEMPLATE_HEADER}\n`], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "products-template.csv";
  link.click();
  URL.revokeObjectURL(url);
}
