"use client";
import React from "react";

export type BottomNavId = "home" | "history" | "scan-pay" | "call-staff";

export interface BottomNavItem {
  id: BottomNavId;
  label: string;
  icon: React.ReactNode;
  pulsing?: boolean;
}

export interface BottomNavProps {
  activeId: BottomNavId;
  onSelect: (id: BottomNavId) => void;
  today: string;
  items?: BottomNavItem[];
  className?: string;
}

export const DEFAULT_BOTTOM_NAV_ITEMS: BottomNavItem[] = [
  { id: "home", label: "Home", icon: <span aria-hidden>🏠</span> },
  { id: "history", label: "Order History", icon: <span aria-hidden>🧾</span> },
  { id: "scan-pay", label: "Scan & Pay", icon: <span aria-hidden>📷</span> },
  { id: "call-staff", label: "Call Staff", icon: <span aria-hidden>🛎</span> },
];

export function BottomNav({
  activeId,
  onSelect,
  today,
  items = DEFAULT_BOTTOM_NAV_ITEMS,
  className = "",
}: BottomNavProps) {
  return (
    <nav
      className={`flex h-[72px] w-full items-center justify-between gap-4 bg-bg-surface px-6 ${className}`}
      data-testid="bottom-nav"
      aria-label="Primary"
    >
      <div className="flex items-center gap-2">
        {items.map((it) => {
          const active = it.id === activeId;
          return (
            <button
              key={it.id}
              type="button"
              onClick={() => onSelect(it.id)}
              className={`flex items-center gap-2 rounded-2xl px-4 py-2 text-sm font-semibold transition focus:outline-none focus:ring-[3px] focus:ring-accent-cyan ${
                active
                  ? "text-accent-cyan underline decoration-accent-cyan decoration-4 underline-offset-8"
                  : "text-text-primary hover:bg-bg-surface-2"
              } ${it.pulsing ? "animate-voice-pulse" : ""}`}
              data-testid={`nav-${it.id}`}
              aria-current={active ? "page" : undefined}
            >
              {it.icon}
              <span>{it.label}</span>
            </button>
          );
        })}
      </div>
      <time className="text-sm text-text-secondary" dateTime={today}>
        {today}
      </time>
    </nav>
  );
}

export default BottomNav;
