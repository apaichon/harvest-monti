"use client";
import React from "react";
import type { Category } from "../../types";

export interface CategoryChipRowProps {
  categories: Category[];
  activeId: string;
  onChange: (id: string) => void;
  pulseId?: string | null;
  className?: string;
}

export function CategoryChipRow({
  categories,
  activeId,
  onChange,
  pulseId,
  className = "",
}: CategoryChipRowProps) {
  return (
    <div
      className={`flex w-full gap-3 overflow-x-auto pb-2 ${className}`}
      role="tablist"
      aria-label="Menu categories"
      data-testid="category-chip-row"
    >
      {categories.map((c) => {
        const active = c.id === activeId;
        const pulse = c.id === pulseId;
        return (
          <button
            key={c.id}
            type="button"
            role="tab"
            aria-selected={active}
            onClick={() => onChange(c.id)}
            className={`shrink-0 rounded-2xl px-5 py-2 text-base font-semibold transition focus:outline-none focus:ring-[3px] focus:ring-accent-cyan focus:ring-offset-4 focus:ring-offset-bg-base ${
              active
                ? "bg-accent-cyan text-bg-base"
                : "bg-bg-surface text-text-primary hover:bg-bg-surface-2"
            } ${pulse ? "animate-voice-pulse" : ""}`}
            data-testid={`category-chip-${c.id}`}
          >
            {c.name}
          </button>
        );
      })}
    </div>
  );
}

export default CategoryChipRow;
