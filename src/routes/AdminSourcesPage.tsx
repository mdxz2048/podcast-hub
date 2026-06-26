import { Link } from "react-router-dom";
import { useEffect, useState } from "react";
import { Plus } from "lucide-react";
import { listSources } from "../api/sources";
import type { ConnectorSource } from "../api/sources";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState } from "../components/StateBlocks";

export function AdminSourcesPage() {
  const [sources, setSources] = useState<ConnectorSource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    listSources()
      .then((result) => {
        if (!cancelled) setSources(result.sources);
      })
      .catch(() => {
        if (!cancelled) setError("无法加载 Source 列表。");
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
      <PageHeader eyebrow="Connector Source" title="来源配置实例">
        <Link to="/admin/sources/new"><Button icon={<Plus className="h-4 w-4" />}>新建 Source</Button></Link>
      </PageHeader>
      {loading ? <LoadingState title="正在加载 Source" /> : null}
      {error ? <ErrorState title={error} /> : null}
      {!loading && !error && sources.length === 0 ? <EmptyState title="暂无 Source" /> : null}
      <div className="grid gap-3">
        {sources.map((source) => (
          <Link key={source.id} to={`/admin/sources/${source.id}`} className="rounded-lg border border-border bg-surface p-5 shadow-subtle hover:border-strong">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 className="font-semibold">{source.name}</h2>
                <p className="mt-1 text-sm text-secondary">{source.description || "未填写描述"}</p>
              </div>
              <Badge>{source.status}</Badge>
            </div>
            <div className="mt-4 grid gap-2 text-sm text-secondary md:grid-cols-4">
              <span>Trigger: {source.trigger_type}</span>
              <span>Auth: {source.auth_mode}</span>
              <span>Execution: {source.execution_mode}</span>
              <span>Network: {source.network_mode}</span>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
