// Shared domain types. Mirrors DES-0007 menu/cart contracts.

export type Currency = "THB";

export interface ModifierOption {
  id: string;
  name: string;
  priceDelta: number;
}

export interface ModifierGroup {
  id: string;
  name: string;
  kind: "single" | "multi";
  required: boolean;
  min?: number;
  max?: number;
  options: ModifierOption[];
}

export interface Item {
  id: string;
  sku: string;
  categoryId: string;
  name: string;
  description?: string;
  imageUrl: string;
  price: number;
  badges?: Array<"BEST SELLER" | "NEW" | "Vegetarian">;
  isBestSeller?: boolean;
  isNew?: boolean;
  modifierGroups?: ModifierGroup[];
  inStock?: boolean;
}

export interface Category {
  id: string;
  name: string;
  iconUrl?: string;
}

export interface CartLine {
  id: string;
  itemId: string;
  name: string;
  imageUrl: string;
  unitPrice: number;
  qty: number;
  modifiers: Record<string, string | string[]>;
  modifiersLabel?: string;
}

export interface CartTotals {
  subtotal: number;
  serviceCharge: number;
  tax: number;
  discount: number;
  total: number;
}

export interface Promo {
  code: string;
  label: string;
  description?: string;
  discountKind: "percent" | "fixed";
  discountValue: number;
  minSubtotal?: number;
  validUntil?: string;
}

export type VoiceState = "idle" | "listening" | "processing" | "speaking";

export type Lang = "en-US" | "th-TH" | "zh-CN" | "ja-JP";

export interface TranscriptTurn {
  role: "user" | "assistant";
  text: string;
  final: boolean;
  interrupted?: boolean;
  ts: number;
}

export type OrderStatusStep = "received" | "preparing" | "ready";
