import { MailCheck } from "lucide-react";
import { FormEvent, useMemo, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { requestRegisterCode } from "../api/auth";
import type { ApiError } from "../api/client";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { TurnstileField } from "../components/TurnstileField";
import { ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";
import { useViewState } from "./state";

export function RegisterPage() {
  const state = useViewState();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [email, setEmail] = useState(params.get("email") ?? "");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [turnstileToken, setTurnstileToken] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const mockedStateMessage = useMemo(() => {
    if (state === "success") return "验证码已发送到示例邮箱。这只是模拟反馈。";
    if (state === "error") return "暂时无法发送验证码，请稍后重试。";
    if (state === "loading") return "正在发送验证码";
    return null;
  }, [state]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSuccess(null);
    if (!email || !password || !confirmPassword) {
      setError("请完整填写邮箱和密码。");
      return;
    }
    if (password !== confirmPassword) {
      setError("两次输入的密码不一致。");
      return;
    }
    if (!turnstileToken) {
      setError("请先完成人机验证。");
      return;
    }
    setIsSubmitting(true);
    try {
      await requestRegisterCode({
        email,
        password,
        confirm_password: confirmPassword,
        turnstile_token: turnstileToken
      });
      setSuccess("验证码已发送，请前往邮箱查看并继续验证。");
      navigate(`/register/verify?email=${encodeURIComponent(email)}`);
    } catch (e) {
      const apiError = e as ApiError;
      if (apiError.code === "rate_limited") {
        setError("请求过于频繁，请稍后再试。");
      } else if (apiError.code === "turnstile_failed" || apiError.code === "turnstile_required") {
        setError("人机验证失败，请重新验证。");
      } else {
        setError(apiError.message || "注册请求失败，请稍后重试。");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="mx-auto grid max-w-6xl gap-8 px-5 py-10 md:grid-cols-[0.92fr_1.08fr] md:py-14">
      <div>
        <p className="text-xs font-semibold uppercase text-muted">普通用户注册</p>
        <h1 className="mt-3 text-4xl font-semibold leading-tight">创建已验证的用户账号</h1>
        <p className="mt-4 text-secondary">
          公开注册仅创建普通用户账号。提交后会先发送邮箱验证码，验证成功后建立会话。
        </p>
        <div className="mt-6 rounded-lg border border-border bg-surface p-5 text-sm text-secondary">
          公开注册只会创建 <strong className="text-primary">user</strong> 账号。管理员权限只能由初始化流程或既有管理员授予。
        </div>
      </div>
      <form className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle" onSubmit={handleSubmit}>
        {state === "loading" || isSubmitting ? <LoadingState title="正在发送验证码" /> : null}
        {mockedStateMessage && state === "success" ? <SuccessFeedback message={mockedStateMessage} /> : null}
        {success ? <SuccessFeedback message={success} /> : null}
        {state === "error" ? <ErrorState title="暂时无法发送验证码" /> : null}
        {error ? <ErrorState title={error} /> : null}
        <Input label="邮箱" placeholder="name@example.com" type="email" value={email} onChange={(event) => setEmail(event.target.value)} />
        <Input label="密码" placeholder="至少 12 个字符" type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
        <Input label="确认密码" placeholder="再次输入密码" type="password" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} />
        <TurnstileField value={turnstileToken} onChange={setTurnstileToken} error={null} />
        <div className="rounded-md border border-dashed border-strong bg-subtle p-4 text-sm text-secondary">
          <div className="flex items-center gap-2 font-medium text-primary">
            <MailCheck className="h-4 w-4" /> 安全策略
          </div>
          <p className="mt-2">验证码短时有效、单次使用，重发会使旧验证码失效。</p>
        </div>
        <Button type="submit" disabled={isSubmitting}>发送验证码</Button>
        <p className="text-xs text-secondary">
          已有账号？<Link to="/login" className="text-action hover:text-action-hover">去登录</Link>
        </p>
      </form>
    </section>
  );
}
