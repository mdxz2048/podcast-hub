import { AlertTriangle, CheckCircle2, Clock3, RadioTower } from "lucide-react";
import type { ReactNode } from "react";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { ImportJobCard } from "../components/ImportJobCard";
import { PageHeader, Section } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import { jobs, programs, sources } from "../mock/data";
import { authModeLabel, executionModeLabel, ingestionTypeLabel, sourceStatusLabel, triggerTypeLabel } from "../utils/labels";
import { useViewState } from "./state";

export function AdminOverviewPage() {
  const state = useViewState();
  if (state === "loading") return <LoadingState title="正在加载管理概览" />;
  if (state === "empty") return <EmptyState title="暂无运营活动" action="创建节目" />;
  if (state === "error") return <ErrorState title="管理概览暂不可用" />;
  if (state === "denied") return <PermissionDeniedState />;

  const waitingJobs = jobs.filter((job) => job.status === "waiting_for_auth").length;
  const rightsHold = programs.filter((program) => program.status === "rights_hold").length;

  return (
    <>
      <PageHeader eyebrow="管理概览" title="运营健康度与下一步操作">
        <Button>运行手动导入</Button>
        <Button variant="secondary">打开审核队列</Button>
      </PageHeader>
      {state === "success" ? <SuccessFeedback message="静态操作已进入队列，没有发起后端请求。" /> : null}
      <section className="mb-8 grid gap-4 md:grid-cols-4">
        <Metric title="节目" value={String(programs.length)} note="3 个活跃模拟视图" icon={<CheckCircle2 className="h-5 w-5" />} />
        <Metric title="来源" value={String(sources.length)} note="已应用四维来源模型" icon={<RadioTower className="h-5 w-5" />} />
        <Metric title="授权任务" value={String(waitingJobs)} note="需要运营处理" icon={<Clock3 className="h-5 w-5" />} tone="warning" />
        <Metric title="权利暂缓" value={String(rightsHold)} note="会阻止发布" icon={<AlertTriangle className="h-5 w-5" />} tone="danger" />
      </section>
      <div className="grid gap-8 xl:grid-cols-[1.1fr_0.9fr]">
        <Section title="优先任务队列">
          <div className="grid gap-4">
            {jobs.map((job) => (
              <ImportJobCard key={job.id} job={job} />
            ))}
          </div>
        </Section>
        <Section title="来源模型快照">
          <div className="grid gap-3">
            {sources.map((source) => (
              <div key={source.id} className="rounded-lg border border-border bg-surface p-4">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h3 className="font-semibold">{source.name}</h3>
                    <p className="mt-1 text-sm text-secondary">{ingestionTypeLabel[source.ingestionType]} / {triggerTypeLabel[source.triggerType]} / {authModeLabel[source.authMode]} / {executionModeLabel[source.executionMode]}</p>
                  </div>
                  <Badge tone={source.status === "ready" ? "success" : "warning"}>{sourceStatusLabel[source.status]}</Badge>
                </div>
              </div>
            ))}
          </div>
        </Section>
      </div>
    </>
  );
}

const metricTone = {
  info: "text-info",
  warning: "text-warning",
  danger: "text-danger"
};

function Metric({ title, value, note, icon, tone = "info" }: { title: string; value: string; note: string; icon: ReactNode; tone?: "info" | "warning" | "danger" }) {
  return (
    <div className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
      <div className={`mb-4 inline-flex rounded-md p-2 ${metricTone[tone]}`}>{icon}</div>
      <p className="text-sm text-secondary">{title}</p>
      <p className="mt-1 text-3xl font-semibold">{value}</p>
      <p className="mt-2 text-xs text-muted">{note}</p>
    </div>
  );
}
