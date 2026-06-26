type Tone = "success" | "warning" | "danger" | "info" | "neutral";

const toneClass: Record<Tone, string> = {
  success: "border-success/25 bg-success/10 text-success",
  warning: "border-warning/25 bg-warning/10 text-warning",
  danger: "border-danger/25 bg-danger/10 text-danger",
  info: "border-info/25 bg-info/10 text-info",
  neutral: "border-neutral/25 bg-neutral/10 text-neutral"
};

export function Badge({ children, tone = "neutral" }: { children: string; tone?: Tone }) {
  return (
    <span className={`inline-flex items-center rounded-sm border px-2 py-1 text-xs font-medium ${toneClass[tone]}`}>
      {children}
    </span>
  );
}

