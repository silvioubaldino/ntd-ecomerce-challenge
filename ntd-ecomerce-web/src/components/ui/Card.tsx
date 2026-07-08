import type { HTMLAttributes } from "react";
import { cn } from "../../lib/cn";

export function Card({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn(
        "overflow-hidden rounded-2xl bg-white shadow-card ring-1 ring-slate-900/5",
        className,
      )}
      {...props}
    />
  );
}
