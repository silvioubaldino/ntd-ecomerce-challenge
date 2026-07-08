import { CSV_TEMPLATE_HEADER, downloadTemplate } from "./csvTemplate";

describe("csvTemplate", () => {
  it("exposes the exact header expected by the api", () => {
    expect(CSV_TEMPLATE_HEADER).toBe(
      "name,sku,description,category,price,stock,weight_kg",
    );
  });

  it("triggers a client-side download with no network request", () => {
    const clickSpy = vi
      .spyOn(HTMLAnchorElement.prototype, "click")
      .mockImplementation(() => {});
    URL.createObjectURL = vi.fn(() => "blob:mock-url");
    URL.revokeObjectURL = vi.fn();

    downloadTemplate();

    expect(clickSpy).toHaveBeenCalledTimes(1);
    expect(URL.createObjectURL).toHaveBeenCalledTimes(1);
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:mock-url");

    clickSpy.mockRestore();
  });
});
