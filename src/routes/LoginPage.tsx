import { KeyRound } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { ErrorState, SuccessFeedback } from "../components/StateBlocks";
import { useViewState } from "./state";

export function LoginPage() {
  const state = useViewState();
  return (
    <section className="mx-auto grid max-w-5xl gap-8 px-5 py-10 md:grid-cols-[0.9fr_1.1fr] md:py-14">
      <div>
        <p className="text-xs font-semibold uppercase text-muted">安全会话占位</p>
        <h1 className="mt-3 text-4xl font-semibold leading-tight">回到你的音频内容库</h1>
        <p className="mt-4 text-secondary">
          登录页使用统一的错误文案和静态状态展示，不会执行真实认证。
        </p>
      </div>
      <form className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle">
        {state === "error" ? <ErrorState title="无法完成登录" /> : null}
        {state === "success" ? <SuccessFeedback message="已登录状态仅作为模拟反馈展示。" /> : null}
        <Input label="邮箱" placeholder="name@example.invalid" type="email" />
        <Input label="密码" placeholder="请输入密码" type="password" />
        <div className="rounded-md border border-border bg-subtle p-4 text-sm text-secondary">
          <div className="flex items-center gap-2 font-medium text-primary">
            <KeyRound className="h-4 w-4" /> 高风险校验占位
          </div>
          <p className="mt-2">后续可在这里出现 Turnstile。M0.1 只保留静态视觉状态。</p>
        </div>
        <Button type="button">登录</Button>
        <div className="flex flex-wrap justify-between gap-3 text-sm text-secondary">
          <Link className="hover:text-action" to="/forgot-password">忘记密码</Link>
          <Link className="hover:text-action" to="/register">创建账号</Link>
        </div>
      </form>
    </section>
  );
}
