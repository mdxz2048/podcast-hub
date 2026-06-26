import { KeyRound } from "lucide-react";
import { FormEvent, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useNavigate, useSearchParams } from "react-router-dom";
import type { ApiError } from "../api/client";
import { useAuth } from "../auth/AuthProvider";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { TurnstileField } from "../components/TurnstileField";
import { ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";
import { useViewState } from "./state";

export function LoginPage() {
  const state = useViewState();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [turnstileToken, setTurnstileToken] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const mockedMessage = useMemo(() => {
    if (state === "error") return "登录失败，请检查凭据后重试。";
    if (state === "success") return "登录成功，示例状态用于截图验收。";
    if (state === "focus") return "键盘焦点已定位到邮箱输入框，用于可访问性截图验证。";
    return null;
  }, [state]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSuccess(null);
    if (!email || !password) {
      setError("请输入邮箱和密码。");
      return;
    }
    setIsSubmitting(true);
    try {
      await login({
        email,
        password,
        turnstile_token: turnstileToken || undefined
      });
      setSuccess("登录成功，正在进入节目页。");
      navigate("/programs");
    } catch (e) {
      const apiError = e as ApiError;
      if (apiError.code === "invalid_credentials") {
        setError("邮箱或密码错误。");
      } else if (apiError.code === "rate_limited") {
        setError("请求过于频繁，请稍后再试。");
      } else {
        setError(apiError.message || "登录失败，请稍后重试。");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="mx-auto grid max-w-5xl gap-8 px-5 py-10 md:grid-cols-[0.9fr_1.1fr] md:py-14">
      <div>
        <p className="text-xs font-semibold uppercase text-muted">安全会话占位</p>
        <h1 className="mt-3 text-4xl font-semibold leading-tight">回到你的音频内容库</h1>
        <p className="mt-4 text-secondary">
          登录失败统一返回通用文案，避免暴露邮箱是否存在。
        </p>
      </div>
      <form className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle" onSubmit={handleSubmit}>
        {isSubmitting ? <LoadingState title="正在验证登录信息" /> : null}
        {state === "error" ? <ErrorState title={mockedMessage ?? "无法完成登录"} /> : null}
        {state === "success" || state === "focus" ? <SuccessFeedback message={mockedMessage ?? "登录成功"} /> : null}
        {params.get("reset") === "done" ? <SuccessFeedback message="密码已重置，其他登录设备已退出，请使用新密码重新登录。" /> : null}
        {success ? <SuccessFeedback message={success} /> : null}
        {error ? <ErrorState title={error} /> : null}
        <Input label="邮箱" placeholder="name@example.com" type="email" value={email} onChange={(event) => setEmail(event.target.value)} autoFocus={state === "focus"} />
        <Input label="密码" placeholder="请输入密码" type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
        <TurnstileField value={turnstileToken} onChange={setTurnstileToken} error={null} />
        <div className="rounded-md border border-border bg-subtle p-4 text-sm text-secondary">
          <div className="flex items-center gap-2 font-medium text-primary">
            <KeyRound className="h-4 w-4" /> 会话安全
          </div>
          <p className="mt-2">成功登录后会写入 HttpOnly Cookie，不会写入 localStorage。</p>
        </div>
        <Button type="submit" disabled={isSubmitting}>登录</Button>
        <div className="flex flex-wrap justify-between gap-3 text-sm text-secondary">
          <Link className="hover:text-action" to="/forgot-password">忘记密码</Link>
          <Link className="hover:text-action" to="/register">创建账号</Link>
        </div>
      </form>
    </section>
  );
}
