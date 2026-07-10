import { apiClient } from "./client";
import type { Customer, Order } from "./types";

export function createOrder(cartId: string, customer: Customer): Promise<Order> {
  return apiClient.post<Order>("/orders", { cart_id: cartId, customer });
}

export function getOrder(orderId: string): Promise<Order> {
  return apiClient.get<Order>(`/orders/${orderId}`);
}
