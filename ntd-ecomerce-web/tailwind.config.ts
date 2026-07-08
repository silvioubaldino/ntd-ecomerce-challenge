import type { Config } from "tailwindcss";
import defaultTheme from "tailwindcss/defaultTheme";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["InterVariable", "Inter", ...defaultTheme.fontFamily.sans],
      },
      colors: {
        // Brand: deep, desaturated indigo — trustworthy, calm, "fintech" blue.
        brand: {
          50: "#f2f5fc",
          100: "#e2e9f7",
          200: "#cbd8f2",
          300: "#a7bfe9",
          400: "#7d9dde",
          500: "#5e7ed4",
          600: "#4a63c7",
          700: "#4053b6",
          800: "#394694",
          900: "#323d76",
          950: "#1e2447",
        },
      },
      boxShadow: {
        card: "0 1px 2px 0 rgb(15 23 42 / 0.04), 0 8px 24px -12px rgb(15 23 42 / 0.12)",
        "card-hover":
          "0 1px 2px 0 rgb(15 23 42 / 0.04), 0 12px 32px -12px rgb(15 23 42 / 0.18)",
        dialog: "0 24px 64px -16px rgb(15 23 42 / 0.35)",
      },
    },
  },
  plugins: [],
} satisfies Config;
