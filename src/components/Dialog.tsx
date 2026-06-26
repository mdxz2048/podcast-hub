import type { ReactNode } from "react";
import { Button } from "./Button";

interface DialogProps {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  children?: ReactNode;
  onCancel: () => void;
  onConfirm: () => void;
}

export function Dialog({ open, title, description, confirmLabel, children, onCancel, onConfirm }: DialogProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-black/25 px-4">
      <section className="w-full max-w-md rounded-lg border border-border bg-surface p-5 shadow-overlay" role="dialog" aria-modal="true" aria-labelledby="dialog-title">
        <h2 id="dialog-title" className="text-lg font-semibold">{title}</h2>
        <p className="mt-2 text-sm text-secondary">{description}</p>
        {children ? <div className="mt-4">{children}</div> : null}
        <div className="mt-6 flex flex-wrap justify-end gap-2">
          <Button variant="secondary" onClick={onCancel}>取消</Button>
          <Button variant="danger" onClick={onConfirm}>{confirmLabel}</Button>
        </div>
      </section>
    </div>
  );
}
