import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "node:path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: [
      {
        find: /^@monti\/shared$/,
        replacement: path.resolve(__dirname, "../_shared/index.ts"),
      },
      {
        find: /^@monti\/shared\/(.*)$/,
        replacement: path.resolve(__dirname, "../_shared") + "/$1",
      },
      { find: "@", replacement: path.resolve(__dirname, ".") },
    ],
  },
  test: {
    environment: "happy-dom",
    globals: true,
    setupFiles: ["./tests/setup.ts"],
    include: ["tests/**/*.test.{ts,tsx}"],
  },
});
