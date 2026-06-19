"use client";
import React from "react";
import type { CartTotals } from "../../types";

export interface OrderSummaryProps {
  totals: CartTotals;
  taxLabel?: string;
  className?: string;
}

/**
 * OrderSummary — Subtotal / Service / Tax / Discount / Total.
 * Discount line shown only when discount > 0. DES-0008 §2 (kiosk) + §3.6 (mobile).
 */
export function OrderSummary({ totals, taxLabel = "Tax 7%", className = "" }: OrderSummaryProps) {
  return (
    <div
      className={`flex flex-col gap-1 text-sm text-text-primary ${className}`}
      data-testid="order-summary"
    >
      <Row label="Subtotal" value={totals.subtotal} />
      <Row label="Service" value={totals.serviceCharge} />
      <Row label={taxLabel} value={totals.tax} />
      {totals.discount > 0 && (
        <Row label="Discount" value={-totals.discount} negative data-testid="row-discount" />
      )}
      <div className="my-1 border-t border-white/10" />
      <div
        className="flex items-center justify-between text-lg font-bold"
        data-testid="row-total"
      >
        <span>TOTAL</span>
        <span>฿ {totals.total}</span>
      </div>
    </div>
  );
}

function Row({
  label,
  value,
  negative = false,
  ...rest
}: {
  label: string;
  value: number;
  negative?: boolean;
} & React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className="flex items-center justify-between" {...rest}>
      <span className="text-text-secondary">{label}</span>
      <span>
        {negative ? "−" : ""}฿ {Math.abs(value)}
      </span>
    </div>
  );
}

export default OrderSummary;
