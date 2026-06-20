import type { Config } from "tailwindcss";
import { montiTheme } from "../_shared/theme/tailwind.config";

const config: Config = {
  content: [
    "./app/**/*.{ts,tsx}",
    "./lib/**/*.{ts,tsx}",
    "../_shared/components/**/*.{ts,tsx}",
  ],
  theme: montiTheme as unknown as Config["theme"],
  plugins: [],
};

export default config;
