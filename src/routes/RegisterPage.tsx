import { MailCheck } from "lucide-react";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { SuccessFeedback } from "../components/StateBlocks";
import { useViewState } from "./state";

export function RegisterPage() {
  const state = useViewState();
  return (
    <section className="mx-auto grid max-w-6xl gap-8 px-5 py-10 md:grid-cols-[0.92fr_1.08fr] md:py-14">
      <div>
        <p className="text-xs font-semibold uppercase text-muted">普通用户注册</p>
        <h1 className="mt-3 text-4xl font-semibold leading-tight">创建已验证的用户账号</h1>
        <p className="mt-4 text-secondary">
          这个静态表单用于展示邮箱、密码、Turnstile 占位和验证码状态，不会创建真实账号。
        </p>
        <div className="mt-6 rounded-lg border border-border bg-surface p-5 text-sm text-secondary">
          公开注册只会创建 <strong className="text-primary">user</strong> 账号。管理员权限只能由初始化流程或既有管理员授予。
        </div>
      </div>
      <form className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle">
        {state === "success" ? <SuccessFeedback message="验证码已发送到示例邮箱。这只是模拟反馈。" /> : null}
        <Input label="邮箱" placeholder="name@example.invalid" type="email" />
        <Input label="密码" placeholder="至少 12 个字符" type="password" />
        <Input label="确认密码" placeholder="再次输入密码" type="password" />
        <div className="rounded-md border border-dashed border-strong bg-subtle p-4 text-sm text-secondary">
          <div className="flex items-center gap-2 font-medium text-primary">
            <MailCheck className="h-4 w-4" /> Turnstile 占位
          </div>
          <p className="mt-2">M0.1 不接入 Cloudflare Turnstile 或邮件服务，这里只用于视觉占位。</p>
        </div>
        <Button type="button">发送验证码</Button>
      </form>
    </section>
  );
}
