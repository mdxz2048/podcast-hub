import { ArrowLeft, MailCheck } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function RegisterVerifyPage() {
  const [params] = useSearchParams();
  const state = params.get("state");

  return (
    <section className="mx-auto grid max-w-5xl gap-8 px-5 py-10 md:grid-cols-[0.9fr_1.1fr] md:py-14">
      <div>
        <p className="text-xs font-semibold uppercase text-muted">邮箱验证</p>
        <h1 className="mt-3 text-4xl font-semibold leading-tight">输入 6 位验证码</h1>
        <p className="mt-4 text-secondary">
          验证码已发送到 <strong className="text-primary">user@example.invalid</strong>。这是静态模拟流程，不会发送真实邮件。
        </p>
        <Link className="mt-6 inline-flex items-center gap-2 text-sm text-secondary hover:text-action" to="/register">
          <ArrowLeft className="h-4 w-4" /> 更换邮箱
        </Link>
      </div>
      <form className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle">
        {state === "loading" ? <LoadingState title="正在验证验证码" /> : null}
        {state === "success" ? <SuccessFeedback message="邮箱验证成功，模拟会话已建立。" /> : null}
        {state === "error" ? <ErrorState title="验证码不正确" /> : null}
        {state === "expired" ? <ErrorState title="验证码已过期" /> : null}
        {state === "rate_limited" ? <ErrorState title="重发过于频繁" /> : null}
        {state === "network" ? <ErrorState title="网络暂不可用" /> : null}
        <Input label="验证码" inputMode="numeric" maxLength={6} placeholder="000000" />
        <div className="rounded-md border border-border bg-subtle p-4 text-sm text-secondary">
          <div className="flex items-center gap-2 font-medium text-primary">
            <MailCheck className="h-4 w-4" /> 剩余 04:32
          </div>
          <p className="mt-2">验证码仅允许单次使用。M0.2A 不连接真实验证服务。</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button type="button">验证并进入</Button>
          <Button type="button" variant="secondary">重新发送</Button>
        </div>
      </form>
    </section>
  );
}
