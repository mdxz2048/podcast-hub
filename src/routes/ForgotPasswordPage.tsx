import { ArrowLeft, Mail } from "lucide-react";
import { FormEvent, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { requestPasswordReset } from "../api/auth";
import type { ApiError } from "../api/client";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { TurnstileField } from "../components/TurnstileField";
import { ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";
import { useViewState } from "./state";

export function ForgotPasswordPage() {
  const navigate = useNavigate();
  const state = useViewState();
  const [email, setEmail] = useState("");
  const [turnstileToken, setTurnstileToken] = useState("");
  const [requesting, setRequesting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const mockedMessage = useMemo(() => {
    if (state === "loading") return "正在准备重置说明";
    if (state === "success") return "如果该邮箱可用，你将收到后续重置说明。";
    if (state === "error") return "暂时无法发送重置说明";
    if (state === "denied") return "请求过于频繁，请稍后重试";
    return null;
  }, [state]);

  async function handleRequest(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSuccess(null);
    if (!email) {
      setError("请输入邮箱。");
      return;
    }
    if (!turnstileToken) {
      setError("请先完成人机验证。");
      return;
    }
    setRequesting(true);
    try {
      await requestPasswordReset({ email, turnstile_token: turnstileToken });
      setSuccess("如果该邮箱可用，你将收到后续重置说明。");
    } catch (e) {
      const apiError = e as ApiError;
      if (apiError.code === "rate_limited") {
        setError("请求过于频繁，请稍后重试。");
      } else {
        setError(apiError.message || "暂时无法发送重置说明。");
      }
    } finally {
      setRequesting(false);
    }
  }

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
      <div className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle">
        {state === "loading" || requesting ? <LoadingState title="正在准备重置说明" /> : null}
        {state === "success" ? <SuccessFeedback message={mockedMessage ?? "如果该邮箱可用，你将收到后续重置说明。"} /> : null}
        {state === "error" || state === "denied" ? <ErrorState title={mockedMessage ?? "请求失败"} /> : null}
        {success ? <SuccessFeedback message={success} /> : null}
        {error ? <ErrorState title={error} /> : null}
        <form className="grid gap-4" onSubmit={handleRequest}>
          <Input label="邮箱" placeholder="name@example.com" type="email" value={email} onChange={(event) => setEmail(event.target.value)} />
          <TurnstileField value={turnstileToken} onChange={setTurnstileToken} error={null} />
          <Button type="submit" disabled={requesting}>发送重置说明</Button>
        </form>
        <Button type="button" variant="secondary" onClick={() => navigate(`/reset-password?email=${encodeURIComponent(email)}`)} disabled={!email}>
          我已收到邮件，去重置密码
        </Button>
        <div className="rounded-md border border-border bg-subtle p-4 text-sm text-secondary">
          <div className="flex items-center gap-2 font-medium text-primary">
            <Mail className="h-4 w-4" /> 安全提示
          </div>
          <p className="mt-2">重置成功后，系统会撤销该账号旧会话并发送安全通知邮件。</p>
        </div>
      </div>
    </section>
  );
}
