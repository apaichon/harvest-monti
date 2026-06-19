"use client";
import React, { useMemo, useState } from "react";
import type { Item, ModifierGroup } from "../../types";

export interface ItemDetailDialogProps {
  item: Item | null;
  open: boolean;
  onClose: () => void;
  onAdd: (args: {
    item: Item;
    qty: number;
    modifiers: Record<string, string | string[]>;
    unitPrice: number;
  }) => void;
}

/**
 * ItemDetailDialog — 960x720 modal (kiosk) with hero, modifier groups, qty, add CTA.
 * Modifier rendering per DES-0008 §7.3:
 *   single+required → radio gated
 *   single+optional → radio + None
 *   multi+required(min≥1) → checkbox gated
 *   multi+optional → checkbox no gate
 */
export function ItemDetailDialog({ item, open, onClose, onAdd }: ItemDetailDialogProps) {
  const [qty, setQty] = useState(1);
  const [state, setState] = useState<Record<string, string | string[]>>({});

  React.useEffect(() => {
    if (open && item) {
      setQty(1);
      const init: Record<string, string | string[]> = {};
      for (const g of item.modifierGroups ?? []) {
        if (g.kind === "multi") init[g.id] = [];
        else if (!g.required) init[g.id] = "__none__";
      }
      setState(init);
    }
  }, [open, item]);

  const priceDelta = useMemo(() => {
    if (!item) return 0;
    let delta = 0;
    for (const g of item.modifierGroups ?? []) {
      const v = state[g.id];
      if (g.kind === "single" && typeof v === "string" && v !== "__none__") {
        const opt = g.options.find((o) => o.id === v);
        delta += opt?.priceDelta ?? 0;
      } else if (g.kind === "multi" && Array.isArray(v)) {
        for (const optId of v) {
          const opt = g.options.find((o) => o.id === optId);
          delta += opt?.priceDelta ?? 0;
        }
      }
    }
    return delta;
  }, [item, state]);

  const gated = useMemo(() => {
    if (!item) return true;
    for (const g of item.modifierGroups ?? []) {
      if (g.required) {
        const v = state[g.id];
        if (g.kind === "single") {
          if (!v || v === "__none__") return true;
        } else if (g.kind === "multi") {
          if (!Array.isArray(v) || v.length < (g.min ?? 1)) return true;
        }
      }
    }
    return false;
  }, [item, state]);

  if (!open || !item) return null;
  const unitPrice = item.price + priceDelta;
  const total = unitPrice * qty;

  return (
    <div
      className="fixed inset-0 z-50 grid place-items-center bg-black/60 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-labelledby="item-detail-title"
      onClick={onClose}
      data-testid="item-detail-dialog"
    >
      <div
        className="grid w-[960px] max-w-[95vw] grid-cols-2 gap-6 rounded-2xl bg-bg-surface p-8 text-text-primary"
        style={{ minHeight: 600, maxHeight: 720 }}
        onClick={(e) => e.stopPropagation()}
      >
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="absolute right-6 top-6 grid h-10 w-10 place-items-center rounded-2xl bg-bg-surface-2 text-text-primary"
        >
          ✕
        </button>
        <div
          className="aspect-video w-full overflow-hidden rounded-2xl bg-bg-surface-2"
          style={{
            backgroundImage: `url(${item.imageUrl})`,
            backgroundSize: "cover",
            backgroundPosition: "center",
          }}
        />
        <div className="flex flex-col gap-4 overflow-y-auto">
          <div>
            <h2 id="item-detail-title" className="text-2xl font-bold">
              {item.name}
            </h2>
            {item.description && (
              <p className="mt-2 text-text-secondary">{item.description}</p>
            )}
          </div>

          {(item.modifierGroups ?? []).map((g) => (
            <ModifierGroupView key={g.id} group={g} value={state[g.id]} setValue={(v) => setState((s) => ({ ...s, [g.id]: v }))} />
          ))}

          <div className="mt-2 flex items-center gap-4">
            <span className="font-semibold">Quantity</span>
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() => setQty((q) => Math.max(1, q - 1))}
                className="h-10 w-10 rounded-2xl bg-bg-surface-2 text-xl"
                aria-label="Decrease quantity"
              >
                −
              </button>
              <span className="w-10 text-center font-bold" data-testid="dialog-qty">
                {qty}
              </span>
              <button
                type="button"
                onClick={() => setQty((q) => q + 1)}
                className="h-10 w-10 rounded-2xl bg-bg-surface-2 text-xl"
                aria-label="Increase quantity"
              >
                +
              </button>
            </div>
          </div>

          <button
            type="button"
            disabled={gated}
            onClick={() =>
              onAdd({
                item,
                qty,
                modifiers: { ...state },
                unitPrice,
              })
            }
            className="mt-auto rounded-2xl bg-accent-cyan py-4 text-center text-base font-bold uppercase tracking-wide text-bg-base disabled:bg-text-disabled"
            data-testid="dialog-add-to-cart"
          >
            ADD TO CART ( ฿ {total} )
          </button>
        </div>
      </div>
    </div>
  );
}

function ModifierGroupView({
  group,
  value,
  setValue,
}: {
  group: ModifierGroup;
  value: string | string[] | undefined;
  setValue: (v: string | string[]) => void;
}) {
  if (group.kind === "single") {
    return (
      <fieldset className="flex flex-col gap-2">
        <legend className="font-semibold">
          {group.name}{" "}
          <span className="text-sm text-text-secondary">
            ({group.required ? "required" : "optional"}, single)
          </span>
        </legend>
        {!group.required && (
          <label className="flex cursor-pointer items-center gap-3">
            <input
              type="radio"
              name={group.id}
              checked={value === "__none__"}
              onChange={() => setValue("__none__")}
            />
            <span>None</span>
          </label>
        )}
        {group.options.map((o) => (
          <label key={o.id} className="flex cursor-pointer items-center gap-3">
            <input
              type="radio"
              name={group.id}
              value={o.id}
              checked={value === o.id}
              onChange={() => setValue(o.id)}
              data-testid={`mod-${group.id}-${o.id}`}
            />
            <span>
              {o.name}
              {o.priceDelta > 0 ? `  + ฿ ${o.priceDelta}` : ""}
            </span>
          </label>
        ))}
      </fieldset>
    );
  }
  // multi
  const arr = Array.isArray(value) ? value : [];
  return (
    <fieldset className="flex flex-col gap-2">
      <legend className="font-semibold">
        {group.name}{" "}
        <span className="text-sm text-text-secondary">
          ({group.required ? "required" : "optional"}, multi)
        </span>
      </legend>
      {group.options.map((o) => {
        const checked = arr.includes(o.id);
        return (
          <label key={o.id} className="flex cursor-pointer items-center gap-3">
            <input
              type="checkbox"
              checked={checked}
              onChange={() =>
                setValue(checked ? arr.filter((x) => x !== o.id) : [...arr, o.id])
              }
              data-testid={`mod-${group.id}-${o.id}`}
            />
            <span>
              {o.name}
              {o.priceDelta > 0 ? `  + ฿ ${o.priceDelta}` : ""}
            </span>
          </label>
        );
      })}
    </fieldset>
  );
}

export default ItemDetailDialog;
