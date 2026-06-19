"use client";
// Voice → UI dispatch bus per DES-0008 §5.2.
// Both real WS frames and tests funnel through this single bus.

import type { ServerFrame } from "./frames";

export type VoiceIntent =
  | "menu_search"
  | "cart_add"
  | "cart_remove"
  | "order_submit"
  | "language_switch"
  | "call_staff"
  | "quick_action";

export interface IntentEvent {
  intent: VoiceIntent;
  args: Record<string, unknown>;
  callId?: string;
}

type FrameListener = (frame: ServerFrame) => void;
type IntentListener = (evt: IntentEvent) => void;

class VoiceBus {
  private frameListeners = new Set<FrameListener>();
  private intentListeners = new Set<IntentListener>();

  emitFrame(frame: ServerFrame) {
    this.frameListeners.forEach((l) => l(frame));
    if (frame.type === "function_call") {
      let args: Record<string, unknown> = {};
      try {
        args = JSON.parse(frame.arguments_json || "{}");
      } catch {
        // ignore parse errors — dispatcher will treat as empty args
      }
      this.dispatchIntent({
        intent: frame.name as VoiceIntent,
        args,
        callId: frame.call_id,
      });
    }
  }

  dispatchIntent(evt: IntentEvent) {
    this.intentListeners.forEach((l) => l(evt));
  }

  onFrame(l: FrameListener) {
    this.frameListeners.add(l);
    return () => this.frameListeners.delete(l);
  }

  onIntent(l: IntentListener) {
    this.intentListeners.add(l);
    return () => this.intentListeners.delete(l);
  }

  /** Test hook — equivalent to a `function_call` frame. DES-0008 §5.2. */
  __test_inject(intent: VoiceIntent, args: Record<string, unknown> = {}) {
    this.dispatchIntent({ intent, args });
  }
}

export const voiceBus = new VoiceBus();
