import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listAdminPrograms } from "../api/adminContent";
import type { AdminProgram } from "../api/adminContent";
import { Badge } from "../components/Badge";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState } from "../components/StateBlocks";

export function AdminProgramsPage() {
  const [programs, setPrograms] = useState<AdminProgram[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listAdminPrograms()
      .then((result) => setPrograms(result.programs))
      .catch(() => setError("节目管理暂不可用"))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingState title="正在加载节目列表" />;
  if (error) return <ErrorState title={error} />;

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="节目" title="管理节目审核、发布和归档状态" />
      {programs.length === 0 ? <EmptyState title="暂无节目" /> : (
        <div className="overflow-hidden rounded-lg border border-border bg-surface shadow-subtle">
          <div className="hidden grid-cols-[1.4fr_140px_1fr_180px] gap-4 border-b border-border px-4 py-3 text-xs font-semibold uppercase text-muted lg:grid">
            <span>节目</span>
            <span>状态</span>
            <span>来源</span>
            <span>更新时间</span>
          </div>
          {programs.map((program) => (
            <article key={program.id} className="grid gap-3 border-b border-border px-4 py-4 last:border-b-0 lg:grid-cols-[1.4fr_140px_1fr_180px] lg:items-center">
              <div className="min-w-0">
                <Link to={`/admin/programs/${program.id}`} className="break-words font-semibold leading-tight hover:text-action">{program.title}</Link>
                <p className="mt-1 line-clamp-2 text-sm text-secondary">{program.description}</p>
              </div>
              <Badge tone={program.status === "published" ? "success" : program.status === "rejected" || program.status === "archived" ? "danger" : "warning"}>{program.status}</Badge>
              <span className="break-words text-sm text-secondary">{program.created_from_source_id}</span>
              <span className="text-sm text-secondary">{formatDate(program.updated_at)}</span>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}

function formatDate(value?: string) {
  if (!value) return "not set";
  return new Date(value).toLocaleString();
}
