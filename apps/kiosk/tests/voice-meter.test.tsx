import { describe, it, expect } from "vitest";
import { render, act } from "@testing-library/react";
import { VoiceMeter } from "@monti/shared/components/VoiceMeter";
import { VoiceWsClient } from "../lib/voice/ws-client";
import { useVoiceStore } from "../lib/store/voice";

describe("VoiceMeter pulses on audio_meter frames (DES-0008 §1.5, §5.2)", () => {
  it("listening state has the 120ms pulse class and audio_meter updates store level", () => {
    useVoiceStore.setState({ state: "listening", level: 0 });

    const r0 = render(<VoiceMeter level={0} state="listening" />);
    const ring0 = r0.container.querySelector('[data-testid="voice-meter-ring"]') as HTMLElement;
    expect(ring0.className).toContain("animate-voice-pulse");

    // Simulate an audio_meter frame arriving at the WS client (DES-0009 §2 gateway→client).
    const client = new VoiceWsClient({ url: "ws://test", token: "t", reconnect: false });
    act(() => {
      (client as unknown as { handleMessage: (s: string) => void }).handleMessage(
        JSON.stringify({ type: "audio_meter", db: -10 }),
      );
    });
    const after = useVoiceStore.getState().level;
    expect(after).toBeGreaterThan(0);

    // Speaking state: no pulse class.
    const r2 = render(<VoiceMeter level={after} state="speaking" />);
    const ring2 = r2.container.querySelector('[data-testid="voice-meter-ring"]') as HTMLElement;
    expect(ring2.className).not.toContain("animate-voice-pulse");
  });
});
