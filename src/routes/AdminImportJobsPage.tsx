import { Link } from "react-router-dom";
import { useEffect, useState } from "react";
import { Badge } from "../components/Badge";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState } from "../components/StateBlocks";
import { listImportJobs } from "../api/jobs";
import type { ImportJob } from "../api/jobs";
import { getAdminSystemStatus } from "../api/system";
import type { RunnerStatus } from "../api/system";

export function AdminImportJobsPage() {
  const [jobs, setJobs] = useState<ImportJob[]>([]);
  const [runner, setRunner] = useState<RunnerStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    Promise.all([listImportJobs(), getAdminSystemStatus()])
      .then(([jobResult, statusResult]) => {
        if (!cancelled) {
          setJobs(jobResult.jobs);
          setRunner(statusResult.runner ?? null);
        }
      })
      .catch(() => {
        if (!cancelled) setError("无法加载 Import Job 列表。");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="Import Jobs" title="手动导入任务" />
      {runner && !runner.can_run_jobs ? (
        <section className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
          <p className="font-medium text-primary">Runner disabled</p>
          <p className="mt-1">{runner.reason}</p>
        </section>
      ) : null}
      {loading ? <LoadingState title="正在加载 Import Job" /> : null}
      {error ? <ErrorState title={error} /> : null}
      {!loading && !error && jobs.length === 0 ? <EmptyState title="暂无 Import Job" /> : null}
      <div className="grid gap-3">
        {jobs.map((job) => (
          <Link key={job.id} to={`/admin/import-jobs/${job.id}`} className="rounded-lg border border-border bg-surface p-5 shadow-subtle hover:border-strong">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 className="font-semibold">{job.id}</h2>
                <p className="mt-1 text-sm text-secondary">Source {job.connector_source_id}</p>
              </div>
              <Badge tone={job.status === "failed" || job.status === "cancelled" ? "danger" : job.status === "completed" ? "success" : "warning"}>{job.status}</Badge>
            </div>
            <div className="mt-4 grid gap-2 text-sm text-secondary md:grid-cols-3">
              <span>Trigger: {job.trigger_type}</span>
              <span>Auth: {job.auth_mode}</span>
              <span>Execution: {job.execution_mode}</span>
              <span>Version: {job.connector_version_id}</span>
              <span>Created: {formatDate(job.created_at)}</span>
              <span>Started: {formatDate(job.started_at)}</span>
              <span>Finished: {formatDate(job.finished_at)}</span>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}

function formatDate(value?: string) {
  if (!value) return "not set";
  return new Date(value).toLocaleString();
}
