"use client";
import React from "react";
import type { VoiceState } from "../../types";

export interface VoiceMeterProps {
  level: number; // 0..1
  state: VoiceState;
  children?: React.ReactNode;
  className?: string;
}

/**
 * VoiceMeter — pulse-animated cyan ring around mascot.
 * DES-0008 §1.5 (pulse 120ms loop scale 0.92→1.08), §5.2 (states).
 *
 * - idle:       ring dim, no pulse
 * - listening:  cyan pulse 120ms loop, scale by level
 * - processing: clockwise rotation
 * - speaking:   steady cyan ring
 */
export function VoiceMeter({ level, state, children, className = "" }: VoiceMeterProps) {
  const clamped = Math.max(0, Math.min(1, level));
  const ringOpacity =
    state === "idle" ? 0.15 : state === "speaking" ? 1 : 0.6 + clamped * 0.4;
  const ringScale = state === "listening" ? 1 + clamped * 0.1 : 1;
  const ringClass =
    state === "listening"
      ? "animate-voice-pulse"
      : state === "processing"
        ? "animate-ring-rotate"
        : "";

  return (
    <div
      className={`relative inline-flex items-center justify-center ${className}`}
      data-state={state}
      data-testid="voice-meter"
      data-level={clamped.toFixed(2)}
    >
      <div
        className={`absolute inset-0 rounded-full border-4 border-accent-cyan ${ringClass}`}
        style={{
          opacity: ringOpacity,
          transform: `scale(${ringScale})`,
          transition: "opacity 120ms linear",
        }}
        data-testid="voice-meter-ring"
      />
      {children}
    </div>
  );
}

export default VoiceMeter;
