import { Plus, ShieldCheck } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { SearchBar } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { connectors, sources } from "../mock/data";
import { authModeLabel, connectorKindLabel, connectorStatusLabel, jobStatusLabel, triggerTypeLabel } from "../utils/labels";

export function AdminConnectorsPage() {
  const [params] = useSearchParams();
  const query = params.get("q") ?? "";
  const visible = connectors.filter((connector) => `${connector.name} ${connector.kind}`.toLowerCase().includes(query.toLowerCase()));

  return (
    <div>
      <PageHeader eyebrow="Connector Registry" title="管理内容接入能力和版本状态">
        <Link to="/admin/connectors/new">
          <Button icon={<Plus className="h-4 w-4" />}>登记 Connector</Button>
        </Link>
      </PageHeader>
      <div className="mb-5 rounded-lg border border-border bg-surface p-4">
        <SearchBar placeholder="搜索 Connector 或接入方式" defaultValue={query} />
      </div>
      <div className="grid gap-4">
        {visible.map((connector) => {
          const boundSources = sources.filter((source) => connector.boundSourceIds.includes(source.id));
          return (
            <article key={connector.id} className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
              <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap gap-2">
                    <Badge tone={connector.status === "approved" || connector.status === "native_builtin" ? "success" : "warning"}>{connectorStatusLabel[connector.status]}</Badge>
                    <Badge>{connectorKindLabel[connector.kind]}</Badge>
                  </div>
                  <h2 className="mt-3 text-xl font-semibold">{connector.name}</h2>
                  <p className="mt-2 text-sm text-secondary">版本 {connector.version} · 最近任务 {jobStatusLabel[connector.lastJobStatus]} · {connector.nextAction}</p>
                </div>
                <Link to={`/admin/connectors/${connector.id}`}>
                  <Button variant="secondary" icon={<ShieldCheck className="h-4 w-4" />}>查看详情</Button>
                </Link>
              </div>
              <div className="mt-4 grid gap-3 md:grid-cols-3">
                <Info label="支持触发" value={connector.supportedTriggerTypes.map((item) => triggerTypeLabel[item]).join(" / ")} />
                <Info label="授权模式" value={connector.authModes.map((item) => authModeLabel[item]).join(" / ")} />
                <Info label="绑定来源" value={`${boundSources.length} 个`} />
              </div>
            </article>
          );
        })}
      </div>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-border bg-subtle p-3">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-1 text-sm font-medium">{value}</p>
    </div>
  );
}
