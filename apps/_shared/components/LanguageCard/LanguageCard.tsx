"use client";
import React from "react";
import type { Lang } from "../../types";

export interface LanguageCardProps {
  lang: Lang;
  flag: string;
  label: string;
  selected: boolean;
  onSelect: (lang: Lang) => void;
  className?: string;
}

/** Mobile-only language card. Tagged kiosk-hidden via parent layout. */
export function LanguageCard({
  lang,
  flag,
  label,
  selected,
  onSelect,
  className = "",
}: LanguageCardProps) {
  return (
    <button
      type="button"
      onClick={() => onSelect(lang)}
      className={`flex flex-col items-center gap-2 rounded-2xl border-2 p-4 transition focus:outline-none focus:ring-[3px] focus:ring-accent-cyan ${
        selected ? "border-accent-cyan bg-bg-surface-2" : "border-transparent bg-bg-surface"
      } ${className}`}
      role="radio"
      aria-checked={selected}
      data-testid={`lang-${lang}`}
    >
      <span className="text-3xl" aria-hidden>
        {flag}
      </span>
      <span className="font-semibold text-text-primary">{label}</span>
      <span
        className={`grid h-5 w-5 place-items-center rounded-full border-2 ${
          selected ? "border-accent-cyan" : "border-text-secondary"
        }`}
      >
        {selected && <span className="h-2.5 w-2.5 rounded-full bg-accent-cyan" />}
      </span>
    </button>
  );
}

export default LanguageCard;
