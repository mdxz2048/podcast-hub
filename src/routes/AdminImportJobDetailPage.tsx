import { ArrowLeft, Ban, RotateCcw } from "lucide-react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Dialog } from "../components/Dialog";
import { EpisodeRow } from "../components/EpisodeRow";
import { KeyValue, ProgressBar, StaticQrPlaceholder, Timeline } from "../components/AdminPrimitives";
import { EmptyState, SuccessFeedback } from "../components/StateBlocks";
import { connectors, episodes, jobs, programs, sources } from "../mock/data";
import { useMockState } from "../mock/MockState";
import { authModeLabel, errorCategoryLabel, executionModeLabel, ingestionTypeLabel, jobStatusLabel, triggerTypeLabel } from "../utils/labels";

export function AdminImportJobDetailPage() {
  const { jobId } = useParams();
  const [params] = useSearchParams();
  const { showToast } = useMockState();
  const dialogOpen = params.get("dialog") === "cancel";
  const forcedState = params.get("state");
  const baseJob = jobs.find((item) => item.id === jobId);
  if (!baseJob) return <EmptyState title="任务不存在" />;
  const job = forcedState ? { ...baseJob, status: forcedState as typeof baseJob.status } : baseJob;
  const program = programs.find((item) => item.id === job.programId);
  const source = sources.find((item) => item.id === job.sourceId);
  const connector = connectors.find((item) => item.id === job.connectorId);
  const outputEpisodes = (job.outputEpisodeIds ?? []).map((id) => episodes.find((episode) => episode.id === id)).filter(Boolean);
  const showQr = job.status === "waiting_auth" && job.authMode === "qr_each_run";

  function mockAction(label: string) {
    showToast({ tone: "success", title: label, message: "这是静态 Mock 操作，没有调用真实任务系统。" });
  }

  return (
    <div className="grid gap-6">
      <Link to="/admin/import-jobs" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回任务列表
      </Link>
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <div className="flex flex-wrap gap-2">
              <Badge tone={job.status === "failed" || job.status === "cancelled" ? "danger" : job.status === "completed" ? "success" : "warning"}>{jobStatusLabel[job.status]}</Badge>
              <Badge>{ingestionTypeLabel[job.ingestionType]}</Badge>
            </div>
            <h1 className="mt-3 text-3xl font-semibold">{job.id}</h1>
            <p className="mt-2 text-secondary">{job.nextAction}</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button icon={<RotateCcw className="h-4 w-4" />} onClick={() => mockAction("已触发重试")}>重试 Mock</Button>
            <Link to={`/admin/import-jobs/${job.id}?dialog=cancel`}>
              <Button variant="danger" icon={<Ban className="h-4 w-4" />}>取消任务</Button>
            </Link>
          </div>
        </div>
      </section>
      {params.get("toast") === "success" ? <SuccessFeedback message="Mock 操作已成功反馈，没有访问真实任务系统。" /> : null}
      <section className="grid gap-4 md:grid-cols-4">
        <Info label="节目" value={program?.title ?? "未知节目"} />
        <Info label="来源" value={source?.name ?? "未知来源"} />
        <Info label="Connector" value={connector ? `${connector.name} ${job.connectorVersion ?? ""}` : "无 Connector"} />
        <Info label="触发/授权" value={`${triggerTypeLabel[job.triggerType]} / ${authModeLabel[job.authMode]}`} />
      </section>
      <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr]">
        <section className="rounded-lg border border-border bg-surface p-5">
          <h2 className="mb-4 text-lg font-semibold">任务状态</h2>
          <ProgressBar value={job.progress ?? 0} />
          <dl className="mt-5 grid gap-4 sm:grid-cols-2">
            <KeyValue label="execution_mode" value={executionModeLabel[job.executionMode]} />
            <KeyValue label="开始时间" value={job.startedAt} />
            <KeyValue label="结束时间" value={job.finishedAt ?? "尚未结束"} />
            <KeyValue label="失败原因" value={job.errorCategory ? errorCategoryLabel[job.errorCategory] ?? job.errorCategory : "无"} />
          </dl>
          {showQr ? <div className="mt-5"><StaticQrPlaceholder /></div> : null}
          {job.status === "waiting_auth" && !showQr ? (
            <div className="mt-5 rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">可复用会话已失效，不能无限重试。请由管理员重新授权。</div>
          ) : null}
        </section>
        <section className="rounded-lg border border-border bg-surface p-5">
          <h2 className="mb-4 text-lg font-semibold">状态时间线</h2>
          <Timeline items={job.timeline ?? []} />
        </section>
        <section className="rounded-lg border border-border bg-surface p-5">
          <h2 className="mb-4 text-lg font-semibold">脱敏日志</h2>
          <pre className="max-h-72 overflow-auto rounded-md bg-subtle p-4 text-xs text-secondary">{(job.logEvents ?? ["暂无日志"]).join("\n")}</pre>
        </section>
        <section className="rounded-lg border border-border bg-surface p-5">
          <h2 className="mb-4 text-lg font-semibold">输出单集摘要</h2>
          {outputEpisodes.length === 0 ? <EmptyState title="暂无输出单集" /> : (
            <div className="grid gap-3">
              {outputEpisodes.map((episode) => episode ? <EpisodeRow key={episode.id} episode={episode} /> : null)}
            </div>
          )}
        </section>
      </div>
      <Dialog
        open={dialogOpen}
        title="取消导入任务"
        description={`确认取消 ${job.id}？取消后任务仍会保留审计记录，已产生的原始失败媒体默认不保留。`}
        confirmLabel="确认取消"
        onCancel={() => window.history.back()}
        onConfirm={() => mockAction("任务已取消")}
      />
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-surface p-4 shadow-subtle">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-1 text-sm font-medium">{value}</p>
    </div>
  );
}
