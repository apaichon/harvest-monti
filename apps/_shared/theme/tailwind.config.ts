// Shared Tailwind theme tokens — DES-0008 §1.1 palette.
// Consumed by apps/kiosk/tailwind.config.ts and apps/mobile (future).
// Untyped to avoid requiring tailwindcss as a dep of the shared workspace pkg —
// consumers cast/import as their own Config["theme"].

export const montiTheme = {
  extend: {
    colors: {
      "bg-base": "#0B1438",
      "bg-surface": "#142051",
      "bg-surface-2": "#1B2A66",
      "accent-cyan": "#00BFFF",
      "accent-cyan-soft": "rgba(51,204,255,0.667)",
      "text-primary": "#FFFFFF",
      "text-secondary": "rgba(255,255,255,0.60)",
      "text-disabled": "rgba(255,255,255,0.38)",
      success: "#22C55E",
      warn: "#F59E0B",
      danger: "#EF4444",
    },
    fontFamily: {
      sans: ["Inter", "system-ui", "sans-serif"],
    },
    borderRadius: {
      "2xl": "24px",
      "3xl": "32px",
    },
    keyframes: {
      "voice-pulse": {
        "0%, 100%": { transform: "scale(0.92)", opacity: "0.6" },
        "50%": { transform: "scale(1.08)", opacity: "1" },
      },
      "monti-bob": {
        "0%, 100%": { transform: "translateY(-6px)" },
        "50%": { transform: "translateY(6px)" },
      },
      "row-flash": {
        "0%": { backgroundColor: "rgba(51,204,255,0.667)" },
        "100%": { backgroundColor: "transparent" },
      },
      "ring-rotate": {
        "0%": { transform: "rotate(0deg)" },
        "100%": { transform: "rotate(360deg)" },
      },
    },
    animation: {
      "voice-pulse": "voice-pulse 120ms ease-in-out infinite",
      "monti-bob": "monti-bob 4000ms ease-in-out infinite",
      "row-flash": "row-flash 600ms ease-out",
      "ring-rotate": "ring-rotate 1200ms linear infinite",
    },
  },
};

export default montiTheme;
export type MontiTheme = typeof montiTheme;
