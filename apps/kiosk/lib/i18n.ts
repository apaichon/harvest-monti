"use client";
// Minimal client-side string bundle. Enough for TC-11 language_switch.

import type { Lang } from "@monti/shared";

type StringKey =
  | "goodAppetite"
  | "viewAllCategories"
  | "searchPlaceholder"
  | "sortBy"
  | "yourOrder"
  | "table"
  | "promoCode"
  | "checkout"
  | "greeting";

const bundles: Record<Lang, Record<StringKey, string>> = {
  "en-US": {
    goodAppetite: "Good appetite!",
    viewAllCategories: "View All Categories ›",
    searchPlaceholder: "Search dishes...",
    sortBy: "Sort by: Popular",
    yourOrder: "Your Order",
    table: "Table",
    promoCode: "Promo code",
    checkout: "CHECKOUT",
    greeting: "Hi! I'm Monti\nHow can I help you today?",
  },
  "th-TH": {
    goodAppetite: "ทานให้อร่อย!",
    viewAllCategories: "ดูหมวดทั้งหมด ›",
    searchPlaceholder: "ค้นหาเมนู...",
    sortBy: "เรียงตาม: ยอดนิยม",
    yourOrder: "รายการของคุณ",
    table: "โต๊ะ",
    promoCode: "โค้ดส่วนลด",
    checkout: "ชำระเงิน",
    greeting: "สวัสดี! ผมมอนติ\nวันนี้ให้ช่วยอะไรดีครับ?",
  },
  "zh-CN": {
    goodAppetite: "祝您用餐愉快！",
    viewAllCategories: "查看全部分类 ›",
    searchPlaceholder: "搜索菜品...",
    sortBy: "排序：热门",
    yourOrder: "您的订单",
    table: "桌号",
    promoCode: "优惠码",
    checkout: "结账",
    greeting: "你好！我是 Monti\n今天有什么可以帮您的？",
  },
  "ja-JP": {
    goodAppetite: "召し上がれ！",
    viewAllCategories: "すべてのカテゴリ ›",
    searchPlaceholder: "メニューを検索...",
    sortBy: "並び替え：人気",
    yourOrder: "ご注文",
    table: "テーブル",
    promoCode: "プロモコード",
    checkout: "お会計",
    greeting: "こんにちは！モンティです\n本日はどうされますか？",
  },
};

export function t(lang: Lang, key: StringKey): string {
  return bundles[lang]?.[key] ?? bundles["en-US"][key];
}
