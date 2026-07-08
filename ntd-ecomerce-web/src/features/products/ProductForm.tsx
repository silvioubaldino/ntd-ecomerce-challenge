import { useEffect, useId } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ApiError } from "../../api/types";
import { Button, ButtonLink } from "../../components/ui/Button";
import { Card } from "../../components/ui/Card";
import { cn } from "../../lib/cn";
import { productInputSchema, type ProductFormValues } from "./schema";

interface ProductFormProps {
  defaultValues?: ProductFormValues;
  onSubmit: (values: ProductFormValues) => Promise<unknown>;
  submitLabel: string;
}

const emptyValues: ProductFormValues = {
  name: "",
  sku: "",
  description: "",
  category: "",
  price: "0",
  stock: 0,
  weight_kg: "0",
};

export function ProductForm({ defaultValues, onSubmit, submitLabel }: ProductFormProps) {
  const priceId = useId();
  const {
    register,
    handleSubmit,
    setError,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<ProductFormValues>({
    resolver: zodResolver(productInputSchema),
    defaultValues: defaultValues ?? emptyValues,
  });

  useEffect(() => {
    if (defaultValues) reset(defaultValues);
  }, [defaultValues, reset]);

  async function submit(values: ProductFormValues) {
    try {
      await onSubmit(values);
    } catch (error) {
      if (error instanceof ApiError && error.code === "sku_already_exists") {
        setError("sku", { message: "SKU already exists" });
        return;
      }
      if (error instanceof ApiError && error.code === "validation_error" && error.details) {
        for (const [field, problem] of Object.entries(error.details)) {
          setError(field as keyof ProductFormValues, {
            message: problem.replace(/_/g, " "),
          });
        }
        return;
      }
      throw error;
    }
  }

  return (
    <form onSubmit={handleSubmit(submit)} className="flex max-w-2xl flex-col gap-6">
      <Card className="p-6 sm:p-8">
        <div className="flex flex-col gap-6">
          <SectionTitle
            title="Details"
            description="How the product appears in the catalog."
          />

          <div className="grid gap-5 sm:grid-cols-2">
            <Field label="Name" error={errors.name?.message}>
              <input
                className={cn("input", errors.name && "input-invalid")}
                placeholder="e.g. Wireless Headphones"
                {...register("name")}
              />
            </Field>
            <Field label="SKU" error={errors.sku?.message}>
              <input
                className={cn("input font-mono", errors.sku && "input-invalid")}
                placeholder="e.g. WH-1000"
                {...register("sku")}
              />
            </Field>
          </div>

          <Field label="Description" error={errors.description?.message}>
            <textarea
              rows={3}
              className={cn("input resize-y", errors.description && "input-invalid")}
              placeholder="A short description customers will see."
              {...register("description")}
            />
          </Field>

          <Field label="Category" error={errors.category?.message}>
            <input
              className={cn("input", errors.category && "input-invalid")}
              placeholder="e.g. Electronics"
              {...register("category")}
            />
          </Field>

          <hr className="border-slate-900/5" />

          <SectionTitle
            title="Pricing & inventory"
            description="Price, available stock and shipping weight."
          />

          <div className="grid gap-5 sm:grid-cols-3">
            <Field label="Price" htmlFor={priceId} error={errors.price?.message}>
              <div className="relative">
                <span
                  aria-hidden
                  className="pointer-events-none absolute inset-y-0 left-3.5 flex items-center text-sm text-slate-400"
                >
                  $
                </span>
                <input
                  id={priceId}
                  inputMode="decimal"
                  className={cn("input pl-8", errors.price && "input-invalid")}
                  {...register("price")}
                />
              </div>
            </Field>
            <Field label="Stock" error={errors.stock?.message}>
              <input
                type="number"
                className={cn("input", errors.stock && "input-invalid")}
                {...register("stock")}
              />
            </Field>
            <Field label="Weight (kg)" error={errors.weight_kg?.message}>
              <input
                inputMode="decimal"
                className={cn("input", errors.weight_kg && "input-invalid")}
                {...register("weight_kg")}
              />
            </Field>
          </div>
        </div>
      </Card>

      <div className="flex items-center justify-end gap-3">
        <ButtonLink to="/products" variant="secondary">
          Cancel
        </ButtonLink>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Saving…" : submitLabel}
        </Button>
      </div>
    </form>
  );
}

function SectionTitle({ title, description }: { title: string; description: string }) {
  return (
    <div className="flex flex-col gap-0.5">
      <h2 className="text-sm font-semibold text-slate-900">{title}</h2>
      <p className="text-xs text-slate-500">{description}</p>
    </div>
  );
}

function Field({
  label,
  error,
  htmlFor,
  children,
}: {
  label: string;
  error?: string;
  htmlFor?: string;
  children: React.ReactNode;
}) {
  const fieldClasses = "flex flex-col gap-1.5 text-sm font-medium text-slate-700";
  const errorEl = error && (
    <span className="text-xs font-normal text-red-600">{error}</span>
  );

  // When the control has visual adornments (e.g. the "$" prefix), the label can't
  // wrap it — the adornment text would leak into the accessible label name.
  if (htmlFor) {
    return (
      <div className={fieldClasses}>
        <label htmlFor={htmlFor}>{label}</label>
        {children}
        {errorEl}
      </div>
    );
  }

  return (
    <label className={fieldClasses}>
      {label}
      {children}
      {errorEl}
    </label>
  );
}
