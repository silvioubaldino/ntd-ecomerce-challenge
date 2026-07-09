// Guest Cart identity (AYD-004): the server-generated cart_id is persisted locally and
// sent on every cart call. A tiny external store lets React components (nav badge, pages)
// re-render when the cart_id is created or reset.

const CART_ID_KEY = "ntd.cart_id";
const listeners = new Set<() => void>();

function notify() {
  for (const listener of listeners) listener();
}

export function getStoredCartId(): string | null {
  return localStorage.getItem(CART_ID_KEY);
}

export function setStoredCartId(cartId: string): void {
  localStorage.setItem(CART_ID_KEY, cartId);
  notify();
}

export function clearStoredCartId(): void {
  localStorage.removeItem(CART_ID_KEY);
  notify();
}

export function subscribeCartId(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}
