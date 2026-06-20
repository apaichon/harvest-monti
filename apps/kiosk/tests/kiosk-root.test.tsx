import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import KioskRoot from "../app/page";
import { useCartStore } from "../lib/store/cart";
import { useMenuStore } from "../lib/store/menu";
import { useVoiceStore } from "../lib/store/voice";

describe("Kiosk landscape root (DES-0008 §2)", () => {
  beforeEach(() => {
    useCartStore.getState().reset();
    useMenuStore.setState({ query: "", activeCategoryId: "all", bestSellersOnly: false });
    useVoiceStore.setState({ transcript: [], language: "en-US", banner: null, toast: null });
  });

  it("renders the three columns + bottom nav with the required labels", () => {
    render(<KioskRoot />);

    // Region presence
    expect(screen.getByTestId("kiosk-root")).toBeInTheDocument();
    expect(screen.getByTestId("voice-rail")).toBeInTheDocument();
    expect(screen.getByTestId("center-menu")).toBeInTheDocument();
    expect(screen.getByTestId("order-rail")).toBeInTheDocument();
    expect(screen.getByTestId("bottom-nav")).toBeInTheDocument();
    expect(screen.getByTestId("category-chip-row")).toBeInTheDocument();
    expect(screen.getByTestId("menu-grid")).toBeInTheDocument();

    // Required strings per DES-0008 §2
    expect(screen.getAllByText(/Hi! I'm Monti/).length).toBeGreaterThan(0);
    expect(screen.getByText("Good appetite!")).toBeInTheDocument();
    expect(screen.getByText("Your Order")).toBeInTheDocument();
    expect(screen.getByText(/Table.*A12/)).toBeInTheDocument();

    // Quick actions
    expect(screen.getByTestId("quick-action-recommendations")).toBeInTheDocument();
    expect(screen.getByTestId("quick-action-promotions")).toBeInTheDocument();
    expect(screen.getByTestId("quick-action-allergy")).toBeInTheDocument();
    expect(screen.getByTestId("quick-action-help")).toBeInTheDocument();

    // Bottom nav (4 nav items)
    expect(screen.getByTestId("nav-home")).toBeInTheDocument();
    expect(screen.getByTestId("nav-history")).toBeInTheDocument();
    expect(screen.getByTestId("nav-scan-pay")).toBeInTheDocument();
    expect(screen.getByTestId("nav-call-staff")).toBeInTheDocument();
  });

  it("matches the structural snapshot of the root grid", () => {
    const { container } = render(<KioskRoot />);
    const root = container.querySelector('[data-testid="kiosk-root"]') as HTMLElement;
    expect(root).toBeTruthy();
    // Snapshot the grid template only — fixture image URLs are deterministic.
    expect(root.style.gridTemplateColumns).toMatchInlineSnapshot(
      `"320px 1fr 380px"`,
    );
    expect(root.style.gridTemplateRows).toMatchInlineSnapshot(`"1fr 72px"`);
  });
});
