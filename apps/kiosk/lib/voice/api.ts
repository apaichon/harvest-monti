"use client";
// REST client stub against TASK-0009 endpoints.
// Kiosk-side renders are powered by fixtures; these helpers exist so wiring is
// already in place when the backend lands.

import type { Category, Item, Promo } from "@monti/shared";
import { categories as fxCategories, items as fxItems, promos as fxPromos } from "../fixtures/menu";

const BASE = process.env.NEXT_PUBLIC_MONTI_API ?? "/api/v1";

async function safe<T>(path: string, fallback: T, init?: RequestInit): Promise<T> {
  try {
    const res = await fetch(`${BASE}${path}`, init);
    if (!res.ok) return fallback;
    return (await res.json()) as T;
  } catch {
    return fallback;
  }
}

export const api = {
  async getCategories(): Promise<Category[]> {
    return safe("/public/menu/categories", fxCategories);
  },
  async getMenuItems(): Promise<Item[]> {
    return safe("/public/menu/items", fxItems);
  },
  async getActivePromos(outletId: string): Promise<Promo[]> {
    return safe(`/public/outlets/${outletId}/promos/active`, fxPromos);
  },
};

export async function postStaffCall(payload: {
  tenant_id: string;
  outlet_id: string;
  table_code: string;
  reason: string;
  requested_at: string;
}) {
  try {
    await fetch(`${BASE}/public/staff/call`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(payload),
    });
  } catch {
    // ignored — sinks may be unavailable in dev
  }
}
