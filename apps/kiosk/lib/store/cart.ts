"use client";
import { create } from "zustand";
import type { CartLine, CartTotals, Item } from "@monti/shared";

export type CartMachineState = "empty" | "adding" | "checking_out" | "submitted" | "abandoned";

export interface AddLineInput {
  item: Item;
  qty: number;
  modifiers?: Record<string, string | string[]>;
  unitPrice?: number;
  modifiersLabel?: string;
}

interface CartState {
  state: CartMachineState;
  lines: CartLine[];
  tableCode: string;
  promoCode: string | null;
  promoError: string | null;
  promoDiscount: number;
  flashLineId: string | null;
  totalAnimating: boolean;

  setTableCode: (code: string) => void;
  addLine: (input: AddLineInput) => string;
  updateQty: (lineId: string, qty: number) => void;
  removeLine: (lineId: string) => void;
  applyPromo: (code: string) => void;
  clearPromo: () => void;
  checkout: () => boolean;
  reset: () => void;
}

const SERVICE_RATE = 0.1;
const TAX_RATE = 0.07;

function modifiersKey(mods: Record<string, string | string[]>): string {
  return JSON.stringify(
    Object.entries(mods).sort(([a], [b]) => a.localeCompare(b)),
  );
}

function computeUnitPrice(item: Item, mods: Record<string, string | string[]>): number {
  let p = item.price;
  for (const g of item.modifierGroups ?? []) {
    const v = mods[g.id];
    if (g.kind === "single" && typeof v === "string" && v !== "__none__") {
      p += g.options.find((o) => o.id === v)?.priceDelta ?? 0;
    } else if (g.kind === "multi" && Array.isArray(v)) {
      for (const optId of v) p += g.options.find((o) => o.id === optId)?.priceDelta ?? 0;
    }
  }
  return p;
}

function modsLabel(item: Item, mods: Record<string, string | string[]>): string {
  const parts: string[] = [];
  for (const g of item.modifierGroups ?? []) {
    const v = mods[g.id];
    if (g.kind === "single" && typeof v === "string" && v !== "__none__") {
      const o = g.options.find((x) => x.id === v);
      if (o) parts.push(o.name);
    } else if (g.kind === "multi" && Array.isArray(v) && v.length > 0) {
      for (const optId of v) {
        const o = g.options.find((x) => x.id === optId);
        if (o) parts.push(o.name);
      }
    }
  }
  return parts.join(" + ");
}

function defaultModifiers(item: Item): Record<string, string | string[]> {
  const out: Record<string, string | string[]> = {};
  for (const g of item.modifierGroups ?? []) {
    if (g.kind === "single") {
      out[g.id] = g.required ? g.options[0]?.id ?? "__none__" : "__none__";
    } else {
      out[g.id] = [];
    }
  }
  return out;
}

export const useCartStore = create<CartState>((set, get) => ({
  state: "empty",
  lines: [],
  tableCode: "A12",
  promoCode: null,
  promoError: null,
  promoDiscount: 0,
  flashLineId: null,
  totalAnimating: false,

  setTableCode: (code) => set({ tableCode: code }),

  addLine: ({ item, qty, modifiers, unitPrice, modifiersLabel }) => {
    const mods = modifiers ?? defaultModifiers(item);
    const price = unitPrice ?? computeUnitPrice(item, mods);
    const label = modifiersLabel ?? modsLabel(item, mods);
    const key = modifiersKey(mods);

    let createdId = "";
    set((s) => {
      const existingIdx = s.lines.findIndex(
        (l) => l.itemId === item.id && modifiersKey(l.modifiers) === key,
      );
      let lines = s.lines;
      if (existingIdx >= 0) {
        lines = s.lines.map((l, i) => (i === existingIdx ? { ...l, qty: l.qty + qty } : l));
        createdId = lines[existingIdx].id;
      } else {
        const id = `line_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 7)}`;
        createdId = id;
        const newLine: CartLine = {
          id,
          itemId: item.id,
          name: item.name,
          imageUrl: item.imageUrl,
          unitPrice: price,
          qty,
          modifiers: mods,
          modifiersLabel: label,
        };
        lines = [...s.lines, newLine];
      }
      return { lines, state: "adding", flashLineId: createdId, totalAnimating: true };
    });
    setTimeout(() => {
      const st = get();
      if (st.flashLineId === createdId) set({ flashLineId: null });
      set({ totalAnimating: false });
    }, 600);
    return createdId;
  },

  updateQty: (lineId, qty) =>
    set((s) => {
      if (qty <= 0) {
        const lines = s.lines.filter((l) => l.id !== lineId);
        const newState = lines.length === 0 ? "empty" : "adding";
        return {
          lines,
          state: newState,
          promoCode: newState === "empty" ? null : s.promoCode,
          promoDiscount: newState === "empty" ? 0 : s.promoDiscount,
        };
      }
      return {
        lines: s.lines.map((l) => (l.id === lineId ? { ...l, qty } : l)),
      };
    }),

  removeLine: (lineId) =>
    set((s) => {
      const lines = s.lines.filter((l) => l.id !== lineId);
      const newState = lines.length === 0 ? "empty" : "adding";
      return {
        lines,
        state: newState,
        promoCode: newState === "empty" ? null : s.promoCode,
        promoDiscount: newState === "empty" ? 0 : s.promoDiscount,
      };
    }),

  applyPromo: (code) => {
    const normalized = code.trim().toUpperCase();
    if (normalized === "LUNCH20") {
      const subtotal = selectSubtotal(get());
      set({
        promoCode: normalized,
        promoError: null,
        promoDiscount: Math.round(subtotal * 0.2),
      });
    } else if (normalized === "FAMILY200") {
      const subtotal = selectSubtotal(get());
      if (subtotal < 1000) {
        set({ promoError: `Min subtotal ฿1000`, promoCode: null, promoDiscount: 0 });
      } else {
        set({ promoCode: normalized, promoError: null, promoDiscount: 200 });
      }
    } else {
      set({ promoError: "Invalid code", promoCode: null, promoDiscount: 0 });
    }
  },

  clearPromo: () => set({ promoCode: null, promoError: null, promoDiscount: 0 }),

  checkout: () => {
    const totals = selectTotals(get());
    if (totals.subtotal <= 0) return false;
    set({ state: "checking_out" });
    return true;
  },

  reset: () =>
    set({
      state: "empty",
      lines: [],
      promoCode: null,
      promoError: null,
      promoDiscount: 0,
      flashLineId: null,
    }),
}));

export function selectSubtotal(s: Pick<CartState, "lines">): number {
  return s.lines.reduce((acc, l) => acc + l.unitPrice * l.qty, 0);
}

export function selectTotals(s: Pick<CartState, "lines" | "promoDiscount">): CartTotals {
  const subtotal = selectSubtotal(s);
  const serviceCharge = Math.round(subtotal * SERVICE_RATE);
  const tax = Math.round(subtotal * TAX_RATE);
  const discount = Math.min(s.promoDiscount, subtotal);
  const total = Math.max(0, subtotal + serviceCharge + tax - discount);
  return { subtotal, serviceCharge, tax, discount, total };
}
