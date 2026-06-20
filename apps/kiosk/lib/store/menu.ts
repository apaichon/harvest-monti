"use client";
import { create } from "zustand";
import type { Category, Item } from "@monti/shared";
import { categories as fxCategories, items as fxItems } from "../fixtures/menu";

export interface MenuState {
  categories: Category[];
  items: Item[];
  activeCategoryId: string;
  query: string;
  pulseCategoryId: string | null;
  setActiveCategory: (id: string) => void;
  setQuery: (q: string) => void;
  pulseCategory: (id: string | null) => void;
  setBestSellersOnly: (on: boolean) => void;
  bestSellersOnly: boolean;
}

export const useMenuStore = create<MenuState>((set) => ({
  categories: fxCategories,
  items: fxItems,
  activeCategoryId: "all",
  query: "",
  pulseCategoryId: null,
  bestSellersOnly: false,
  setActiveCategory: (id) => set({ activeCategoryId: id }),
  setQuery: (q) => set({ query: q }),
  pulseCategory: (id) => {
    set({ pulseCategoryId: id });
    if (id) {
      setTimeout(() => set((s) => (s.pulseCategoryId === id ? { pulseCategoryId: null } : s)), 200);
    }
  },
  setBestSellersOnly: (on) => set({ bestSellersOnly: on }),
}));

export function selectFilteredItems(state: MenuState): Item[] {
  return state.items.filter((it) => {
    if (state.bestSellersOnly && !it.isBestSeller) return false;
    if (state.activeCategoryId !== "all" && it.categoryId !== state.activeCategoryId) return false;
    if (state.query) {
      const q = state.query.toLowerCase();
      if (!it.name.toLowerCase().includes(q) && !(it.description ?? "").toLowerCase().includes(q))
        return false;
    }
    return true;
  });
}
