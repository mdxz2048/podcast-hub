import { ArrowRight, Library, Radio, ShieldCheck } from "lucide-react";
import type { CSSProperties } from "react";
import { Link } from "react-router-dom";
import { Button } from "../components/Button";
import { Badge } from "../components/Badge";
import { programs } from "../mock/data";

export function HomePage() {
  return (
    <div className="px-5 pb-12">
      <section className="mx-auto grid max-w-7xl gap-10 py-10 md:grid-cols-[1.02fr_0.98fr] md:items-center md:py-14">
        <div>
          <Badge tone="info">授权音频内容库</Badge>
          <h1 className="mt-5 max-w-3xl text-5xl font-semibold leading-tight md:text-6xl">Podcast Hub</h1>
          <p className="mt-5 max-w-2xl text-lg leading-relaxed text-secondary">
            一个克制、内容导向的节目接入与发布平台，用于管理内容来源、审核单集、控制授权，并向用户交付可信赖的 RSS 订阅。
          </p>
          <div className="mt-7 flex flex-wrap gap-3">
            <Link to="/register">
              <Button icon={<ArrowRight className="h-4 w-4" />}>创建账号</Button>
            </Link>
            <Link to="/login">
              <Button variant="secondary">登录</Button>
            </Link>
          </div>
          <div className="mt-8 grid gap-3 text-sm text-secondary sm:grid-cols-3">
            <span className="flex items-center gap-2"><Library className="h-4 w-4 text-action" /> 节目库</span>
            <span className="flex items-center gap-2"><ShieldCheck className="h-4 w-4 text-action" /> 权限控制</span>
            <span className="flex items-center gap-2"><Radio className="h-4 w-4 text-action" /> 个人 RSS</span>
          </div>
        </div>
        <div className="grid gap-4">
          {programs.slice(0, 3).map((program, index) => (
            <article key={program.id} className="grid grid-cols-[84px_1fr] gap-4 rounded-lg border border-border bg-surface p-3 shadow-subtle">
              <div
                className="cover-art relative aspect-square rounded-md"
                style={{ "--cover-a": program.coverTone[0], "--cover-b": program.coverTone[1] } as CSSProperties}
                aria-hidden="true"
              />
              <div className="min-w-0 py-1">
                <p className="text-xs text-muted">精选 {index + 1}</p>
                <h2 className="truncate font-semibold">{program.title}</h2>
                <p className="mt-1 line-clamp-2 text-sm text-secondary">{program.description}</p>
              </div>
            </article>
          ))}
        </div>
      </section>
      <section className="mx-auto grid max-w-7xl gap-4 border-t border-border pt-8 md:grid-cols-3">
        {["内容接入", "发布前审核", "实时 RSS 授权"].map((title) => (
          <div key={title} className="rounded-lg border border-border bg-surface p-5">
            <h2 className="font-semibold">{title}</h2>
            <p className="mt-2 text-sm text-secondary">M0.1 静态页面，只使用共享 Token、基础组件和模拟数据。</p>
          </div>
        ))}
      </section>
    </div>
  );
}
