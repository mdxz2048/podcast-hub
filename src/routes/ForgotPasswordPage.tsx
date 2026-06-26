import { ArrowLeft, Mail } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";
import { useViewState } from "./state";

export function ForgotPasswordPage() {
  const state = useViewState();

  return (
    <section className="mx-auto grid max-w-5xl gap-8 px-5 py-10 md:grid-cols-[0.9fr_1.1fr] md:py-14">
      <div>
        <p className="text-xs font-semibold uppercase text-muted">找回密码</p>
        <h1 className="mt-3 text-4xl font-semibold leading-tight">获取重置说明</h1>
        <p className="mt-4 text-secondary">
          输入邮箱后，页面会显示通用提示，不会暴露该邮箱是否已注册。
        </p>
        <Link className="mt-6 inline-flex items-center gap-2 text-sm text-secondary hover:text-action" to="/login">
          <ArrowLeft className="h-4 w-4" /> 返回登录
        </Link>
      </div>
      <form className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle">
        {state === "loading" ? <LoadingState title="正在准备重置说明" /> : null}
        {state === "success" ? <SuccessFeedback message="如果该邮箱可用，你将收到后续重置说明。这是模拟反馈。" /> : null}
        {state === "error" ? <ErrorState title="暂时无法发送重置说明" /> : null}
        {state === "denied" ? <ErrorState title="请求过于频繁，请稍后重试" /> : null}
        <Input label="邮箱" placeholder="name@example.invalid" type="email" />
        <div className="rounded-md border border-border bg-subtle p-4 text-sm text-secondary">
          <div className="flex items-center gap-2 font-medium text-primary">
            <Mail className="h-4 w-4" /> 安全提示
          </div>
          <p className="mt-2">真实版本会让旧会话失效，并发送安全通知邮件。M0.2A 不发送邮件。</p>
        </div>
        <Button type="button">发送重置说明</Button>
      </form>
    </section>
  );
}
