import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { ChevronLeftIcon } from "./ui/icons";

interface PageHeaderProps {
  title: string;
  description?: string;
  backTo?: { to: string; label: string };
  actions?: ReactNode;
}

export function PageHeader({ title, description, backTo, actions }: PageHeaderProps) {
  return (
    <div className="flex flex-wrap items-end justify-between gap-4">
      <div className="flex flex-col gap-1">
        {backTo && (
          <Link
            to={backTo.to}
            className="mb-1 inline-flex items-center gap-1 text-sm font-medium text-slate-500 transition hover:text-brand-700"
          >
            <ChevronLeftIcon className="h-4 w-4" />
            {backTo.label}
          </Link>
        )}
        <h1 className="text-2xl font-semibold tracking-tight text-slate-900">{title}</h1>
        {description && <p className="text-sm text-slate-500">{description}</p>}
      </div>
      {actions && <div className="flex items-center gap-3">{actions}</div>}
    </div>
  );
}
