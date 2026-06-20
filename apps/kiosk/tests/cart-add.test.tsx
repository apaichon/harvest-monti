import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import KioskRoot from "../app/page";
import { useCartStore } from "../lib/store/cart";
import { useMenuStore } from "../lib/store/menu";
import { useVoiceStore } from "../lib/store/voice";
import { voiceBus } from "../lib/voice/event-bus";
import { attachVoiceDispatcher } from "../lib/voice/dispatcher";

describe("voice cart_add function_call mutates cart + renders OrderLineRow (DES-0008 §6 row 2)", () => {
  beforeEach(() => {
    useCartStore.getState().reset();
    useMenuStore.setState({ query: "", activeCategoryId: "all", bestSellersOnly: false });
    useVoiceStore.setState({ transcript: [], language: "en-US" });
    attachVoiceDispatcher();
  });

  it("function_call frame appends a line and updates the OrderRail", () => {
    render(<KioskRoot />);

    expect(useCartStore.getState().lines).toHaveLength(0);

    act(() => {
      voiceBus.emitFrame({
        type: "function_call",
        name: "cart_add",
        arguments_json: JSON.stringify({
          item_id: "truffle-pasta-001",
          qty: 1,
          modifiers: { "add-side": "salad" },
        }),
        call_id: "call-1",
      });
    });

    const state = useCartStore.getState();
    expect(state.lines).toHaveLength(1);
    expect(state.lines[0].name).toBe("Truffle Pasta");
    expect(state.lines[0].qty).toBe(1);
    expect(state.state).toBe("adding");

    // Order rail row rendered.
    const lineId = state.lines[0].id;
    expect(screen.getByTestId(`order-line-${lineId}`)).toBeInTheDocument();
    expect(screen.getByTestId(`order-line-${lineId}`).getAttribute("data-flash")).toBe("1");

    // Transcript line appended.
    expect(useVoiceStore.getState().transcript.at(-1)?.text).toMatch(/Added: Truffle Pasta/);
  });
});
