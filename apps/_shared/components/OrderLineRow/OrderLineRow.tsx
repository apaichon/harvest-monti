"use client";
import React from "react";
import type { CartLine } from "../../types";

export interface OrderLineRowProps {
  line: CartLine;
  flash?: boolean;
  onQtyChange: (lineId: string, qty: number) => void;
  onRemove: (lineId: string) => void;
  className?: string;
}

export function OrderLineRow({
  line,
  flash = false,
  onQtyChange,
  onRemove,
  className = "",
}: OrderLineRowProps) {
  const lineTotal = line.unitPrice * line.qty;
  return (
    <div
      className={`flex items-center gap-3 rounded-2xl p-2 ${flash ? "animate-row-flash" : ""} ${className}`}
      data-testid={`order-line-${line.id}`}
      data-flash={flash ? "1" : "0"}
    >
      <div
        className="h-12 w-12 shrink-0 rounded-2xl bg-bg-surface-2"
        style={{
          backgroundImage: `url(${line.imageUrl})`,
          backgroundSize: "cover",
          backgroundPosition: "center",
        }}
        aria-hidden="true"
      />
      <div className="flex min-w-0 flex-1 flex-col">
        <span className="truncate text-sm font-semibold text-text-primary" title={line.name}>
          {line.name}
        </span>
        {line.modifiersLabel && (
          <span className="truncate text-xs text-text-secondary">{line.modifiersLabel}</span>
        )}
        <div className="mt-1 flex items-center gap-2">
          <button
            type="button"
            onClick={() => onQtyChange(line.id, line.qty - 1)}
            aria-label="Decrease quantity"
            className="h-8 w-8 rounded-2xl bg-bg-surface-2 text-sm font-bold"
            data-testid={`line-dec-${line.id}`}
          >
            −
          </button>
          <span className="w-6 text-center text-sm" data-testid={`line-qty-${line.id}`}>
            {line.qty}
          </span>
          <button
            type="button"
            onClick={() => onQtyChange(line.id, line.qty + 1)}
            aria-label="Increase quantity"
            className="h-8 w-8 rounded-2xl bg-bg-surface-2 text-sm font-bold"
            data-testid={`line-inc-${line.id}`}
          >
            +
          </button>
          <span className="ml-auto text-sm font-semibold text-text-primary">฿ {lineTotal}</span>
        </div>
      </div>
      <button
        type="button"
        onClick={() => onRemove(line.id)}
        aria-label={`Remove ${line.name}`}
        className="h-8 w-8 rounded-2xl text-danger hover:bg-bg-surface-2"
        data-testid={`line-remove-${line.id}`}
      >
        ✕
      </button>
    </div>
  );
}

export default OrderLineRow;
