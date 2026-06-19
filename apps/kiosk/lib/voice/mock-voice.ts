"use client";
// Lightweight in-process mock that emulates the WS gateway during dev.
// Drives the voice store with periodic audio_meter frames and demo transcripts
// so the UI can be exercised without TASK-0010 merged.

import { voiceBus } from "./event-bus";
import { useVoiceStore } from "../store/voice";

let started = false;

export function startMockVoiceLoop() {
  if (started) return;
  started = true;
  const voice = useVoiceStore.getState();
  voice.setSession("mock-session-001", "gemini");
  voice.setState("idle");

  // Idle ambient meter pulse
  setInterval(() => {
    const { state } = useVoiceStore.getState();
    if (state === "listening") {
      voiceBus.emitFrame({ type: "audio_meter", db: -20 - Math.random() * 30 });
    } else {
      voiceBus.emitFrame({ type: "audio_meter", db: -55 });
    }
  }, 120);
}
