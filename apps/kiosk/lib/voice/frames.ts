// DES-0009 §2 WebSocket control frame schema.

// ───── Client → gateway (8 frames; `ping` is keepalive) ─────
export type ClientFrame =
  | { type: "start_audio" }
  | { type: "stop_audio" }
  | { type: "mute" }
  | { type: "unmute" }
  | { type: "cancel_speech" }
  | { type: "language_switch"; bcp47: string }
  | { type: "quick_action"; name: "recommendations" | "promotions" | "allergy" | "help" }
  | { type: "ping" };

export const CLIENT_FRAME_TYPES = [
  "start_audio",
  "stop_audio",
  "mute",
  "unmute",
  "cancel_speech",
  "language_switch",
  "quick_action",
  "ping",
] as const;

// ───── Gateway → client (10 frames) ─────
export type ServerFrame =
  | { type: "session_open"; session_id: string; provider: string }
  | { type: "session_closed"; reason: string }
  | {
      type: "transcript";
      role: "user" | "assistant";
      text: string;
      final: boolean;
      interrupted?: boolean;
    }
  | { type: "function_call"; name: string; arguments_json: string; call_id: string }
  | {
      type: "function_result";
      call_id: string;
      success: boolean;
      payload?: unknown;
      error?: string;
    }
  | { type: "audio_meter"; db: number }
  | { type: "provider_switched"; from: string; to: string; reason: string }
  | { type: "pong" }
  | { type: "error"; code: string; message: string }
  | { type: "quota_warning"; tenant_id: string; used_minutes: number; cap_minutes: number };

export const SERVER_FRAME_TYPES = [
  "session_open",
  "session_closed",
  "transcript",
  "function_call",
  "function_result",
  "audio_meter",
  "provider_switched",
  "pong",
  "error",
  "quota_warning",
] as const;
