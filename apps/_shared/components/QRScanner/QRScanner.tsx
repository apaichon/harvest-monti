"use client";
import React from "react";

export interface QRScannerProps {
  onScan: (value: string) => void;
  onError?: (err: Error) => void;
  /** Stub mode renders placeholder; full impl lives in TASK-0012. */
  stub?: boolean;
  className?: string;
}

/**
 * QRScanner — mobile-only. On kiosk this is a stub (no camera permission).
 * Real impl arrives in TASK-0012 (Flutter mobile + html5-qrcode for web fallback).
 */
export function QRScanner({ onScan, onError, stub = true, className = "" }: QRScannerProps) {
  if (stub) {
    return (
      <div
        className={`grid aspect-square w-full place-items-center rounded-2xl border-2 border-dashed border-accent-cyan bg-bg-surface ${className}`}
        role="img"
        aria-label="QR scanner placeholder"
        data-testid="qr-scanner-stub"
      >
        <div className="text-center text-text-secondary">
          <p className="text-4xl">📷</p>
          <p className="mt-2 text-sm">QR Scanner</p>
          <p className="text-xs text-text-disabled">(stub — TASK-0012)</p>
          <button
            type="button"
            className="mt-3 rounded-2xl bg-accent-cyan px-3 py-1 text-xs font-bold text-bg-base"
            onClick={() => onScan("DEMO-TABLE-A12")}
          >
            Simulate scan
          </button>
        </div>
      </div>
    );
  }
  // Live mode would set up getUserMedia — out of scope for kiosk.
  return (
    <div className={className} data-testid="qr-scanner-live">
      <video aria-label="Camera viewport" />
    </div>
  );
}

export default QRScanner;
