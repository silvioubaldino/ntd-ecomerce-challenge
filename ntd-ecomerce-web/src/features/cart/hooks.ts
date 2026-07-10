import { useEffect, useSyncExternalStore } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  addCartItem,
  createCart,
  getCart,
  removeCartItem,
  updateCartItem,
} from "../../api/cart";
import { ApiError, type Cart } from "../../api/types";
import {
  clearStoredCartId,
  getStoredCartId,
  setStoredCartId,
  subscribeCartId,
} from "./cartStorage";

const cartKeys = {
  detail: (cartId: string | null) => ["cart", cartId] as const,
};

export function useCartId(): string | null {
  return useSyncExternalStore(subscribeCartId, getStoredCartId, () => null);
}

export function useCart() {
  const cartId = useCartId();
  const query = useQuery({
    queryKey: cartKeys.detail(cartId),
    queryFn: () => getCart(cartId as string),
    enabled: cartId !== null,
  });

  useEffect(() => {
    if (query.error instanceof ApiError && query.error.code === "cart_not_found") {
      clearStoredCartId();
    }
  }, [query.error]);

  return query;
}

export function useAddToCart() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (vars: { productId: string; quantity?: number }) => {
      const quantity = vars.quantity ?? 1;
      let cartId = getStoredCartId();
      if (cartId === null) {
        const created = await createCart();
        setStoredCartId(created.id);
        cartId = created.id;
      }
      try {
        return await addCartItem(cartId, vars.productId, quantity);
      } catch (error) {
        if (error instanceof ApiError && error.code === "cart_not_found") {
          const created = await createCart();
          setStoredCartId(created.id);
          return await addCartItem(created.id, vars.productId, quantity);
        }
        throw error;
      }
    },
    onSuccess: (cart) => cacheCart(queryClient, cart),
  });
}

export function useUpdateCartItem() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (vars: { productId: string; quantity: number }) =>
      withCart((cartId) => updateCartItem(cartId, vars.productId, vars.quantity)),
    onSuccess: (cart) => cacheCart(queryClient, cart),
    onError: resetOnMissingCart,
  });
}

export function useRemoveCartItem() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (vars: { productId: string }) =>
      withCart((cartId) => removeCartItem(cartId, vars.productId)),
    onSuccess: (cart) => cacheCart(queryClient, cart),
    onError: resetOnMissingCart,
  });
}

function withCart(fn: (cartId: string) => Promise<Cart>): Promise<Cart> {
  const cartId = getStoredCartId();
  if (cartId === null) {
    return Promise.reject(
      new ApiError(404, { code: "cart_not_found", message: "No cart." }),
    );
  }
  return fn(cartId);
}

function cacheCart(
  queryClient: ReturnType<typeof useQueryClient>,
  cart: Cart,
): void {
  queryClient.setQueryData(cartKeys.detail(cart.id), cart);
}

function resetOnMissingCart(error: unknown): void {
  if (error instanceof ApiError && error.code === "cart_not_found") {
    clearStoredCartId();
  }
}
