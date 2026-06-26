import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft, Ban } from "lucide-react";
import { cancelImportJob, getImportJob, listImportJobArtifacts, listImportJobEvents } from "../api/jobs";
import type { ImportJob, ImportJobArtifact, ImportJobEvent } from "../api/jobs";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminImportJobDetailPage() {
  const { jobId = "" } = useParams();
  const [job, setJob] = useState<ImportJob | null>(null);
  const [events, setEvents] = useState<ImportJobEvent[]>([]);
  const [artifacts, setArtifacts] = useState<ImportJobArtifact[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function reload() {
    const [jobResult, eventResult, artifactResult] = await Promise.all([getImportJob(jobId), listImportJobEvents(jobId), listImportJobArtifacts(jobId)]);
    setJob(jobResult.job);
    setEvents(eventResult.events);
    setArtifacts(artifactResult.artifacts);
  }

  useEffect(() => {
    reload().catch(() => setError("无法加载 Import Job。")).finally(() => setLoading(false));
  }, [jobId]);

  async function cancel() {
    const result = await cancelImportJob(jobId);
    setJob(result.job);
    setSuccess("取消请求已记录。");
  }

  if (loading) return <LoadingState title="正在加载 Import Job" />;
  if (error && !job) return <ErrorState title={error} />;
  if (!job) return <EmptyState title="任务不存在" />;

  return (
    <div className="grid gap-6">
      <Link to="/admin/import-jobs" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回任务列表
      </Link>
      <PageHeader eyebrow="Import Job" title={job.id}>
        {(job.status === "queued" || job.status === "running") ? <Button variant="danger" icon={<Ban className="h-4 w-4" />} onClick={cancel}>取消任务</Button> : null}
      </PageHeader>
      {success ? <SuccessFeedback message={success} /> : null}
      {error ? <ErrorState title={error} /> : null}
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <div className="flex flex-wrap gap-2">
          <Badge tone={job.status === "failed" || job.status === "cancelled" ? "danger" : job.status === "completed" ? "success" : "warning"}>{job.status}</Badge>
          <Badge>{job.trigger_type}</Badge>
          <Badge>{job.auth_mode}</Badge>
          <Badge>{job.execution_mode}</Badge>
        </div>
        <div className="mt-4 grid gap-2 text-sm text-secondary md:grid-cols-2">
          <span>Source: {job.connector_source_id}</span>
          <span>Version: {job.connector_version_id}</span>
          <span>Failure: {job.failure_message_redacted || "无"}</span>
          <span>Created: {job.created_at}</span>
        </div>
      </section>
      <section className="rounded-lg border border-border bg-surface p-5">
        <h2 className="mb-4 text-lg font-semibold">脱敏事件</h2>
        {events.length === 0 ? <EmptyState title="暂无事件" /> : (
          <div className="grid gap-2">
            {events.map((event) => (
              <div key={event.id} className="rounded-md border border-border p-3 text-sm">
                <Badge>{event.level}</Badge>
                <p className="mt-2 font-medium">{event.event_type}</p>
                <p className="text-secondary">{event.message_redacted}</p>
              </div>
            ))}
          </div>
        )}
      </section>
      <section className="rounded-lg border border-border bg-surface p-5">
        <h2 className="mb-4 text-lg font-semibold">Artifact 元数据</h2>
        {artifacts.length === 0 ? <EmptyState title="暂无 Artifact" /> : (
          <div className="grid gap-2">
            {artifacts.map((artifact) => (
              <div key={artifact.id} className="rounded-md border border-border p-3 text-sm text-secondary">
                {artifact.artifact_type} · {artifact.relative_path} · {artifact.size_bytes} bytes · {artifact.sha256}
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
