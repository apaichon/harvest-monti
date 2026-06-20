"use client";
// WebSocket client for /api/v1/voice/sessions (DES-0009 §2).
// Implements all 8 client→gateway frames and routes the 10 gateway→client frames
// to the voice store + event bus.

import type { ClientFrame, ServerFrame } from "./frames";
import { voiceBus } from "./event-bus";
import { useVoiceStore } from "../store/voice";

export interface WsClientOptions {
  url: string;
  token: string;
  /** Auto-reconnect with exponential backoff (default true). */
  reconnect?: boolean;
}

export class VoiceWsClient {
  private ws: WebSocket | null = null;
  private opts: Required<WsClientOptions>;
  private backoffMs = 500;
  private pingTimer: ReturnType<typeof setInterval> | null = null;
  private closed = false;

  constructor(opts: WsClientOptions) {
    this.opts = { reconnect: true, ...opts };
  }

  connect() {
    this.closed = false;
    const ws = new WebSocket(this.opts.url);
    this.ws = ws;
    ws.onopen = () => {
      this.backoffMs = 500;
      useVoiceStore.getState().pushBanner(null);
      this.pingTimer = setInterval(() => this.send({ type: "ping" }), 15_000);
    };
    ws.onmessage = (ev) => this.handleMessage(ev.data);
    ws.onerror = () => {
      useVoiceStore.getState().pushBanner("Voice paused");
    };
    ws.onclose = () => {
      if (this.pingTimer) clearInterval(this.pingTimer);
      this.pingTimer = null;
      useVoiceStore.getState().closeSession("Disconnected");
      if (!this.closed && this.opts.reconnect) {
        setTimeout(() => this.connect(), this.backoffMs);
        this.backoffMs = Math.min(this.backoffMs * 2, 15_000);
      }
    };
  }

  close() {
    this.closed = true;
    this.ws?.close();
  }

  send(frame: ClientFrame) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
    this.ws.send(JSON.stringify(frame));
  }

  // ── Client → gateway (8 frames) ──
  startAudio() { this.send({ type: "start_audio" }); }
  stopAudio() { this.send({ type: "stop_audio" }); }
  mute() { this.send({ type: "mute" }); useVoiceStore.getState().setMuted(true); }
  unmute() { this.send({ type: "unmute" }); useVoiceStore.getState().setMuted(false); }
  cancelSpeech() { this.send({ type: "cancel_speech" }); }
  languageSwitch(bcp47: string) { this.send({ type: "language_switch", bcp47 }); }
  quickAction(name: "recommendations" | "promotions" | "allergy" | "help") {
    this.send({ type: "quick_action", name });
  }
  ping() { this.send({ type: "ping" }); }

  // ── Gateway → client (10 frames) ──
  private handleMessage(raw: unknown) {
    if (typeof raw !== "string") return; // ignore binary audio echo on this hop
    let frame: ServerFrame;
    try {
      frame = JSON.parse(raw) as ServerFrame;
    } catch {
      return;
    }
    const voice = useVoiceStore.getState();
    switch (frame.type) {
      case "session_open":
        voice.setSession(frame.session_id, frame.provider);
        voice.setState("idle");
        break;
      case "session_closed":
        voice.closeSession(frame.reason);
        break;
      case "transcript":
        voice.appendTranscript({
          role: frame.role,
          text: frame.text,
          final: frame.final,
          interrupted: frame.interrupted,
          ts: Date.now(),
        });
        voice.setState(frame.role === "assistant" ? "speaking" : "listening");
        break;
      case "function_call":
        voice.setState("processing");
        // bus handles dispatch + JSON parse
        break;
      case "function_result":
        voice.setState("speaking");
        break;
      case "audio_meter": {
        // Map dBFS (-60..0) to 0..1
        const norm = Math.max(0, Math.min(1, (frame.db + 60) / 60));
        voice.setLevel(norm);
        break;
      }
      case "provider_switched":
        voice.pushBanner(`Switched ${frame.from} → ${frame.to}`);
        break;
      case "pong":
        break;
      case "error":
        voice.pushBanner(frame.message);
        break;
      case "quota_warning":
        voice.pushBanner(
          `Voice quota ${frame.used_minutes}/${frame.cap_minutes} min`,
        );
        break;
      default: {
        // Forward-compat: unknown frame → no-op + warn.
        // eslint-disable-next-line no-console
        console.warn("[voice] unknown frame", frame);
      }
    }
    voiceBus.emitFrame(frame);
  }
}
