"use client";
// DES-0008 §6 — six voice intent → UI side-effect mappings.
// Subscribes to the in-process bus and translates each intent into store actions.

import { voiceBus, type IntentEvent } from "./event-bus";
import { useCartStore } from "../store/cart";
import { useMenuStore } from "../store/menu";
import { useOrderStore } from "../store/order";
import { useVoiceStore } from "../store/voice";
import type { Lang } from "@monti/shared";
import { items as fxItems } from "../fixtures/menu";
import { callStaff } from "./call-staff";

let attached = false;

export function attachVoiceDispatcher() {
  if (attached) return;
  attached = true;

  voiceBus.onIntent((evt: IntentEvent) => {
    switch (evt.intent) {
      case "menu_search": {
        // (1) Filter grid; pulse category chip; transcript.
        const query = String(evt.args.query ?? "");
        const category = evt.args.category ? String(evt.args.category) : null;
        const menu = useMenuStore.getState();
        menu.setQuery(query);
        if (category) {
          menu.setActiveCategory(category);
          menu.pulseCategory(category);
        }
        menu.setBestSellersOnly(false);
        useVoiceStore.getState().appendTranscript({
          role: "assistant",
          text: `Looking for ${query || category || "items"}`,
          final: true,
          ts: Date.now(),
        });
        break;
      }
      case "cart_add": {
        // (2) Append line; flash row; count-up total; transcript.
        const itemId = String(evt.args.item_id ?? "");
        const qty = Number(evt.args.qty ?? 1);
        const modifiers = (evt.args.modifiers as Record<string, string | string[]>) ?? {};
        const item = fxItems.find((i) => i.id === itemId);
        if (!item) return;
        useCartStore.getState().addLine({ item, qty, modifiers });
        useVoiceStore.getState().appendTranscript({
          role: "assistant",
          text: `Added: ${item.name} x${qty}`,
          final: true,
          ts: Date.now(),
        });
        break;
      }
      case "cart_remove": {
        // (3) Remove line; recompute; toast.
        const lineId = evt.args.line_id ? String(evt.args.line_id) : null;
        const itemId = evt.args.item_id ? String(evt.args.item_id) : null;
        const cart = useCartStore.getState();
        const target = lineId
          ? cart.lines.find((l) => l.id === lineId)
          : cart.lines.find((l) => l.itemId === itemId);
        if (!target) return;
        cart.removeLine(target.id);
        useVoiceStore.getState().pushToast(`Removed ${target.name}`);
        break;
      }
      case "order_submit": {
        // (4) Move to checkout state; banner.
        const cart = useCartStore.getState();
        cart.checkout();
        useVoiceStore.getState().appendTranscript({
          role: "assistant",
          text: "Confirming your order…",
          final: true,
          ts: Date.now(),
        });
        // Stub: assign a demo order id immediately so Thank-You / Status surfaces have data.
        useOrderStore.getState().setSubmitted(`MO-${Date.now().toString().slice(-6)}`, "15-20 min");
        break;
      }
      case "language_switch": {
        // (5) Swap UI strings; flash flag chip; crossfade.
        const code = (evt.args.lang_code ?? evt.args.bcp47) as Lang | undefined;
        if (!code) return;
        const voice = useVoiceStore.getState();
        voice.setLanguage(code);
        const flagMap: Record<string, string> = {
          "en-US": "🇬🇧",
          "th-TH": "🇹🇭",
          "zh-CN": "🇨🇳",
          "ja-JP": "🇯🇵",
        };
        voice.flashLangFlag(flagMap[code] ?? "🏳");
        break;
      }
      case "call_staff": {
        // (6) Toast; nav button pulse; mascot concerned.
        const reason = (evt.args.reason as string | undefined) ?? "help";
        callStaff(reason);
        break;
      }
      case "quick_action": {
        const name = String(evt.args.name ?? "");
        runQuickAction(name as "recommendations" | "promotions" | "allergy" | "help");
        break;
      }
    }
  });
}

export function runQuickAction(name: "recommendations" | "promotions" | "allergy" | "help") {
  const menu = useMenuStore.getState();
  const voice = useVoiceStore.getState();
  if (name === "recommendations") {
    menu.setBestSellersOnly(true);
    menu.setActiveCategory("all");
    menu.pulseCategory("all");
    voice.appendTranscript({
      role: "assistant",
      text: "Looking for: Recommendations",
      final: true,
      ts: Date.now(),
    });
  } else if (name === "promotions") {
    voice.appendTranscript({
      role: "assistant",
      text: "Two promos today: LUNCH20 — 20% off until 2 PM; FAMILY200 — ฿200 off ฿1000+.",
      final: true,
      ts: Date.now(),
    });
  } else if (name === "allergy") {
    voice.appendTranscript({
      role: "assistant",
      text: "Tell me which allergens to avoid and I'll filter the menu.",
      final: true,
      ts: Date.now(),
    });
  } else if (name === "help") {
    callStaff("help");
  }
}
