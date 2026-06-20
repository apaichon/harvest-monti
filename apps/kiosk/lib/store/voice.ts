"use client";
import { create } from "zustand";
import type { Lang, TranscriptTurn, VoiceState } from "@monti/shared";

export interface VoiceStoreState {
  sessionId: string | null;
  provider: string | null;
  state: VoiceState;
  level: number;
  muted: boolean;
  language: Lang;
  langFlashFlag: string | null;
  transcript: TranscriptTurn[];
  banner: string | null;
  toast: string | null;
  mascotMood: "happy" | "concerned" | "thinking" | "listening" | "speaking";

  setSession: (sessionId: string, provider: string) => void;
  closeSession: (reason?: string) => void;
  setState: (s: VoiceState) => void;
  setLevel: (l: number) => void;
  setMuted: (m: boolean) => void;
  setLanguage: (l: Lang) => void;
  flashLangFlag: (flag: string | null) => void;
  appendTranscript: (t: TranscriptTurn) => void;
  pushBanner: (msg: string | null) => void;
  pushToast: (msg: string | null, ms?: number) => void;
  setMascotMood: (m: VoiceStoreState["mascotMood"], ms?: number) => void;
}

export const useVoiceStore = create<VoiceStoreState>((set) => ({
  sessionId: null,
  provider: null,
  state: "idle",
  level: 0,
  muted: false,
  language: "en-US",
  langFlashFlag: null,
  transcript: [],
  banner: null,
  toast: null,
  mascotMood: "happy",

  setSession: (sessionId, provider) => set({ sessionId, provider, banner: null }),
  closeSession: (reason) => set({ sessionId: null, provider: null, state: "idle", banner: reason ?? null }),
  setState: (s) => set({ state: s }),
  setLevel: (l) => set({ level: Math.max(0, Math.min(1, l)) }),
  setMuted: (m) => set({ muted: m }),
  setLanguage: (l) => {
    set({ language: l });
    if (typeof document !== "undefined") {
      document.documentElement.lang = l;
    }
  },
  flashLangFlag: (flag) => {
    set({ langFlashFlag: flag });
    if (flag) setTimeout(() => set((s) => (s.langFlashFlag === flag ? { langFlashFlag: null } : s)), 2000);
  },
  appendTranscript: (t) => set((s) => ({ transcript: [...s.transcript.slice(-20), t] })),
  pushBanner: (msg) => set({ banner: msg }),
  pushToast: (msg, ms = 3000) => {
    set({ toast: msg });
    if (msg) setTimeout(() => set((s) => (s.toast === msg ? { toast: null } : s)), ms);
  },
  setMascotMood: (m, ms) => {
    set({ mascotMood: m });
    if (ms) setTimeout(() => set({ mascotMood: "happy" }), ms);
  },
}));
