import { z } from "zod";

const decimalString = z
  .string()
  .min(1, "required")
  .regex(/^\d+(\.\d+)?$/, "must be a non-negative decimal")
  .refine((value) => !Number.isNaN(Number(value)), "must be a non-negative decimal");

export const productInputSchema = z.object({
  name: z.string().min(1, "required").max(255, "too long"),
  sku: z.string().min(1, "required").max(64, "too long"),
  description: z.string().max(1000, "too long").default(""),
  category: z.string().min(1, "required").max(100, "too long"),
  price: decimalString,
  stock: z.coerce
    .number()
    .int("must be a non-negative integer")
    .min(0, "must be a non-negative integer"),
  weight_kg: decimalString,
});

export type ProductFormValues = z.infer<typeof productInputSchema>;
