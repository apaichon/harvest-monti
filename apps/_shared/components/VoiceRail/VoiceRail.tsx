"use client";
import React from "react";
import { MontiMascot } from "../MontiMascot";
import { VoiceMeter } from "../VoiceMeter";
import { QuickActionButton, type QuickActionName } from "../QuickActionButton";
import type { TranscriptTurn, VoiceState } from "../../types";

export interface VoiceRailProps {
  state: VoiceState;
  level: number;
  transcript: TranscriptTurn[];
  greeting?: string;
  onQuickAction: (name: QuickActionName) => void;
  className?: string;
}

/**
 * VoiceRail — kiosk left 320px column.
 * Hosts MontiMascot + VoiceMeter + greeting + 4 quick actions + transcript.
 * DES-0008 §2 (region), §6.
 */
export function VoiceRail({
  state,
  level,
  transcript,
  greeting = "Hi! I'm Monti\nHow can I help you today?",
  onQuickAction,
  className = "",
}: VoiceRailProps) {
  return (
    <aside
      className={`flex h-full w-[320px] flex-col gap-6 rounded-r-3xl bg-bg-surface/60 p-6 ${className}`}
      data-testid="voice-rail"
      aria-label="Monti voice assistant"
    >
      <div className="flex items-center justify-center pt-2">
        <VoiceMeter level={level} state={state} className="h-[220px] w-[220px]">
          <MontiMascot size={200} bobbing={state === "idle"} />
        </VoiceMeter>
      </div>

      <p className="whitespace-pre-line text-center text-base font-medium text-text-primary">
        {greeting}
      </p>

      <div className="flex flex-col gap-3">
        <QuickActionButton
          name="recommendations"
          label="Recommendations"
          icon={<span aria-hidden>{"✦"}</span>}
          onPress={onQuickAction}
        />
        <QuickActionButton
          name="promotions"
          label="Promotions"
          icon={<span aria-hidden>{"🎁"}</span>}
          onPress={onQuickAction}
        />
        <QuickActionButton
          name="allergy"
          label="Food Allergy"
          icon={<span aria-hidden>{"⚠"}</span>}
          onPress={onQuickAction}
        />
        <QuickActionButton
          name="help"
          label="Need Help?"
          icon={<span aria-hidden>{"🛎"}</span>}
          onPress={onQuickAction}
        />
      </div>

      <div className="mt-auto flex min-h-[120px] flex-col gap-1 overflow-y-auto rounded-2xl bg-bg-base/40 p-3 text-sm">
        <div className="text-text-secondary">— transcript —</div>
        <div
          role="log"
          aria-live="polite"
          aria-label="Voice transcript"
          className="flex flex-col gap-1"
          data-testid="voice-transcript"
        >
          {transcript.slice(-6).map((t, i) => (
            <div
              key={i}
              className={
                t.role === "user"
                  ? "text-text-primary"
                  : t.interrupted
                    ? "text-text-disabled italic"
                    : "text-accent-cyan"
              }
            >
              {t.role === "user" ? "> " : "Monti: "}
              {t.text}
            </div>
          ))}
        </div>
      </div>
    </aside>
  );
}

export default VoiceRail;
