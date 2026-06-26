import { ArrowLeft, ShieldCheck } from "lucide-react";
import { FormEvent, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { verifyPasswordReset } from "../api/auth";
import type { ApiError } from "../api/client";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";
import { useViewState } from "./state";

export function ResetPasswordPage() {
  const state = useViewState();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const [email, setEmail] = useState(params.get("email") ?? "");
  const [proof, setProof] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setSuccess(null);
    if (!email || !proof || !newPassword || !confirmPassword) {
      setError("请完整填写邮箱、验证码和新密码。");
      return;
    }
    setIsSubmitting(true);
    try {
      await verifyPasswordReset({
        email,
        proof,
        new_password: newPassword,
        confirm_password: confirmPassword
      });
      setSuccess("密码重置成功，其他登录设备已退出。即将跳转到登录页。");
      setTimeout(() => navigate("/login?reset=done"), 1200);
    } catch (e) {
      const apiError = e as ApiError;
      if (apiError.code === "invalid_reset_proof") {
        setError("验证码错误，请检查后重试。");
      } else if (apiError.code === "reset_proof_expired") {
        setError("验证码已过期，请重新申请重置。");
      } else if (apiError.code === "rate_limited" || apiError.code === "too_many_attempts") {
        setError("请求过于频繁，请稍后再试。");
      } else if (apiError.code === "network_error") {
        setError("网络连接失败，请稍后重试。");
      } else {
        setError(apiError.message || "重置失败，请稍后重试。");
      }
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="mx-auto grid max-w-5xl gap-8 px-5 py-10 md:grid-cols-[0.9fr_1.1fr] md:py-14">
      <div>
        <p className="text-xs font-semibold uppercase text-muted">重置密码</p>
        <h1 className="mt-3 text-4xl font-semibold leading-tight">输入验证码并设置新密码</h1>
        <p className="mt-4 text-secondary">
          此页面用于提交重置验证码和新密码。成功后会撤销旧会话并返回登录页。
        </p>
        <Link className="mt-6 inline-flex items-center gap-2 text-sm text-secondary hover:text-action" to="/forgot-password">
          <ArrowLeft className="h-4 w-4" /> 返回找回密码
        </Link>
      </div>
      <form className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle" onSubmit={handleSubmit}>
        {state === "loading" || isSubmitting ? <LoadingState title="正在重置密码" /> : null}
        {state === "error" ? <ErrorState title="重置失败，请稍后重试" /> : null}
        {state === "expired" ? <ErrorState title="验证码已过期，请重新申请重置" /> : null}
        {state === "denied" || state === "rate_limited" ? <ErrorState title="请求过于频繁，请稍后再试" /> : null}
        {state === "network" ? <ErrorState title="网络连接失败，请稍后重试" /> : null}
        {success ? <SuccessFeedback message={success} /> : null}
        {error ? <ErrorState title={error} /> : null}
        <Input label="邮箱" type="email" placeholder="name@example.com" value={email} onChange={(event) => setEmail(event.target.value)} />
        <Input label="验证码" inputMode="numeric" placeholder="输入邮件中的验证码" value={proof} onChange={(event) => setProof(event.target.value.trim())} />
        <Input label="新密码" type="password" placeholder="至少 12 个字符" value={newPassword} onChange={(event) => setNewPassword(event.target.value)} />
        <Input label="确认新密码" type="password" placeholder="再次输入新密码" value={confirmPassword} onChange={(event) => setConfirmPassword(event.target.value)} />
        <div className="rounded-md border border-border bg-subtle p-4 text-sm text-secondary">
          <div className="flex items-center gap-2 font-medium text-primary">
            <ShieldCheck className="h-4 w-4" /> 安全提示
          </div>
          <p className="mt-2">重置成功后，系统会使该账号的历史登录会话失效。</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button type="submit" disabled={isSubmitting}>确认重置密码</Button>
          <Button type="button" variant="secondary" onClick={() => navigate("/login")}>返回登录</Button>
        </div>
      </form>
    </section>
  );
}
