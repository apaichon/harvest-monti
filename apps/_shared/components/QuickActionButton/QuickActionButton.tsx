"use client";
import React from "react";

export type QuickActionName = "recommendations" | "promotions" | "allergy" | "help";

export interface QuickActionButtonProps {
  name: QuickActionName;
  label: string;
  icon: React.ReactNode;
  onPress: (name: QuickActionName) => void;
  className?: string;
}

/**
 * QuickActionButton — stacked rail button (kiosk) / sheet row (mobile).
 * Voice intent and tap parity: clicking dispatches same action as voice.
 * DES-0008 §4, §6 row 6.
 */
export function QuickActionButton({
  name,
  label,
  icon,
  onPress,
  className = "",
}: QuickActionButtonProps) {
  return (
    <button
      type="button"
      onClick={() => onPress(name)}
      className={`flex w-full items-center gap-3 rounded-2xl bg-bg-surface px-5 py-4 text-left text-base font-semibold text-text-primary transition hover:bg-bg-surface-2 focus:outline-none focus:ring-[3px] focus:ring-accent-cyan focus:ring-offset-4 focus:ring-offset-bg-base ${className}`}
      data-testid={`quick-action-${name}`}
      aria-label={label}
    >
      <span className="text-accent-cyan" aria-hidden="true">
        {icon}
      </span>
      <span>{label}</span>
    </button>
  );
}

export default QuickActionButton;
