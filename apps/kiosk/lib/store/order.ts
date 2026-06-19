"use client";
import { create } from "zustand";
import type { OrderStatusStep } from "@monti/shared";

export interface OrderState {
  orderId: string | null;
  status: OrderStatusStep | null;
  eta: string | null;
  receivedAt: string | null;
  setSubmitted: (id: string, eta?: string) => void;
  setStatus: (status: OrderStatusStep) => void;
  reset: () => void;
}

export const useOrderStore = create<OrderState>((set) => ({
  orderId: null,
  status: null,
  eta: null,
  receivedAt: null,
  setSubmitted: (id, eta) => set({ orderId: id, status: "received", eta: eta ?? "15-20 min", receivedAt: new Date().toISOString() }),
  setStatus: (status) => set({ status }),
  reset: () => set({ orderId: null, status: null, eta: null, receivedAt: null }),
}));
