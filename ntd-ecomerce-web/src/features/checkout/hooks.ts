import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createOrder, getOrder } from "../../api/orders";
import { ApiError, type Customer, type Order } from "../../api/types";
import { clearStoredCartId, getStoredCartId } from "../cart/cartStorage";

const orderKeys = {
  detail: (orderId: string) => ["order", orderId] as const,
};

/**
 * Turns the current guest Cart into a confirmed Order (POST /orders). On success the Order
 * is seeded into the cache, the stored cart_id is cleared and the cart query is removed so
 * the nav badge resets — the cart is consumed server-side. A stale cart_not_found also
 * resets the stored cart.
 */
export function useCheckout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (customer: Customer): Promise<Order> => {
      const cartId = getStoredCartId();
      if (cartId === null) {
        return Promise.reject(
          new ApiError(404, { code: "cart_not_found", message: "No cart." }),
        );
      }
      return createOrder(cartId, customer);
    },
    onSuccess: (order) => {
      queryClient.setQueryData(orderKeys.detail(order.id), order);
      clearStoredCartId();
      queryClient.removeQueries({ queryKey: ["cart"] });
    },
    onError: (error) => {
      if (error instanceof ApiError && error.code === "cart_not_found") {
        clearStoredCartId();
      }
    },
  });
}

/** Loads a single confirmed Order; disabled until an orderId is present. */
export function useOrder(orderId: string | undefined) {
  return useQuery({
    queryKey: orderKeys.detail(orderId ?? ""),
    queryFn: () => getOrder(orderId as string),
    enabled: Boolean(orderId),
  });
}
