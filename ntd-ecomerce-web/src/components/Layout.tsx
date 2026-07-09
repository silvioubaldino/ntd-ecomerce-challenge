import { NavLink, Link, Outlet } from "react-router-dom";
import { cn } from "../lib/cn";
import { useCart } from "../features/cart/hooks";
import { CartIcon, LogoMark, ShieldIcon } from "./ui/icons";

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    "rounded-lg px-3 py-2 text-sm font-medium transition",
    isActive
      ? "bg-brand-50 text-brand-800"
      : "text-slate-600 hover:bg-slate-100 hover:text-slate-900",
  );

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
            <NavLink to="/products" end className={navLinkClass}>
              Catalog
            </NavLink>
            <NavLink to="/store" className={navLinkClass}>
              Store
            </NavLink>
            <CartNavLink />
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

function CartNavLink() {
  const cartQuery = useCart();
  const count =
    cartQuery.data?.items.reduce((sum, item) => sum + item.quantity, 0) ?? 0;

  return (
    <NavLink to="/cart" className={navLinkClass} aria-label="Cart">
      <span className="inline-flex items-center gap-1.5">
        <CartIcon className="h-4 w-4" />
        Cart
        {count > 0 && (
          <span
            aria-label={`${count} items in cart`}
            className="inline-flex min-w-[1.25rem] items-center justify-center rounded-full bg-brand-700 px-1.5 py-0.5 text-xs font-semibold leading-none text-white"
          >
            {count}
          </span>
        )}
      </span>
    </NavLink>
  );
}
