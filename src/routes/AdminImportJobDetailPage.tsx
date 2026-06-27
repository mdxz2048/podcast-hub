import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft, Ban, PackageCheck } from "lucide-react";
import type { ApiError } from "../api/client";
import { cancelImportJob, getImportJob, listImportJobArtifacts, listImportJobEvents } from "../api/jobs";
import type { ImportJob, ImportJobArtifact, ImportJobEvent } from "../api/jobs";
import { canRunIntake, getImportJobIntakeStatus, runImportJobIntake } from "../api/staging";
import type { IntakeRun } from "../api/staging";
import { getAdminSystemStatus } from "../api/system";
import type { RunnerStatus } from "../api/system";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminImportJobDetailPage() {
  const { jobId = "" } = useParams();
  const [job, setJob] = useState<ImportJob | null>(null);
  const [events, setEvents] = useState<ImportJobEvent[]>([]);
  const [artifacts, setArtifacts] = useState<ImportJobArtifact[]>([]);
  const [intakeRun, setIntakeRun] = useState<IntakeRun | null>(null);
  const [validationIssues, setValidationIssues] = useState<string[]>([]);
  const [runner, setRunner] = useState<RunnerStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function reload() {
    const [jobResult, eventResult, artifactResult, statusResult, intakeResult] = await Promise.all([getImportJob(jobId), listImportJobEvents(jobId), listImportJobArtifacts(jobId), getAdminSystemStatus(), getImportJobIntakeStatus(jobId)]);
    setJob(jobResult.job);
    setEvents(eventResult.events);
    setArtifacts(artifactResult.artifacts);
    setRunner(statusResult.runner ?? null);
    setIntakeRun(intakeResult.intake_run);
  }

  useEffect(() => {
    reload().catch(() => setError("无法加载 Import Job。")).finally(() => setLoading(false));
  }, [jobId]);

  async function cancel() {
    const result = await cancelImportJob(jobId);
    setJob(result.job);
    setSuccess("取消请求已记录。");
  }

  async function runIntake() {
    setBusy(true);
    setError(null);
    setValidationIssues([]);
    try {
      const result = await runImportJobIntake(jobId);
      setIntakeRun(result.intake_run);
      setValidationIssues(result.validation_issues ?? []);
      setSuccess(result.program ? `已导入待审核区：${result.program.title}` : "已导入待审核区。");
    } catch (err) {
      const apiError = err as ApiError;
      setValidationIssues(apiError.validation_issues ?? []);
      setError(apiError.message);
    } finally {
      setBusy(false);
    }
  }

  if (loading) return <LoadingState title="正在加载 Import Job" />;
  if (error && !job) return <ErrorState title={error} />;
  if (!job) return <EmptyState title="任务不存在" />;

  const hasMetadataBundle = artifacts.some((artifact) => artifact.artifact_type === "metadata_bundle");
  const intakeGate = canRunIntake(job, hasMetadataBundle, intakeRun);

  return (
    <div className="grid gap-6">
      <Link to="/admin/import-jobs" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回任务列表
      </Link>
      <PageHeader eyebrow="Import Job" title={job.id}>
        <Button variant="secondary" icon={<PackageCheck className="h-4 w-4" />} disabled={!intakeGate.allowed || busy} onClick={runIntake}>
          导入到待审核区
        </Button>
        {(job.status === "queued" || job.status === "running") ? <Button variant="danger" icon={<Ban className="h-4 w-4" />} onClick={cancel}>取消任务</Button> : null}
      </PageHeader>
      {success ? <SuccessFeedback message={success} /> : null}
      {error ? <ErrorState title={error} /> : null}
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
          <div>
            <h2 className="text-lg font-semibold">待审核区导入</h2>
            <p className="mt-2 text-sm text-secondary">{intakeGate.reason}</p>
            {intakeRun ? (
              <p className="mt-2 text-sm text-secondary">
                Intake: <span className="font-medium text-primary">{intakeRun.status}</span> · {formatDate(intakeRun.updated_at)}
              </p>
            ) : null}
          </div>
          <Badge tone={intakeGate.allowed ? "success" : intakeRun?.status === "failed" ? "danger" : "warning"}>{intakeRun?.status ?? "not_started"}</Badge>
        </div>
        {validationIssues.length > 0 ? (
          <div className="mt-4 rounded-md border border-warning/30 bg-warning/5 p-3 text-sm text-secondary">
            <p className="font-medium text-primary">Validation issues</p>
            <ul className="mt-2 list-disc space-y-1 pl-5">
              {validationIssues.map((issue) => <li key={issue}>{issue}</li>)}
            </ul>
          </div>
        ) : null}
      </section>
      {runner && !runner.can_run_jobs && (job.status === "queued" || job.status === "running") ? (
        <section className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
          <p className="font-medium text-primary">Runner disabled</p>
          <p className="mt-1">{runner.reason}</p>
        </section>
      ) : null}
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
          <span>Created: {formatDate(job.created_at)}</span>
          <span>Started: {formatDate(job.started_at)}</span>
          <span>Finished: {formatDate(job.finished_at)}</span>
          <span>Cancel requested: {formatDate(job.cancellation_requested_at)}</span>
          <span>Failure code: {job.failure_code || "无"}</span>
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
                <p className="mt-2 break-words font-mono text-xs text-muted">{event.metadata_redacted}</p>
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
              <div key={artifact.id} className="grid gap-1 rounded-md border border-border p-3 text-sm text-secondary md:grid-cols-[1fr_1fr]">
                <span>Type: {artifact.artifact_type}</span>
                <span>Path: {artifact.relative_path}</span>
                <span>Size: {artifact.size_bytes} bytes</span>
                <span className="break-words font-mono text-xs">SHA-256: {artifact.sha256}</span>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

function formatDate(value?: string) {
  if (!value) return "not set";
  return new Date(value).toLocaleString();
}
