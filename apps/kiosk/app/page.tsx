"use client";
import { useEffect, useMemo, useState } from "react";
import { VoiceRail } from "@monti/shared/components/VoiceRail";
import { CategoryChipRow } from "@monti/shared/components/CategoryChipRow";
import { ItemCard } from "@monti/shared/components/ItemCard";
import { ItemDetailDialog } from "@monti/shared/components/ItemDetailDialog";
import { OrderLineRow } from "@monti/shared/components/OrderLineRow";
import { OrderSummary } from "@monti/shared/components/OrderSummary";
import { PromoCodeInput } from "@monti/shared/components/PromoCodeInput";
import { CheckoutBar } from "@monti/shared/components/CheckoutBar";
import { BottomNav, DEFAULT_BOTTOM_NAV_ITEMS } from "@monti/shared/components/BottomNav";
import type { BottomNavId, BottomNavItem } from "@monti/shared/components/BottomNav";
import type { Item } from "@monti/shared";

import { useMenuStore, selectFilteredItems } from "../lib/store/menu";
import { useCartStore, selectTotals } from "../lib/store/cart";
import { useVoiceStore } from "../lib/store/voice";
import { useSessionStore } from "../lib/store/session";
import { attachVoiceDispatcher, runQuickAction } from "../lib/voice/dispatcher";
import { callStaff, registerStaffPulse } from "../lib/voice/call-staff";
import { startMockVoiceLoop } from "../lib/voice/mock-voice";
import { t } from "../lib/i18n";

