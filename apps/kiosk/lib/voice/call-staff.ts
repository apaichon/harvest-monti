"use client";
// Call Staff action — DES-0008 §6 row 6 + §7.1.
// Triggers Slack webhook proxy (server route in TASK-0010 backend), shows toast,
// pulses bottom-nav button, switches mascot mood for 2s.

import { useVoiceStore } from "../store/voice";
import { postStaffCall } from "./api";
import { useSessionStore } from "../store/session";

let staffPulseSetter: ((on: boolean) => void) | null = null;

export function registerStaffPulse(fn: (on: boolean) => void) {
  staffPulseSetter = fn;
}

export async function callStaff(reason: string = "help") {
  const voice = useVoiceStore.getState();
  const session = useSessionStore.getState();
  voice.pushToast("Staff notified", 3000);
  voice.setMascotMood("concerned", 2000);
  if (staffPulseSetter) {
    staffPulseSetter(true);
    setTimeout(() => staffPulseSetter?.(false), 360); // 3 pulses at 120ms
  }
  try {
    await postStaffCall({
      tenant_id: session.tenantId,
      outlet_id: session.outletId,
      table_code: session.tableCode,
      reason,
      requested_at: new Date().toISOString(),
    });
  } catch {
    // Sink failure is silent per DES-0008 §7.1.
  }
}
