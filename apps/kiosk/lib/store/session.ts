"use client";
import { create } from "zustand";

interface SessionState {
  tableCode: string;
  tenantId: string;
  outletId: string;
  authToken: string | null;
  setTable: (code: string) => void;
}

export const useSessionStore = create<SessionState>((set) => ({
  tableCode: "A12",
  tenantId: "demo-tenant",
  outletId: "demo-outlet",
  authToken: "kiosk-service-jwt-demo",
  setTable: (code) => set({ tableCode: code }),
}));
