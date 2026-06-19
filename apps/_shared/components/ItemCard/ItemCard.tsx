"use client";
import React, { useRef } from "react";
import type { Item } from "../../types";

export interface ItemCardProps {
  item: Item;
  onTapAdd: (item: Item) => void;
  onLongPress?: (item: Item) => void;
  dimmed?: boolean;
  highlighted?: boolean;
  longPressMs?: number;
  className?: string;
}

/**
 * ItemCard — image + badges + name + price + add (+) button.
 * Tap (+) adds with defaults. Long-press (350ms) opens detail dialog.
 * DES-0008 §4, §2.1.
 */
export function ItemCard({
  item,
  onTapAdd,
  onLongPress,
  dimmed = false,
  highlighted = false,
  longPressMs = 350,
  className = "",
}: ItemCardProps) {
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const triggeredLongRef = useRef(false);

  const start = () => {
    triggeredLongRef.current = false;
    if (!onLongPress) return;
    timerRef.current = setTimeout(() => {
      triggeredLongRef.current = true;
      onLongPress?.(item);
    }, longPressMs);
  };
  const clear = () => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  };

  const badges = item.badges ?? [];
  if (item.isBestSeller && !badges.includes("BEST SELLER")) badges.unshift("BEST SELLER");
  if (item.isNew && !badges.includes("NEW")) badges.unshift("NEW");

  return (
    <article
      className={`group flex flex-col overflow-hidden rounded-2xl bg-bg-surface transition ${
        dimmed ? "opacity-40" : "opacity-100"
      } ${highlighted ? "ring-2 ring-accent-cyan" : ""} ${className}`}
      data-testid={`item-card-${item.id}`}
      onMouseDown={start}
      onMouseUp={clear}
      onMouseLeave={clear}
      onTouchStart={start}
      onTouchEnd={clear}
    >
      <div
        className="relative aspect-video w-full bg-bg-surface-2"
        style={{
          backgroundImage: `url(${item.imageUrl})`,
          backgroundSize: "cover",
          backgroundPosition: "center",
        }}
      >
        {badges.length > 0 && (
          <div className="absolute left-3 top-3 flex gap-2">
            {badges.map((b) => (
              <span
                key={b}
                className={`rounded-2xl px-2 py-1 text-xs font-bold uppercase tracking-wide ${
                  b === "BEST SELLER"
                    ? "bg-accent-cyan text-bg-base"
                    : b === "NEW"
                      ? "bg-success text-bg-base"
                      : "bg-warn text-bg-base"
                }`}
                data-testid={`badge-${b.toLowerCase().replace(/\s+/g, "-")}`}
              >
                ▸{b}
              </span>
            ))}
          </div>
        )}
      </div>
      <div className="flex flex-1 items-end justify-between gap-2 p-4">
        <div className="min-w-0 flex-1">
          <h3 className="line-clamp-2 text-base font-semibold text-text-primary" title={item.name}>
            {item.name}
          </h3>
          <p className="mt-1 text-sm text-text-secondary">฿ {item.price}</p>
        </div>
        <button
          type="button"
          onClick={(e) => {
            e.stopPropagation();
            if (triggeredLongRef.current) {
              triggeredLongRef.current = false;
              return;
            }
            onTapAdd(item);
          }}
          aria-label={`Add ${item.name} to cart`}
          className="grid h-12 w-12 place-items-center rounded-2xl bg-accent-cyan text-2xl font-bold text-bg-base transition hover:scale-105 focus:outline-none focus:ring-[3px] focus:ring-accent-cyan focus:ring-offset-4 focus:ring-offset-bg-base"
          data-testid={`add-${item.id}`}
        >
          +
        </button>
      </div>
    </article>
  );
}

export default ItemCard;
