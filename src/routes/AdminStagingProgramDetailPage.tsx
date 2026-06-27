import { useEffect, useState } from "react";
import { ArrowLeft } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { getStagingProgram } from "../api/staging";
import type { StagingProgram } from "../api/staging";
import { Badge } from "../components/Badge";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState } from "../components/StateBlocks";

export function AdminStagingProgramDetailPage() {
  const { programId = "" } = useParams();
  const [program, setProgram] = useState<StagingProgram | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getStagingProgram(programId)
      .then((result) => setProgram(result.program))
      .catch(() => setError("无法加载 Program Candidate。"))
      .finally(() => setLoading(false));
  }, [programId]);

  if (loading) return <LoadingState title="正在加载 Program Candidate" />;
  if (error) return <ErrorState title={error} />;
  if (!program) return <EmptyState title="Program Candidate 不存在" />;

  return (
    <div className="grid gap-6">
      <Link to="/admin/staging" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回待审核区
      </Link>
      <PageHeader eyebrow="Program Candidate" title={program.title}>
        <Badge tone="warning">{program.status}</Badge>
      </PageHeader>
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <dl className="grid gap-3 text-sm text-secondary md:grid-cols-2">
          <Info label="Author" value={program.author} />
          <Info label="Language" value={program.language} />
          <Info label="Source" value={program.created_from_source_id} />
          <Info label="Import Job" value={program.created_from_job_id} />
          <Info label="Created" value={formatDate(program.created_at)} />
          <Info label="Updated" value={formatDate(program.updated_at)} />
        </dl>
        <p className="mt-5 whitespace-pre-wrap text-secondary">{program.description}</p>
      </section>
      <section className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
        <p className="font-medium text-primary">审核前状态</p>
        <p className="mt-1">该 Program 仍在 staging/review_pending 范围内，不提供发布、RSS、订阅或用户可见入口。</p>
      </section>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs uppercase text-muted">{label}</dt>
      <dd className="mt-1 break-words font-medium text-primary">{value}</dd>
    </div>
  );
}

function formatDate(value?: string) {
  if (!value) return "not set";
  return new Date(value).toLocaleString();
}
