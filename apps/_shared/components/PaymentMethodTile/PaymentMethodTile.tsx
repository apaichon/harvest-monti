"use client";
import React from "react";

export type PaymentMethodId = "credit-card" | "apple-pay" | "wallet" | "cash";

export interface PaymentMethodTileProps {
  method: PaymentMethodId;
  label: string;
  icon: React.ReactNode;
  selected: boolean;
  onSelect: (m: PaymentMethodId) => void;
  className?: string;
}

export function PaymentMethodTile({
  method,
  label,
  icon,
  selected,
  onSelect,
  className = "",
}: PaymentMethodTileProps) {
  return (
    <button
      type="button"
      onClick={() => onSelect(method)}
      className={`flex w-full items-center gap-3 rounded-2xl border-2 p-4 text-left transition focus:outline-none focus:ring-[3px] focus:ring-accent-cyan ${
        selected
          ? "border-accent-cyan bg-bg-surface-2"
          : "border-transparent bg-bg-surface hover:bg-bg-surface-2"
      } ${className}`}
      data-testid={`pay-${method}`}
      role="radio"
      aria-checked={selected}
    >
      <span className="text-2xl" aria-hidden>
        {icon}
      </span>
      <span className="flex-1 font-semibold text-text-primary">{label}</span>
      <span
        className={`grid h-6 w-6 place-items-center rounded-full border-2 ${
          selected ? "border-accent-cyan" : "border-text-secondary"
        }`}
      >
        {selected && <span className="h-3 w-3 rounded-full bg-accent-cyan" />}
      </span>
    </button>
  );
}

export default PaymentMethodTile;
