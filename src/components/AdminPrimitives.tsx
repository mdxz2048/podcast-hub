import { QrCode } from "lucide-react";
import type { ReactNode } from "react";
import { Badge } from "./Badge";

export function MetricCard({ title, value, note, tone = "info" }: { title: string; value: string; note: string; tone?: "info" | "success" | "warning" | "danger" }) {
  const toneClass = {
    info: "text-info",
    success: "text-success",
    warning: "text-warning",
    danger: "text-danger"
  }[tone];

  return (
    <div className="rounded-lg border border-border bg-surface p-4 shadow-subtle">
      <p className="text-sm text-secondary">{title}</p>
      <p className={`mt-1 text-3xl font-semibold ${toneClass}`}>{value}</p>
      <p className="mt-2 text-xs text-muted">{note}</p>
    </div>
  );
}

export function KeyValue({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div>
      <dt className="text-xs text-muted">{label}</dt>
      <dd className="mt-1 text-sm font-medium text-primary">{value}</dd>
    </div>
  );
}

export function ProgressBar({ value }: { value: number }) {
  return (
    <div className="h-2 overflow-hidden rounded-full bg-subtle" aria-label={`进度 ${value}%`}>
      <div className="h-full bg-action" style={{ width: `${Math.max(0, Math.min(100, value))}%` }} />
    </div>
  );
}

export function Timeline({ items }: { items: Array<{ label: string; at: string; tone: "success" | "warning" | "danger" | "info" | "neutral" }> }) {
  return (
    <ol className="grid gap-3">
      {items.map((item) => (
        <li key={`${item.label}-${item.at}`} className="flex gap-3">
          <span className="mt-1 h-2.5 w-2.5 shrink-0 rounded-full bg-action" />
          <div>
            <p className="font-medium">{item.label}</p>
            <p className="text-xs text-muted">{item.at}</p>
          </div>
          <Badge tone={item.tone}>{item.tone === "neutral" ? "记录" : "状态"}</Badge>
        </li>
      ))}
    </ol>
  );
}

export function StaticQrPlaceholder() {
  return (
    <div className="grid gap-4 rounded-lg border border-dashed border-strong bg-subtle p-5 text-center">
      <div className="mx-auto grid h-40 w-40 place-items-center rounded-lg border border-border bg-surface">
        <QrCode className="h-16 w-16 text-muted" />
      </div>
      <div>
        <p className="font-semibold">扫码授权等待中</p>
        <p className="mt-2 text-sm text-secondary">此处为静态演示，不包含真实二维码、认证信息、Cookie 或 Token。</p>
      </div>
    </div>
  );
}
