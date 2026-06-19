import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "MONTI AI Kiosk",
  description: "Voice-guided self-order kiosk for Monti AI.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en-US">
      <body className="bg-bg-base text-text-primary antialiased">{children}</body>
    </html>
  );
}
