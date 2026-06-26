import { ArrowLeft, Ban, CheckCircle2, XCircle } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  approveConnectorVersion,
  disableConnector,
  disableConnectorVersion,
  enableConnector,
  getConnector,
  getConnectorVersion,
  listConnectorVersions,
  rejectConnectorVersion
} from "../api/connectors";
import type { AdminConnector, AdminConnectorVersion, ValidationSummary } from "../api/connectors";
import type { ApiError } from "../api/client";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminConnectorDetailPage() {
  const { connectorId } = useParams();
  const [connector, setConnector] = useState<AdminConnector | null>(null);
  const [versions, setVersions] = useState<AdminConnectorVersion[]>([]);
  const [selectedVersion, setSelectedVersion] = useState<AdminConnectorVersion | null>(null);
  const [selectedValidation, setSelectedValidation] = useState<ValidationSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function load() {
    if (!connectorId) return;
    setIsLoading(true);
    setError(null);
    try {
      const [connectorResult, versionsResult] = await Promise.all([
        getConnector(connectorId),
        listConnectorVersions(connectorId)
      ]);
      setConnector(connectorResult.connector);
      setVersions(versionsResult.versions);
      if (versionsResult.versions.length > 0) {
        const first = versionsResult.versions[0];
        const detail = await getConnectorVersion(first.id);
        setSelectedVersion(detail.version);
        setSelectedValidation(detail.validation_summary);
      }
    } catch (e) {
      const apiError = e as ApiError;
      setError(apiError.message || "加载 Connector 详情失败。");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [connectorId]);

  async function selectVersion(version: AdminConnectorVersion) {
    try {
      const detail = await getConnectorVersion(version.id);
      setSelectedVersion(detail.version);
      setSelectedValidation(detail.validation_summary);
    } catch {
      setSelectedVersion(version);
      setSelectedValidation(null);
    }
  }

  async function runAction(action: "enable_connector" | "disable_connector" | "approve" | "reject" | "disable_version") {
    if (!connector || !selectedVersion) return;
    setError(null);
    setSuccess(null);
    try {
      if (action === "enable_connector") {
        await enableConnector(connector.id);
      } else if (action === "disable_connector") {
        await disableConnector(connector.id);
      } else if (action === "approve") {
        await approveConnectorVersion(selectedVersion.id);
      } else if (action === "reject") {
        await rejectConnectorVersion(selectedVersion.id);
      } else {
        await disableConnectorVersion(selectedVersion.id);
      }
      setSuccess("状态更新成功。");
      await load();
    } catch (e) {
      const apiError = e as ApiError;
      setError(apiError.message || "状态更新失败。");
    }
  }

  return (
    <div className="grid gap-6">
      <Link to="/admin/connectors" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回 Connector 列表
      </Link>
      {isLoading ? <LoadingState title="正在加载 Connector 详情" /> : null}
      {error ? <ErrorState title={error} /> : null}
      {success ? <SuccessFeedback message={success} /> : null}
      {connector && !isLoading ? (
        <>
          <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div>
                <div className="flex flex-wrap gap-2">
                  <Badge tone={connector.status === "active" ? "success" : "warning"}>{connector.status === "active" ? "已启用" : "已禁用"}</Badge>
                  <Badge>{connector.slug}</Badge>
                </div>
                <h1 className="mt-3 text-3xl font-semibold leading-tight md:text-4xl">{connector.name}</h1>
                <p className="mt-2 text-secondary">{connector.description || "暂无描述"}</p>
              </div>
              <div className="flex flex-wrap gap-2">
                <Button icon={<CheckCircle2 className="h-4 w-4" />} onClick={() => void runAction("enable_connector")}>启用 Connector</Button>
                <Button variant="danger" icon={<Ban className="h-4 w-4" />} onClick={() => void runAction("disable_connector")}>禁用 Connector</Button>
              </div>
            </div>
            <div className="mt-4 rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
              本页仅展示登记与审核状态，不提供执行、下载媒体、创建 Source 或运行任务按钮。
            </div>
          </section>
          <section className="grid gap-6 xl:grid-cols-[1fr_1fr]">
            <div className="rounded-lg border border-border bg-surface p-5">
              <h2 className="mb-4 text-lg font-semibold">版本列表</h2>
              <div className="grid gap-3">
                {versions.map((version) => (
                  <button key={version.id} className={`rounded-lg border p-4 text-left ${selectedVersion?.id === version.id ? "border-action bg-subtle" : "border-border bg-surface"}`} onClick={() => void selectVersion(version)}>
                    <div className="flex items-center justify-between gap-3">
                      <span className="font-semibold">{version.version}</span>
                      <Badge>{version.review_status}</Badge>
                    </div>
                    <p className="mt-1 text-xs text-muted">runtime={version.runtime_profile} · {version.entrypoint}</p>
                  </button>
                ))}
              </div>
            </div>
            <div className="rounded-lg border border-border bg-surface p-5">
              <h2 className="mb-4 text-lg font-semibold">版本详情</h2>
              {selectedVersion ? (
                <div className="grid gap-4">
                  <div className="flex flex-wrap gap-2">
                    <Badge>{selectedVersion.review_status}</Badge>
                    <Badge>{selectedVersion.runtime_profile}</Badge>
                  </div>
                  <div className="grid gap-2 text-sm">
                    <div>Entrypoint: <code>{selectedVersion.entrypoint}</code></div>
                    <div>SHA256: <code>{selectedVersion.package_sha256}</code></div>
                    <div>包大小: {selectedVersion.package_size_bytes} bytes</div>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Button onClick={() => void runAction("approve")}>Approve</Button>
                    <Button variant="secondary" icon={<XCircle className="h-4 w-4" />} onClick={() => void runAction("reject")}>Reject</Button>
                    <Button variant="danger" icon={<Ban className="h-4 w-4" />} onClick={() => void runAction("disable_version")}>Disable 版本</Button>
                  </div>
                  <div className="rounded-lg border border-border bg-subtle p-4">
                    <p className="text-sm font-semibold">Manifest 摘要</p>
                    <pre className="mt-2 max-h-64 overflow-auto text-xs text-secondary">{JSON.stringify(selectedVersion.manifest, null, 2)}</pre>
                  </div>
                  <div className="rounded-lg border border-border bg-subtle p-4">
                    <p className="text-sm font-semibold">Validation Summary</p>
                    {selectedValidation ? (
                      <ul className="mt-2 list-disc space-y-1 pl-5 text-xs text-secondary">
                        {selectedValidation.issues.length === 0 ? <li>无校验问题</li> : selectedValidation.issues.map((issue) => (
                          <li key={`${issue.code}-${issue.path ?? ""}`}>{issue.message}{issue.path ? `（${issue.path}）` : ""}</li>
                        ))}
                      </ul>
                    ) : <p className="mt-2 text-xs text-secondary">暂无详细校验信息。</p>}
                  </div>
                </div>
              ) : (
                <p className="text-sm text-secondary">暂无版本。</p>
              )}
            </div>
          </section>
        </>
      ) : null}
    </div>
  );
}
