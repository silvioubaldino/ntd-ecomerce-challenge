import { NavLink, Link, Outlet } from "react-router-dom";
import { cn } from "../lib/cn";
import { LogoMark, ShieldIcon } from "./ui/icons";

export function Layout() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="sticky top-0 z-40 border-b border-slate-900/5 bg-white/80 backdrop-blur">
        <nav className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between gap-6 px-4 sm:px-6">
          <Link to="/products" className="flex items-center gap-2.5">
            <LogoMark className="h-8 w-8 text-brand-700" />
            <span className="text-[15px] font-semibold tracking-tight text-slate-900">
              NTD Market
            </span>
          </Link>

          <div className="flex items-center gap-1">
            <NavLink
              to="/products"
              end
              className={({ isActive }) =>
                cn(
                  "rounded-lg px-3 py-2 text-sm font-medium transition",
                  isActive
                    ? "bg-brand-50 text-brand-800"
                    : "text-slate-600 hover:bg-slate-100 hover:text-slate-900",
                )
              }
            >
              Catalog
            </NavLink>
            <NavLink
              to="/store"
              className={({ isActive }) =>
                cn(
                  "rounded-lg px-3 py-2 text-sm font-medium transition",
                  isActive
                    ? "bg-brand-50 text-brand-800"
                    : "text-slate-600 hover:bg-slate-100 hover:text-slate-900",
                )
              }
            >
              Store
            </NavLink>
          </div>
        </nav>
      </header>

      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8 sm:px-6 sm:py-10">
        <Outlet />
      </main>

      <footer className="border-t border-slate-900/5 bg-white">
        <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center justify-between gap-3 px-4 py-5 text-xs text-slate-500 sm:px-6">
          <p>© 2026 NTD Market. All rights reserved.</p>
          <p className="inline-flex items-center gap-1.5">
            <ShieldIcon className="h-3.5 w-3.5 text-emerald-600" />
            Secure catalog management
          </p>
        </div>
      </footer>
    </div>
  );
}
