import { Clock3 } from "lucide-react";
import type { ImportJob } from "../types/domain";
import { authModeLabel, errorCategoryLabel, executionModeLabel, ingestionTypeLabel, jobStatusLabel, triggerTypeLabel } from "../utils/labels";
import { Badge } from "./Badge";

function jobTone(status: ImportJob["status"]) {
  if (status === "completed") return "success" as const;
  if (status === "failed" || status === "timed_out" || status === "cancelled") return "danger" as const;
  if (status === "waiting_auth" || status === "waiting_for_auth" || status === "waiting_manual_upload" || status === "completed_with_warnings") return "warning" as const;
  return "info" as const;
}

export function ImportJobCard({ job }: { job: ImportJob }) {
  return (
    <article className="rounded-lg border border-border bg-surface p-4 shadow-subtle">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="font-mono text-xs text-muted">{job.id}</p>
          <h3 className="mt-1 font-semibold">{job.nextAction}</h3>
        </div>
        <Badge tone={jobTone(job.status)}>{jobStatusLabel[job.status]}</Badge>
      </div>
      <div className="mt-4 grid gap-2 text-sm text-secondary sm:grid-cols-2">
        <span>{ingestionTypeLabel[job.ingestionType]} / {triggerTypeLabel[job.triggerType]}</span>
        <span>{authModeLabel[job.authMode]} / {executionModeLabel[job.executionMode]}</span>
        <span className="flex items-center gap-2"><Clock3 className="h-4 w-4" />{job.startedAt}</span>
        <span>{job.errorCategory ? errorCategoryLabel[job.errorCategory] ?? job.errorCategory : "无错误"}</span>
      </div>
    </article>
  );
}
