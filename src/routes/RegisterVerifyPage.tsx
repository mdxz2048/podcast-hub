import { ArrowLeft, MailCheck } from "lucide-react";
import { FormEvent, useMemo, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { verifyRegisterCode } from "../api/auth";
import type { ApiError } from "../api/client";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function RegisterVerifyPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const state = params.get("state");
  const email = params.get("email") ?? "";
  const [code, setCode] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const mockedMessage = useMemo(() => {
    if (state === "success") return "邮箱验证成功，真实会话已建立。";
    if (state === "error") return "验证码错误，请检查后重试。";
    if (state === "expired") return "验证码已过期，请重新发送。";
    if (state === "rate_limited") return "尝试次数过多，请稍后重试。";
    if (state === "network") return "网络暂不可用，请稍后重试。";
    return null;
  }, [state]);

  async function handleVerify(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSuccess(null);
    if (!email) {
      setError("缺少注册邮箱，请返回注册页重新提交。");
      return;
    }
    if (!code || code.length < 6) {
      setError("请输入 6 位验证码。");
      return;
    }
    setIsSubmitting(true);
    try {
      await verifyRegisterCode({ email, code });
      setSuccess("验证成功，正在进入节目页。");
      navigate("/programs");
    } catch (e) {
      const apiError = e as ApiError;
      if (apiError.code === "invalid_or_expired_code") {
        setError("验证码错误或已过期，请检查后重试。");
      } else if (apiError.code === "too_many_attempts") {
        setError("尝试次数过多，请重新发送验证码。");
      } else {
        setError(apiError.message || "验证失败，请稍后重试。");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="mx-auto grid max-w-5xl gap-8 px-5 py-10 md:grid-cols-[0.9fr_1.1fr] md:py-14">
      <div>
        <p className="text-xs font-semibold uppercase text-muted">邮箱验证</p>
        <h1 className="mt-3 text-4xl font-semibold leading-tight">输入 6 位验证码</h1>
        <p className="mt-4 text-secondary">
          验证码已发送到 <strong className="text-primary break-all">{email || "你的邮箱"}</strong>。
        </p>
        <Link className="mt-6 inline-flex items-center gap-2 text-sm text-secondary hover:text-action" to="/register">
          <ArrowLeft className="h-4 w-4" /> 更换邮箱
        </Link>
      </div>
      <form className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle" onSubmit={handleVerify}>
        {state === "loading" || isSubmitting ? <LoadingState title="正在验证验证码" /> : null}
        {state === "success" ? <SuccessFeedback message={mockedMessage ?? "邮箱验证成功，真实会话已建立。"} /> : null}
        {state === "error" || state === "expired" || state === "rate_limited" || state === "network" ? <ErrorState title={mockedMessage ?? "验证码验证失败"} /> : null}
        {success ? <SuccessFeedback message={success} /> : null}
        {error ? <ErrorState title={error} /> : null}
        <Input label="验证码" inputMode="numeric" maxLength={6} placeholder="000000" value={code} onChange={(event) => setCode(event.target.value.replace(/\D/g, "").slice(0, 6))} />
        <div className="rounded-md border border-border bg-subtle p-4 text-sm text-secondary">
          <div className="flex items-center gap-2 font-medium text-primary">
            <MailCheck className="h-4 w-4" /> 剩余 04:32
          </div>
          <p className="mt-2">验证码仅允许单次使用，失败次数超过限制会被锁定并要求重发。</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button type="submit" disabled={isSubmitting}>验证并进入</Button>
          <Button type="button" variant="secondary" onClick={() => navigate(`/register?email=${encodeURIComponent(email)}`)}>重新发送</Button>
        </div>
      </form>
    </section>
  );
}
