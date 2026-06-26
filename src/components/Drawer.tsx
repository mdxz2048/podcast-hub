import { X } from "lucide-react";
import type { ReactNode } from "react";
import { Button } from "./Button";

interface DrawerProps {
  open: boolean;
  title: string;
  children: ReactNode;
  onClose: () => void;
}

export function Drawer({ open, title, children, onClose }: DrawerProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-40 bg-black/20">
      <div className="ml-auto flex h-full w-full flex-col border-l border-border bg-surface shadow-overlay md:max-w-xl">
        <header className="flex items-center justify-between border-b border-border px-5 py-4">
          <h2 className="text-lg font-semibold">{title}</h2>
          <Button variant="ghost" className="px-2" onClick={onClose} aria-label="关闭抽屉">
            <X className="h-5 w-5" />
          </Button>
        </header>
        <div className="flex-1 overflow-y-auto p-5">{children}</div>
      </div>
    </div>
  );
}
