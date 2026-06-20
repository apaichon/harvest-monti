import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import KioskRoot from "../app/page";
import { useVoiceStore } from "../lib/store/voice";
import { useCartStore } from "../lib/store/cart";
import { voiceBus } from "../lib/voice/event-bus";
import { attachVoiceDispatcher } from "../lib/voice/dispatcher";

describe("language_switch updates lang attribute + re-renders strings (DES-0008 §6 row 5)", () => {
  beforeEach(() => {
    useCartStore.getState().reset();
    useVoiceStore.setState({ transcript: [], language: "en-US", langFlashFlag: null });
    document.documentElement.lang = "en-US";
    attachVoiceDispatcher();
  });

  it("switches to th-TH on function_call language_switch", () => {
    const { rerender } = render(<KioskRoot />);

    expect(screen.getByText("Good appetite!")).toBeInTheDocument();
    expect(document.documentElement.lang).toBe("en-US");

    act(() => {
      voiceBus.emitFrame({
        type: "function_call",
        name: "language_switch",
        arguments_json: JSON.stringify({ lang_code: "th-TH" }),
        call_id: "call-lang",
      });
    });

    rerender(<KioskRoot />);

    expect(useVoiceStore.getState().language).toBe("th-TH");
    expect(document.documentElement.lang).toBe("th-TH");
    expect(screen.getByText("ทานให้อร่อย!")).toBeInTheDocument();
    expect(screen.getByText("รายการของคุณ")).toBeInTheDocument();
    expect(screen.getByTestId("lang-flash").textContent).toBe("🇹🇭");
  });
});
