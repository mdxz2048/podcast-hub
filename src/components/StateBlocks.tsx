import { AlertTriangle, CheckCircle2, Loader2, Lock, SearchX } from "lucide-react";
import { Button } from "./Button";

export function LoadingState({ title = "正在加载工作区" }: { title?: string }) {
  return (
    <div className="rounded-lg border border-border bg-surface p-8 text-center shadow-subtle">
      <Loader2 className="mx-auto mb-4 h-6 w-6 animate-spin text-info" />
      <h2 className="text-lg font-semibold">{title}</h2>
      <p className="mt-2 text-sm text-secondary">正在准备模拟数据和页面状态。</p>
    </div>
  );
}

export function EmptyState({ title, action }: { title: string; action?: string }) {
  return (
    <div className="rounded-lg border border-dashed border-strong bg-surface p-8 text-center">
      <SearchX className="mx-auto mb-4 h-7 w-7 text-muted" />
      <h2 className="text-lg font-semibold">{title}</h2>
      <p className="mx-auto mt-2 max-w-md text-sm text-secondary">这个状态暂时没有模拟内容，页面仍会保持一致的信息节奏。</p>
      {action ? <Button className="mt-5">{action}</Button> : null}
    </div>
  );
}

export function ErrorState({ title = "无法加载当前视图" }: { title?: string }) {
  return (
    <div className="rounded-lg border border-danger/30 bg-danger/5 p-8">
      <AlertTriangle className="mb-4 h-7 w-7 text-danger" />
      <h2 className="text-lg font-semibold">{title}</h2>
      <p className="mt-2 text-sm text-secondary">这是静态错误状态，没有发起真实请求，也不会暴露密钥或后端细节。</p>
      <Button variant="secondary" className="mt-5">重试视图</Button>
    </div>
  );
}

export function PermissionDeniedState() {
  return (
    <div className="rounded-lg border border-warning/30 bg-warning/5 p-8">
      <Lock className="mb-4 h-7 w-7 text-warning" />
      <h2 className="text-lg font-semibold">需要权限</h2>
      <p className="mt-2 text-sm text-secondary">当前模拟账号不能执行该操作，请联系系统负责人检查权限。</p>
    </div>
  );
}

export function SuccessFeedback({ message }: { message: string }) {
  return (
    <div className="flex items-start gap-3 rounded-md border border-success/25 bg-success/10 p-4 text-sm text-primary">
      <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-success" />
      <div>
        <p className="font-semibold">操作成功</p>
        <p className="text-secondary">{message}</p>
      </div>
    </div>
  );
}
