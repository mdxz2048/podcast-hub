import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        canvas: "rgb(var(--color-background-canvas) / <alpha-value>)",
        surface: "rgb(var(--color-background-surface) / <alpha-value>)",
        subtle: "rgb(var(--color-background-subtle) / <alpha-value>)",
        primary: "rgb(var(--color-text-primary) / <alpha-value>)",
        secondary: "rgb(var(--color-text-secondary) / <alpha-value>)",
        muted: "rgb(var(--color-text-muted) / <alpha-value>)",
        border: "rgb(var(--color-border-default) / <alpha-value>)",
        strong: "rgb(var(--color-border-strong) / <alpha-value>)",
        action: "rgb(var(--color-action-primary) / <alpha-value>)",
        "action-hover": "rgb(var(--color-action-primary-hover) / <alpha-value>)",
        success: "rgb(var(--color-status-success) / <alpha-value>)",
        warning: "rgb(var(--color-status-warning) / <alpha-value>)",
        danger: "rgb(var(--color-status-danger) / <alpha-value>)",
        info: "rgb(var(--color-status-info) / <alpha-value>)",
        neutral: "rgb(var(--color-status-neutral) / <alpha-value>)"
      },
      borderRadius: {
        sm: "var(--radius-sm)",
        md: "var(--radius-md)",
        lg: "var(--radius-lg)"
      },
      boxShadow: {
        subtle: "var(--shadow-subtle)",
        overlay: "var(--shadow-overlay)"
      },
      fontFamily: {
        sans: "var(--font-family-sans)",
        mono: "var(--font-family-mono)"
      }
    }
  },
  plugins: []
};

export default config;

