import { ShieldCheck } from "lucide-react";
import { Button } from "./Button";

interface TurnstileFieldProps {
  value: string;
  onChange: (token: string) => void;
  error?: string | null;
}

function readTurnstileMode() {
  return (import.meta.env.VITE_TURNSTILE_MODE ?? "mock").toLowerCase();
}

export function TurnstileField({ value, onChange, error }: TurnstileFieldProps) {
  const mode = readTurnstileMode();
  const siteKey = import.meta.env.VITE_TURNSTILE_SITE_KEY;

  if (mode === "off") {
    return (
      <div className="rounded-md border border-dashed border-strong bg-subtle p-4 text-sm text-secondary">
        已关闭人机验证（仅限本地开发调试）。
      </div>
    );
  }

  if (mode === "cloudflare") {
    if (!siteKey) {
      return (
        <div className="rounded-md border border-danger/30 bg-danger/5 p-4 text-sm text-danger">
          生产模式缺少 Turnstile Site Key，请联系管理员修复配置。
        </div>
      );
    }
    return (
      <div className="rounded-md border border-dashed border-strong bg-subtle p-4 text-sm text-secondary">
        <div className="flex items-center gap-2 font-medium text-primary">
          <ShieldCheck className="h-4 w-4" />
          Cloudflare Turnstile
        </div>
        <p className="mt-2">当前为生产集成占位。请在部署时挂载真实 Turnstile Widget。</p>
      </div>
    );
  }

  return (
    <div className="rounded-md border border-border bg-subtle p-4 text-sm text-secondary">
      <div className="flex items-center gap-2 font-medium text-primary">
        <ShieldCheck className="h-4 w-4" />
        开发环境模拟人机验证
      </div>
      <p className="mt-2">仅用于本地调试，不会调用 Cloudflare。</p>
      <div className="mt-3 flex flex-wrap items-center gap-2">
        <Button type="button" variant="secondary" onClick={() => onChange("mock-pass")}>
          我是人类（模拟）
        </Button>
        {value ? <span className="text-xs text-success">已通过验证</span> : <span className="text-xs text-muted">尚未验证</span>}
      </div>
      {error ? <p className="mt-2 text-xs text-danger">{error}</p> : null}
    </div>
  );
}
