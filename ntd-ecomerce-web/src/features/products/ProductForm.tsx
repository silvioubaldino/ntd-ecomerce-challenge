import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ApiError } from "../../api/types";
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
    <form onSubmit={handleSubmit(submit)} className="flex max-w-md flex-col gap-4">
      <Field label="Name" error={errors.name?.message}>
        <input className="input" {...register("name")} />
      </Field>
      <Field label="SKU" error={errors.sku?.message}>
        <input className="input" {...register("sku")} />
      </Field>
      <Field label="Description" error={errors.description?.message}>
        <textarea className="input" {...register("description")} />
      </Field>
      <Field label="Category" error={errors.category?.message}>
        <input className="input" {...register("category")} />
      </Field>
      <Field label="Price" error={errors.price?.message}>
        <input className="input" {...register("price")} />
      </Field>
      <Field label="Stock" error={errors.stock?.message}>
        <input type="number" className="input" {...register("stock")} />
      </Field>
      <Field label="Weight (kg)" error={errors.weight_kg?.message}>
        <input className="input" {...register("weight_kg")} />
      </Field>
      <button
        type="submit"
        disabled={isSubmitting}
        className="rounded bg-blue-600 px-4 py-2 font-medium text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {submitLabel}
      </button>
    </form>
  );
}

function Field({
  label,
  error,
  children,
}: {
  label: string;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <label className="flex flex-col gap-1 text-sm font-medium text-gray-700">
      {label}
      {children}
      {error && <span className="text-xs font-normal text-red-600">{error}</span>}
    </label>
  );
}
