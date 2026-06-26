import { forwardRef } from "react";
import type { ButtonHTMLAttributes, ReactNode } from "react";

type ButtonVariant = "primary" | "secondary" | "ghost" | "danger";

const variantClass: Record<ButtonVariant, string> = {
  primary: "bg-action text-white hover:bg-action-hover border-action",
  secondary: "bg-surface text-primary hover:bg-subtle border-border",
  ghost: "bg-transparent text-secondary hover:bg-subtle border-transparent",
  danger: "bg-danger text-white hover:brightness-95 border-danger"
};

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  icon?: ReactNode;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  { variant = "primary", icon, children, className = "", ...props },
  ref
) {
  return (
    <button
      ref={ref}
      className={`inline-flex min-h-10 items-center justify-center gap-2 rounded-md border px-4 py-2 text-sm font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-action focus-visible:ring-offset-2 focus-visible:ring-offset-canvas disabled:cursor-not-allowed disabled:opacity-50 ${variantClass[variant]} ${className}`}
      {...props}
    >
      {icon}
      {children ? <span>{children}</span> : null}
    </button>
  );
});
