import { AlertTriangle, CheckCircle2 } from "lucide-react";

export interface ToastMessage {
  id: string;
  tone: "success" | "danger";
  title: string;
  message: string;
}

export function ToastViewport({ messages }: { messages: ToastMessage[] }) {
  if (messages.length === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50 grid w-[calc(100%-2rem)] max-w-sm gap-3" aria-live="polite" aria-atomic="true">
      {messages.map((toast) => {
        const Icon = toast.tone === "success" ? CheckCircle2 : AlertTriangle;
        const toneClass = toast.tone === "success" ? "border-success/25 bg-success/10" : "border-danger/25 bg-danger/10";
        return (
          <div key={toast.id} role="status" className={`flex gap-3 rounded-lg border bg-surface p-4 shadow-overlay ${toneClass}`}>
            <Icon className={`mt-0.5 h-5 w-5 shrink-0 ${toast.tone === "success" ? "text-success" : "text-danger"}`} />
            <div>
              <p className="font-semibold">{toast.title}</p>
              <p className="text-sm text-secondary">{toast.message}</p>
            </div>
          </div>
        );
      })}
    </div>
  );
}