export default function KioskRoot() {
  // ── Wire the voice dispatcher + mock loop once on mount ──
  useEffect(() => {
    attachVoiceDispatcher();
    startMockVoiceLoop();
  }, []);

  // ── Menu state ──
  const categories = useMenuStore((s) => s.categories);
  const activeCategoryId = useMenuStore((s) => s.activeCategoryId);
  const query = useMenuStore((s) => s.query);
  const pulseCategoryId = useMenuStore((s) => s.pulseCategoryId);
  const setActiveCategory = useMenuStore((s) => s.setActiveCategory);
  const setQuery = useMenuStore((s) => s.setQuery);
  const filteredItems = useMenuStore(selectFilteredItems);

  // ── Cart state ──
  const lines = useCartStore((s) => s.lines);
  const flashLineId = useCartStore((s) => s.flashLineId);
  const promoCode = useCartStore((s) => s.promoCode);
  const promoError = useCartStore((s) => s.promoError);
  const cartState = useCartStore((s) => s.state);
  const promoDiscount = useCartStore((s) => s.promoDiscount);
  const totals = useMemo(
    () => selectTotals({ lines, promoDiscount }),
    [lines, promoDiscount],
  );
  const addLine = useCartStore((s) => s.addLine);
  const updateQty = useCartStore((s) => s.updateQty);
  const removeLine = useCartStore((s) => s.removeLine);
  const applyPromo = useCartStore((s) => s.applyPromo);
  const checkoutCart = useCartStore((s) => s.checkout);

  // ── Voice state ──
  const voiceState = useVoiceStore((s) => s.state);
  const voiceLevel = useVoiceStore((s) => s.level);
  const transcript = useVoiceStore((s) => s.transcript);
  const lang = useVoiceStore((s) => s.language);
  const langFlashFlag = useVoiceStore((s) => s.langFlashFlag);
  const banner = useVoiceStore((s) => s.banner);
  const toast = useVoiceStore((s) => s.toast);

  // ── Session ──
  const tableCode = useSessionStore((s) => s.tableCode);

  // ── Detail dialog ──
  const [detailItem, setDetailItem] = useState<Item | null>(null);

  // ── BottomNav active + Call Staff pulsing ──
  const [activeNav, setActiveNav] = useState<BottomNavId>("home");
  const [staffPulsing, setStaffPulsing] = useState(false);
  useEffect(() => {
    registerStaffPulse(setStaffPulsing);
  }, []);

  const navItems: BottomNavItem[] = DEFAULT_BOTTOM_NAV_ITEMS.map((it) =>
    it.id === "call-staff" ? { ...it, pulsing: staffPulsing } : it,
  );

  const onNav = (id: BottomNavId) => {
    setActiveNav(id);
    if (id === "call-staff") {
      callStaff("help");
    }
    // History / Scan & Pay routing is stubbed for kiosk — only Home and call-staff resolve in this surface.
  };

  const today = new Date().toISOString().slice(0, 10);

  return (
    <main
      className="grid h-screen w-screen overflow-hidden bg-bg-base text-text-primary"
      style={{ gridTemplateColumns: "320px 1fr 380px", gridTemplateRows: "1fr 72px" }}
      data-testid="kiosk-root"
    >
      {/* ── Voice rail (col 1, row 1) ── */}
      <div className="row-start-1 col-start-1">
        <VoiceRail
          state={voiceState}
          level={voiceLevel}
          transcript={transcript}
          greeting={t(lang, "greeting")}
          onQuickAction={(name) => runQuickAction(name)}
        />
      </div>

      {/* ── Center menu (col 2, row 1) ── */}
      <section
        className="row-start-1 col-start-2 flex flex-col gap-4 overflow-hidden p-6"
        data-testid="center-menu"
      >
        <header className="flex items-center justify-between">
          <h1 className="text-3xl font-bold">{t(lang, "goodAppetite")}</h1>
          <button
            className="text-sm text-text-secondary hover:text-accent-cyan"
            type="button"
          >
            {t(lang, "viewAllCategories")}
          </button>
        </header>

        <CategoryChipRow
          categories={categories}
          activeId={activeCategoryId}
          onChange={setActiveCategory}
          pulseId={pulseCategoryId}
        />

        <div className="flex items-center gap-3">
          <div className="flex flex-1 items-center gap-2 rounded-2xl bg-bg-surface px-4 py-3">
            <span aria-hidden>🔍</span>
            <input
              type="search"
              placeholder={t(lang, "searchPlaceholder")}
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="flex-1 bg-transparent text-base text-text-primary placeholder:text-text-disabled focus:outline-none"
              aria-label={t(lang, "searchPlaceholder")}
            />
          </div>
          <div className="rounded-2xl bg-bg-surface px-4 py-3 text-sm text-text-secondary">
            {t(lang, "sortBy")} ▼
          </div>
        </div>

        <div className="grid flex-1 grid-cols-2 gap-4 overflow-y-auto pb-4" data-testid="menu-grid">
          {filteredItems.map((item) => (
            <ItemCard
              key={item.id}
              item={item}
              onTapAdd={(it) => addLine({ item: it, qty: 1 })}
              onLongPress={(it) => setDetailItem(it)}
            />
          ))}
          {filteredItems.length === 0 && (
            <p className="col-span-2 mt-12 text-center text-text-secondary">
              No items match your search.
            </p>
          )}
        </div>
      </section>

      {/* ── Order rail (col 3, row 1) ── */}
      <aside
        className="row-start-1 col-start-3 flex h-full flex-col gap-4 overflow-hidden rounded-l-3xl bg-bg-surface/60 p-6"
        data-testid="order-rail"
      >
        <header className="flex items-center justify-between">
          <h2 className="text-xl font-bold">{t(lang, "yourOrder")}</h2>
          <span className="rounded-2xl bg-bg-surface px-3 py-1 text-sm font-semibold">
            {t(lang, "table")} [ {tableCode} ]
          </span>
        </header>

        <div className="flex-1 overflow-y-auto" data-testid="order-lines">
          {lines.length === 0 ? (
            <p className="mt-8 text-center text-text-secondary">Cart is empty.</p>
          ) : (
            lines.map((l) => (
              <OrderLineRow
                key={l.id}
                line={l}
                flash={flashLineId === l.id}
                onQtyChange={updateQty}
                onRemove={removeLine}
              />
            ))
          )}
        </div>

        <PromoCodeInput
          onApply={applyPromo}
          error={promoError}
          appliedCode={promoCode}
        />

        <OrderSummary totals={totals} />

        <CheckoutBar
          total={totals.total}
          disabled={lines.length === 0}
          onCheckout={() => checkoutCart()}
          label={t(lang, "checkout")}
        />

        {cartState === "checking_out" && (
          <p className="text-center text-sm text-accent-cyan">Moving to payment…</p>
        )}
      </aside>

      {/* ── Bottom nav (cols 1-3, row 2) ── */}
      <div className="row-start-2 col-span-3">
        <BottomNav activeId={activeNav} onSelect={onNav} today={today} items={navItems} />
      </div>

      {/* ── Detail dialog ── */}
      <ItemDetailDialog
        item={detailItem}
        open={detailItem !== null}
        onClose={() => setDetailItem(null)}
        onAdd={({ item, qty, modifiers, unitPrice }) => {
          addLine({ item, qty, modifiers, unitPrice });
          setDetailItem(null);
        }}
      />

      {/* ── Banner / toast / lang flash ── */}
      {banner && (
        <div className="fixed left-1/2 top-4 z-40 -translate-x-1/2 rounded-2xl bg-warn px-4 py-2 text-sm font-semibold text-bg-base">
          {banner}
        </div>
      )}
      {toast && (
        <div
          role="status"
          aria-live="polite"
          className="fixed bottom-24 left-1/2 z-40 -translate-x-1/2 rounded-2xl bg-bg-surface-2 px-5 py-3 text-sm font-semibold text-text-primary"
          data-testid="toast"
        >
          {toast}
        </div>
      )}
      {langFlashFlag && (
        <div
          className="fixed right-6 top-6 z-40 rounded-2xl bg-bg-surface-2 px-4 py-2 text-2xl"
          data-testid="lang-flash"
        >
          {langFlashFlag}
        </div>
      )}
    </main>
  );
}
