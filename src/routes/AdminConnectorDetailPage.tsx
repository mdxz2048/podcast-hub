import { ArrowLeft, Ban, CheckCircle2 } from "lucide-react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { KeyValue, MetricCard } from "../components/AdminPrimitives";
import { SuccessFeedback } from "../components/StateBlocks";
import { connectors, jobs, sources } from "../mock/data";
import { authModeLabel, connectorKindLabel, connectorStatusLabel, executionModeLabel, ingestionTypeLabel, jobStatusLabel, triggerTypeLabel } from "../utils/labels";

export function AdminConnectorDetailPage() {
  const { connectorId } = useParams();
  const [params] = useSearchParams();
  const connector = connectors.find((item) => item.id === connectorId);
  if (!connector) return <div>Connector 不存在</div>;
  const boundSources = sources.filter((source) => connector.boundSourceIds.includes(source.id));
  const relatedJobs = jobs.filter((job) => job.connectorId === connector.id);
  const disabled = params.get("state") === "disabled";

  return (
    <div className="grid gap-6">
      <Link to="/admin/connectors" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回 Connector 列表
      </Link>
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <div className="flex flex-wrap gap-2">
              <Badge tone={disabled ? "danger" : connector.status === "approved" || connector.status === "native_builtin" ? "success" : "warning"}>{disabled ? "已禁用" : connectorStatusLabel[connector.status]}</Badge>
              <Badge>{connectorKindLabel[connector.kind]}</Badge>
            </div>
            <h1 className="mt-3 text-3xl font-semibold leading-tight md:text-4xl">{connector.name}</h1>
            <p className="mt-2 text-secondary">版本 {connector.version} · {connector.nextAction}</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button icon={<CheckCircle2 className="h-4 w-4" />}>启用 Mock</Button>
            <Button variant="danger" icon={<Ban className="h-4 w-4" />}>禁用 Mock</Button>
          </div>
        </div>
      </section>
      {params.get("toast") === "success" ? <SuccessFeedback message="Connector 操作已写入模拟状态，未上传、解压或执行任何 ZIP。" /> : null}
      <section className="grid gap-4 md:grid-cols-3">
        <MetricCard title="绑定来源" value={String(boundSources.length)} note="Source 通过 connectorId 绑定" />
        <MetricCard title="最近任务" value={jobStatusLabel[connector.lastJobStatus]} note="任务结果不会直接发布 RSS" />
        <MetricCard title="资源限制" value={`${connector.resourceLimits.memoryMb} MB`} note={`${connector.resourceLimits.timeoutSeconds}s / ${connector.resourceLimits.maxDownloadMb} MB`} />
      </section>
      <div className="grid gap-6 xl:grid-cols-[1fr_1fr]">
        <section className="rounded-lg border border-border bg-surface p-5">
          <h2 className="mb-4 text-lg font-semibold">Manifest 摘要</h2>
          <dl className="grid gap-4 sm:grid-cols-2">
            <KeyValue label="ingestion_type" value={connector.supportedIngestionTypes.map((item) => ingestionTypeLabel[item]).join(" / ")} />
            <KeyValue label="trigger_type" value={connector.supportedTriggerTypes.map((item) => triggerTypeLabel[item]).join(" / ")} />
            <KeyValue label="auth_mode" value={connector.authModes.map((item) => authModeLabel[item]).join(" / ")} />
            <KeyValue label="execution_mode" value={connector.executionModes.map((item) => executionModeLabel[item]).join(" / ")} />
            <KeyValue label="入口文件" value={connector.entrypoint} />
            <KeyValue label="依赖锁" value={connector.dependencyLock} />
          </dl>
          <div className="mt-4 rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
            Connector 始终按不可信代码处理。M0.2B 不读取、上传、解压、校验或执行真实 ZIP。
          </div>
        </section>
        <section className="rounded-lg border border-border bg-surface p-5">
          <h2 className="mb-4 text-lg font-semibold">网络和绑定</h2>
          <div className="grid gap-3">
            <Info label="申请域名白名单" value={connector.networkPolicy.length > 0 ? connector.networkPolicy.join(", ") : "不申请外部网络"} />
            <Info label="绑定来源" value={boundSources.map((source) => source.name).join("、") || "暂无绑定"} />
            <Info label="最近任务" value={relatedJobs.map((job) => `${job.id}: ${jobStatusLabel[job.status]}`).join("；") || "暂无任务"} />
          </div>
        </section>
        <section className="rounded-lg border border-border bg-surface p-5 xl:col-span-2">
          <h2 className="mb-4 text-lg font-semibold">版本历史</h2>
          <div className="grid gap-3 md:grid-cols-2">
            {connector.versionHistory.map((version) => (
              <div key={version.version} className="rounded-lg border border-border p-4">
                <div className="flex items-center justify-between gap-3">
                  <span className="font-semibold">{version.version}</span>
                  <Badge>{connectorStatusLabel[version.status]}</Badge>
                </div>
                <p className="mt-2 text-sm text-muted">{version.date}</p>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-border bg-subtle p-4">
      <p className="text-xs text-muted">{label}</p>
      <p className="mt-1 text-sm font-medium">{value}</p>
    </div>
  );
}
