"use client";
import React from "react";

export interface CheckoutBarProps {
  total: number;
  disabled?: boolean;
  onCheckout: () => void;
  label?: string;
  className?: string;
}

export function CheckoutBar({
  total,
  disabled = false,
  onCheckout,
  label = "CHECKOUT",
  className = "",
}: CheckoutBarProps) {
  return (
    <button
      type="button"
      onClick={onCheckout}
      disabled={disabled}
      className={`flex w-full items-center justify-between rounded-2xl bg-accent-cyan px-5 py-4 text-base font-bold uppercase tracking-wide text-bg-base transition disabled:bg-text-disabled disabled:cursor-not-allowed focus:outline-none focus:ring-[3px] focus:ring-accent-cyan focus:ring-offset-4 focus:ring-offset-bg-base ${className}`}
      data-testid="checkout-bar"
    >
      <span>{label}</span>
      <span>→ ฿ {total}</span>
    </button>
  );
}

export default CheckoutBar;
