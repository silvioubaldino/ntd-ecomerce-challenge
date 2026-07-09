import { apiClient } from "./client";
import type { Cart } from "./types";

export function createCart(): Promise<Cart> {
  return apiClient.post<Cart>("/carts", {});
}

export function getCart(cartId: string): Promise<Cart> {
  return apiClient.get<Cart>(`/carts/${cartId}`);
}

export function addCartItem(
  cartId: string,
  productId: string,
  quantity: number,
): Promise<Cart> {
  return apiClient.post<Cart>(`/carts/${cartId}/items`, {
    product_id: productId,
    quantity,
  });
}

export function updateCartItem(
  cartId: string,
  productId: string,
  quantity: number,
): Promise<Cart> {
  return apiClient.put<Cart>(`/carts/${cartId}/items/${productId}`, { quantity });
}

export function removeCartItem(cartId: string, productId: string): Promise<Cart> {
  return apiClient.delete<Cart>(`/carts/${cartId}/items/${productId}`);
}
