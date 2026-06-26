import { ArrowLeft, CheckCircle2, Save } from "lucide-react";
import type { CSSProperties } from "react";
import type { ReactNode } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { EpisodeRow } from "../components/EpisodeRow";
import { Input } from "../components/Form";
import { KeyValue, MetricCard, Timeline } from "../components/AdminPrimitives";
import { EmptyState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import { episodes, jobs, programs, reviewItems, sources } from "../mock/data";
import { findProgram } from "../mock/selectors";
import { accessStateLabel, authModeLabel, ingestionTypeLabel, jobStatusLabel, programStatusLabel, publicationStateLabel, rightsStateLabel, sourceStatusLabel, triggerTypeLabel } from "../utils/labels";

export function AdminProgramDetailPage() {
  const { programId } = useParams();
  const [params] = useSearchParams();
  const state = params.get("state");
  if (state === "denied") return <PermissionDeniedState />;

  const baseProgram = state === "long"
    ? programs.find((program) => program.id === "program_field_archive")
    : findProgram(programId);
  if (!baseProgram) return <EmptyState title="节目不存在" />;

  const program = state === "draft"
    ? { ...baseProgram, status: "draft" as const, publicationState: "private" as const }
    : state === "auth"
      ? { ...baseProgram, rightsState: "needs_note" as const, accessState: "blocked" as const }
      : state === "paused"
        ? { ...baseProgram, publicationState: "paused" as const }
        : baseProgram;

  const programSources = state === "no_sources" ? [] : sources.filter((source) => source.programId === program.id);
  const programEpisodes = state === "no_episodes" ? [] : episodes.filter((episode) => episode.programId === program.id);
  const programJobs = jobs.filter((job) => job.programId === program.id).slice(0, 4);
  const pendingReviews = reviewItems.filter((item) => item.programId === program.id && item.status !== "approved").length;
  const latestJob = programJobs[0];
  const saved = params.get("toast") === "success";

  return (
    <div className="grid gap-6">
      <Link to="/admin/programs" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回节目列表
      </Link>
      <section className="grid gap-5 rounded-lg border border-border bg-surface p-5 shadow-subtle lg:grid-cols-[180px_1fr_auto]">
        <div className="cover-art relative aspect-square rounded-lg" style={{ "--cover-a": program.coverTone[0], "--cover-b": program.coverTone[1] } as CSSProperties} />
        <div className="min-w-0">
          <div className="flex flex-wrap gap-2">
            <Badge tone={program.status === "active" ? "success" : "warning"}>{programStatusLabel[program.status]}</Badge>
            <Badge tone={program.rightsState === "clear" ? "success" : "warning"}>{rightsStateLabel[program.rightsState]}</Badge>
            <Badge>{publicationStateLabel[program.publicationState]}</Badge>
          </div>
          <h1 className="mt-4 text-3xl font-semibold leading-tight md:text-4xl">{program.title}</h1>
          <p className="mt-3 max-w-3xl text-secondary">{program.description}</p>
          <dl className="mt-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <KeyValue label="访问范围" value={accessStateLabel[program.accessState]} />
            <KeyValue label="作者" value={program.author} />
            <KeyValue label="分类" value={program.category} />
            <KeyValue label="更新频率" value={program.updateFrequency} />
          </dl>
        </div>
        <div className="flex flex-wrap items-start gap-2 lg:justify-end">
          <Button icon={<Save className="h-4 w-4" />}>保存 Mock 编辑</Button>
          <Button variant="secondary">添加来源</Button>
        </div>
      </section>
      {saved ? <SuccessFeedback message="节目配置已保存到模拟状态，没有调用真实 API。" /> : null}
      <section className="grid gap-4 md:grid-cols-4">
        <MetricCard title="绑定来源" value={String(programSources.length)} note="来源决定导入方式" />
        <MetricCard title="最近导入" value={latestJob ? jobStatusLabel[latestJob.status] : "无任务"} note={latestJob?.nextAction ?? "需要添加来源"} tone={latestJob?.status === "failed" ? "danger" : "info"} />
        <MetricCard title="待审核单集" value={String(pendingReviews)} note="未审核内容不得发布" tone={pendingReviews > 0 ? "warning" : "success"} />
        <MetricCard title="发布状态" value={publicationStateLabel[program.publicationState]} note="RSS 发布前实时校验权限" />
      </section>
      <section className="rounded-lg border border-border bg-surface p-5">
        <h2 className="text-lg font-semibold">下一步建议</h2>
        <p className="mt-2 text-secondary">
          {program.rightsState !== "clear" ? "先补充权利说明，避免未授权内容进入 RSS。" : pendingReviews > 0 ? "处理待审核单集，再检查发布范围。" : programSources.length === 0 ? "为节目添加来源后才能导入内容。" : "当前节目链路清晰，可继续观察任务和发布状态。"}
        </p>
      </section>
      <div className="grid gap-6 xl:grid-cols-[1fr_1fr]">
        <Panel title="来源">
          {programSources.length === 0 ? <EmptyState title="这个节目还没有来源" /> : (
            <div className="grid gap-3">
              {programSources.map((source) => (
                <article key={source.id} className="rounded-lg border border-border p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div>
                      <h3 className="font-semibold">{source.name}</h3>
                      <p className="mt-1 text-sm text-secondary">{ingestionTypeLabel[source.ingestionType]} / {triggerTypeLabel[source.triggerType]} / {authModeLabel[source.authMode]}</p>
                    </div>
                    <Badge tone={source.status === "ready" ? "success" : "warning"}>{sourceStatusLabel[source.status]}</Badge>
                  </div>
                </article>
              ))}
            </div>
          )}
        </Panel>
        <Panel title="最近单集">
          {programEpisodes.length === 0 ? <EmptyState title="这个节目还没有单集" /> : (
            <div className="grid gap-3">
              {programEpisodes.slice(0, 3).map((episode) => <EpisodeRow key={episode.id} episode={episode} />)}
            </div>
          )}
        </Panel>
        <Panel title="访问与发布">
          <div className="grid gap-4">
            <Input label="发布备注" defaultValue="Mock 编辑：发布前确认权利与授权范围。" />
            <div className="rounded-lg border border-border bg-subtle p-4 text-sm text-secondary">
              RSS 请求将实时校验 Token、用户状态、节目访问权和发布状态，缓存不得绕过授权检查。
            </div>
          </div>
        </Panel>
        <Panel title="活动记录">
          <Timeline items={latestJob?.timeline ?? [{ label: "暂无活动", at: "M0.2B", tone: "neutral" }]} />
        </Panel>
      </div>
    </div>
  );
}

function Panel({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
      <div className="mb-4 flex items-center gap-2">
        <CheckCircle2 className="h-4 w-4 text-action" />
        <h2 className="text-lg font-semibold">{title}</h2>
      </div>
      {children}
    </section>
  );
}
