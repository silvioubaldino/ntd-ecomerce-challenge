import { http, HttpResponse } from "msw";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import App from "../../App";
import { server } from "../../test/server";
import { renderWithProviders } from "../../test/utils";

function csvFile(name = "products.csv") {
  return new File(["name,sku\nWidget,WID-001"], name, { type: "text/csv" });
}

describe("ProductImportPage", () => {
  it("is reachable from the catalog via the Import CSV action", async () => {
    renderWithProviders(<App />, { route: "/products" });

    await userEvent.click(screen.getByRole("link", { name: /import csv/i }));

    expect(
      await screen.findByRole("heading", { name: "Import Products" }),
    ).toBeInTheDocument();
    expect(screen.getByLabelText("CSV file")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /download template/i })).toBeInTheDocument();
  });

  it("downloads the blank template without calling the api", async () => {
    let importCalled = false;
    server.use(
      http.post("/api/products/import", () => {
        importCalled = true;
        return HttpResponse.json({ summary: { total: 0, imported: 0, rejected: 0 }, rejected: [] });
      }),
    );
    const clickSpy = vi
      .spyOn(HTMLAnchorElement.prototype, "click")
      .mockImplementation(() => {});
    URL.createObjectURL = vi.fn(() => "blob:mock-url");
    URL.revokeObjectURL = vi.fn();

    renderWithProviders(<App />, { route: "/products/import" });
    await userEvent.click(screen.getByRole("button", { name: /download template/i }));

    expect(clickSpy).toHaveBeenCalledTimes(1);
    expect(importCalled).toBe(false);

    clickSpy.mockRestore();
  });

  it("disables the upload button until a file is selected", () => {
    renderWithProviders(<App />, { route: "/products/import" });

    expect(screen.getByRole("button", { name: "Upload" })).toBeDisabled();
  });

  it("rejects a non-CSV file client-side without uploading", async () => {
    let importCalled = false;
    server.use(
      http.post("/api/products/import", () => {
        importCalled = true;
        return HttpResponse.json({ summary: { total: 0, imported: 0, rejected: 0 }, rejected: [] });
      }),
    );

    renderWithProviders(<App />, { route: "/products/import" });
    // applyAccept: false — bypass user-event's own `accept` filtering so the
    // file reaches our onChange handler, where the real client-side check lives.
    const user = userEvent.setup({ applyAccept: false });
    await user.upload(
      screen.getByLabelText("CSV file"),
      new File(["not a csv"], "products.txt", { type: "text/plain" }),
    );

    expect(await screen.findByText("Please select a .csv file.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Upload" })).toBeDisabled();
    expect(importCalled).toBe(false);
  });

  it("shows the import summary for a fully valid CSV", async () => {
    server.use(
      http.post("/api/products/import", () =>
        HttpResponse.json({ summary: { total: 3, imported: 3, rejected: 0 }, rejected: [] }),
      ),
    );

    renderWithProviders(<App />, { route: "/products/import" });
    await userEvent.upload(screen.getByLabelText("CSV file"), csvFile());
    await userEvent.click(screen.getByRole("button", { name: "Upload" }));

    expect(await screen.findByText("3 imported")).toBeInTheDocument();
    expect(screen.getByText("0 rejected")).toBeInTheDocument();
    expect(screen.queryByText(/errors/i)).not.toBeInTheDocument();
  });

  it("lists rejected rows with readable field reasons for a partial import", async () => {
    server.use(
      http.post("/api/products/import", () =>
        HttpResponse.json({
          summary: { total: 2, imported: 1, rejected: 1 },
          rejected: [
            {
              row: 2,
              sku: "BS-021",
              errors: {
                price: "must_be_non_negative_decimal",
                sku: "duplicate_sku",
              },
            },
          ],
        }),
      ),
    );

    renderWithProviders(<App />, { route: "/products/import" });
    await userEvent.upload(screen.getByLabelText("CSV file"), csvFile());
    await userEvent.click(screen.getByRole("button", { name: "Upload" }));

    expect(await screen.findByText("1 imported")).toBeInTheDocument();
    expect(screen.getByText("1 rejected")).toBeInTheDocument();
    expect(screen.getByText("BS-021")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
    expect(screen.getByText("price: Must be a non-negative decimal")).toBeInTheDocument();
    expect(screen.getByText("sku: Duplicate SKU")).toBeInTheDocument();
  });

  it("surfaces a wrong header as a readable error and shows no report", async () => {
    server.use(
      http.post("/api/products/import", () =>
        HttpResponse.json(
          { error: { code: "invalid_header", message: "header mismatch" } },
          { status: 422 },
        ),
      ),
    );

    renderWithProviders(<App />, { route: "/products/import" });
    await userEvent.upload(screen.getByLabelText("CSV file"), csvFile());
    await userEvent.click(screen.getByRole("button", { name: "Upload" }));

    expect(
      await screen.findByText(/csv header is missing or doesn't match/i),
    ).toBeInTheDocument();
    expect(screen.queryByText(/imported$/)).not.toBeInTheDocument();
  });

  it("surfaces an invalid/missing file error (400)", async () => {
    server.use(
      http.post("/api/products/import", () =>
        HttpResponse.json(
          { error: { code: "invalid_file", message: "invalid file" } },
          { status: 400 },
        ),
      ),
    );

    renderWithProviders(<App />, { route: "/products/import" });
    await userEvent.upload(screen.getByLabelText("CSV file"), csvFile());
    await userEvent.click(screen.getByRole("button", { name: "Upload" }));

    expect(
      await screen.findByText(/not a valid csv/i),
    ).toBeInTheDocument();
  });

  it("surfaces a file-too-large error (413)", async () => {
    server.use(
      http.post("/api/products/import", () =>
        HttpResponse.json(
          { error: { code: "file_too_large", message: "file too large" } },
          { status: 413 },
        ),
      ),
    );

    renderWithProviders(<App />, { route: "/products/import" });
    await userEvent.upload(screen.getByLabelText("CSV file"), csvFile());
    await userEvent.click(screen.getByRole("button", { name: "Upload" }));

    expect(
      await screen.findByText(/exceeds the maximum allowed size/i),
    ).toBeInTheDocument();
  });

  it("shows a generic error state on a server/network failure", async () => {
    server.use(
      http.post("/api/products/import", () =>
        HttpResponse.json(
          { error: { code: "internal_error", message: "Unexpected server error." } },
          { status: 500 },
        ),
      ),
    );

    renderWithProviders(<App />, { route: "/products/import" });
    await userEvent.upload(screen.getByLabelText("CSV file"), csvFile());
    await userEvent.click(screen.getByRole("button", { name: "Upload" }));

    expect(await screen.findByText("Unexpected server error.")).toBeInTheDocument();
  });

  it("invalidates the products list after a successful import", async () => {
    let listCalls = 0;
    server.use(
      http.get("/api/products", () => {
        listCalls += 1;
        return HttpResponse.json({
          data: [],
          pagination: { page: 1, page_size: 20, total: 0 },
        });
      }),
      http.post("/api/products/import", () =>
        HttpResponse.json({ summary: { total: 1, imported: 1, rejected: 0 }, rejected: [] }),
      ),
    );

    renderWithProviders(<App />, { route: "/products" });
    await screen.findByRole("heading", { name: "Products" });
    const callsAfterInitialLoad = listCalls;

    await userEvent.click(screen.getByRole("link", { name: /import csv/i }));
    await userEvent.upload(screen.getByLabelText("CSV file"), csvFile());
    await userEvent.click(screen.getByRole("button", { name: "Upload" }));
    await screen.findByText("1 imported");

    await userEvent.click(screen.getByRole("link", { name: /back to products/i }));
    await screen.findByRole("heading", { name: "Products" });

    expect(listCalls).toBeGreaterThan(callsAfterInitialLoad);
  });
});
