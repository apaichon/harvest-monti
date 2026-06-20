"use client";
import React, { useState } from "react";

export interface PromoCodeInputProps {
  value?: string;
  onApply: (code: string) => void | Promise<void>;
  error?: string | null;
  appliedCode?: string | null;
  className?: string;
}

export function PromoCodeInput({
  value = "",
  onApply,
  error,
  appliedCode,
  className = "",
}: PromoCodeInputProps) {
  const [code, setCode] = useState(value);
  return (
    <div className={`flex flex-col gap-1 ${className}`} data-testid="promo-input">
      <label className="text-xs text-text-secondary" htmlFor="promo-code">
        Promo code
      </label>
      <div className="flex items-center gap-2 rounded-2xl bg-bg-base/40 px-3 py-2">
        <input
          id="promo-code"
          type="text"
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase())}
          className="flex-1 bg-transparent text-sm text-text-primary placeholder:text-text-disabled focus:outline-none"
          placeholder="Enter code"
          aria-invalid={!!error}
          aria-describedby={error ? "promo-error" : undefined}
        />
        <button
          type="button"
          onClick={() => onApply(code)}
          aria-label="Apply promo code"
          className="grid h-8 w-8 place-items-center rounded-2xl bg-accent-cyan text-sm font-bold text-bg-base"
          data-testid="promo-apply"
        >
          ✓
        </button>
      </div>
      {appliedCode && !error && (
        <p className="text-xs text-success">Applied: {appliedCode}</p>
      )}
      {error && (
        <p id="promo-error" className="text-xs text-danger" role="alert">
          {error}
        </p>
      )}
    </div>
  );
}

export default PromoCodeInput;
