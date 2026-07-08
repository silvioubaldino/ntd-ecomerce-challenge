import { useState, type ChangeEvent, type FormEvent } from "react";
import { ErrorBanner } from "../../components/ErrorBanner";
import { PageHeader } from "../../components/PageHeader";
import { Badge } from "../../components/ui/Badge";
import { Button } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { DownloadIcon, UploadIcon } from "../../components/ui/icons";
import { ApiError } from "../../api/types";
import { downloadTemplate } from "./csvTemplate";
import { fieldReasonLabel, requestErrorMessage } from "./importMessages";
import { useImportProducts } from "./hooks";

export function ProductImportPage() {
  const [file, setFile] = useState<File | null>(null);
  const [clientError, setClientError] = useState<string | null>(null);
  const importProducts = useImportProducts();

  function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const selected = event.target.files?.[0] ?? null;
    importProducts.reset();

    if (selected && !selected.name.toLowerCase().endsWith(".csv")) {
      setFile(null);
      setClientError("Please select a .csv file.");
      return;
    }

    setClientError(null);
    setFile(selected);
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (!file) return;
    importProducts.mutate(file);
  }

  const apiError =
    importProducts.error instanceof ApiError ? importProducts.error : undefined;
  const report = importProducts.data;

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Import Products"
        description="Bulk-import products from a CSV file."
        backTo={{ to: "/products", label: "Back to products" }}
        actions={
          <Button variant="secondary" onClick={downloadTemplate}>
            <DownloadIcon className="h-4 w-4" />
            Download template
          </Button>
        }
      />

      <Card>
        <form onSubmit={handleSubmit} className="flex flex-col gap-4 p-6">
          <div className="flex flex-col gap-1.5">
            <label htmlFor="csv-file" className="text-sm font-medium text-slate-700">
              CSV file
            </label>
            <input
              id="csv-file"
              type="file"
              accept=".csv,text/csv"
              onChange={handleFileChange}
              className="text-sm text-slate-600 file:mr-3 file:rounded-lg file:border-0 file:bg-brand-50 file:px-3 file:py-1.5 file:text-sm file:font-medium file:text-brand-700 hover:file:bg-brand-100"
            />
            {clientError && <p className="text-sm text-red-600">{clientError}</p>}
          </div>

          <div>
            <Button type="submit" disabled={!file || importProducts.isPending}>
              <UploadIcon className="h-4 w-4" />
              {importProducts.isPending ? "Importing…" : "Upload"}
            </Button>
          </div>
        </form>
      </Card>

      {apiError && (
        <ErrorBanner
          message={requestErrorMessage(apiError.code, apiError.message)}
        />
      )}

      {report && (
        <Card>
          <div className="flex flex-wrap items-center gap-3 border-b border-slate-900/5 px-6 py-4">
            <Badge tone="success">{report.summary.imported} imported</Badge>
            <Badge tone={report.summary.rejected > 0 ? "danger" : "neutral"}>
              {report.summary.rejected} rejected
            </Badge>
            <span className="text-sm text-slate-500">
              {report.summary.total} rows processed
            </span>
          </div>

          {report.rejected.length > 0 && (
            <table className="w-full text-sm">
              <thead className="border-b border-slate-900/5 bg-slate-50/60">
                <tr>
                  <th className="px-4 py-3 pl-6 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">
                    Row
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">
                    SKU
                  </th>
                  <th className="px-4 py-3 pr-6 text-left text-xs font-semibold uppercase tracking-wider text-slate-500">
                    Errors
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-900/5">
                {report.rejected.map((rejectedRow) => (
                  <tr key={rejectedRow.row}>
                    <td className="px-4 py-3 pl-6">{rejectedRow.row}</td>
                    <td className="px-4 py-3 font-mono text-xs text-slate-600">
                      {rejectedRow.sku || "—"}
                    </td>
                    <td className="px-4 py-3 pr-6">
                      <div className="flex flex-wrap gap-1.5">
                        {Object.entries(rejectedRow.errors).map(([field, code]) => (
                          <Badge key={field} tone="warning">
                            {field}: {fieldReasonLabel(code)}
                          </Badge>
                        ))}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </Card>
      )}
    </div>
  );
}
