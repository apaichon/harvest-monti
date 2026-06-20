"use client";
import React from "react";

export type OrderStepState = "complete" | "active" | "pending";

export interface OrderStep {
  id: "received" | "preparing" | "ready";
  label: string;
  state: OrderStepState;
  timestamp?: string;
}

export interface OrderStatusStepperProps {
  steps: OrderStep[];
  eta?: string;
  className?: string;
}

export function OrderStatusStepper({
  steps,
  eta,
  className = "",
}: OrderStatusStepperProps) {
  return (
    <div className={`flex flex-col gap-2 ${className}`} data-testid="order-status-stepper">
      {eta && (
        <div className="rounded-2xl bg-bg-surface px-4 py-2 text-center text-sm font-semibold text-text-primary">
          ETA {eta}
        </div>
      )}
      <ol className="flex flex-col gap-0">
        {steps.map((s, i) => (
          <li key={s.id} className="flex gap-3" data-testid={`step-${s.id}`}>
            <div className="flex flex-col items-center">
              <div
                className={`grid h-10 w-10 place-items-center rounded-full border-2 ${
                  s.state === "complete"
                    ? "border-success bg-success/20 text-success"
                    : s.state === "active"
                      ? "border-accent-cyan text-accent-cyan animate-ring-rotate"
                      : "border-text-disabled text-text-disabled"
                }`}
                aria-label={s.state}
              >
                {s.state === "complete" ? "✓" : s.state === "active" ? "◐" : "◯"}
              </div>
              {i < steps.length - 1 && <div className="h-8 w-0.5 bg-bg-surface-2" />}
            </div>
            <div className="pt-1">
              <p className="font-semibold text-text-primary">{s.label}</p>
              {s.timestamp && (
                <p className="text-xs text-text-secondary">{s.timestamp}</p>
              )}
            </div>
          </li>
        ))}
      </ol>
    </div>
  );
}

export default OrderStatusStepper;
