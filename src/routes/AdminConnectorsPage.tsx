import { Plus } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { listConnectors } from "../api/connectors";
import type { AdminConnector } from "../api/connectors";
import type { ApiError } from "../api/client";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { SearchBar } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState } from "../components/StateBlocks";

export function AdminConnectorsPage() {
  const [params] = useSearchParams();
  const [connectors, setConnectors] = useState<AdminConnector[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const query = params.get("q") ?? "";

  useEffect(() => {
    let active = true;
    (async () => {
      try {
        const result = await listConnectors();
        if (!active) return;
        setConnectors(result.connectors);
      } catch (e) {
        if (!active) return;
        const apiError = e as ApiError;
        setError(apiError.message || "加载 Connector 列表失败。");
      } finally {
        if (active) setIsLoading(false);
      }
    })();
    return () => {
      active = false;
    };
  }, []);

  const visible = useMemo(() => connectors.filter((connector) => `${connector.slug} ${connector.name}`.toLowerCase().includes(query.toLowerCase())), [connectors, query]);

  return (
    <div>
      <PageHeader eyebrow="Connector Registry" title="管理通用 Connector 上传版本与审核状态">
        <Link to="/admin/connectors/new">
          <Button icon={<Plus className="h-4 w-4" />}>上传 Connector ZIP</Button>
        </Link>
      </PageHeader>
      <div className="mb-5 rounded-lg border border-border bg-surface p-4">
        <SearchBar placeholder="搜索 Connector 名称或 slug" defaultValue={query} />
      </div>
      {isLoading ? <LoadingState title="正在加载 Connector 列表" /> : null}
      {error ? <ErrorState title={error} /> : null}
      {!isLoading && !error && visible.length === 0 ? <EmptyState title="暂无 Connector" /> : null}
      {!isLoading && !error ? (
        <div className="grid gap-4">
          {visible.map((connector) => (
            <article key={connector.id} className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
              <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                <div>
                  <div className="flex flex-wrap gap-2">
                    <Badge tone={connector.status === "active" ? "success" : "warning"}>{connector.status === "active" ? "已启用" : "已禁用"}</Badge>
                    <Badge>{connector.slug}</Badge>
                  </div>
                  <h2 className="mt-3 text-xl font-semibold">{connector.name}</h2>
                  <p className="mt-1 text-sm text-secondary">{connector.description || "暂无描述"}</p>
                </div>
                <Link to={`/admin/connectors/${connector.id}`}>
                  <Button variant="secondary">查看详情</Button>
                </Link>
              </div>
            </article>
          ))}
        </div>
      ) : null}
    </div>
  );
}
