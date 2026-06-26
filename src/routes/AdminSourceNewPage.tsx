import { FormEvent, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { listConnectors, listConnectorVersions } from "../api/connectors";
import type { AdminConnectorVersion } from "../api/connectors";
import { createSource, manifestSecrets } from "../api/sources";
import { Button } from "../components/Button";
import { Input, Select } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { ErrorState, LoadingState } from "../components/StateBlocks";

export function AdminSourceNewPage() {
  const navigate = useNavigate();
  const [versions, setVersions] = useState<AdminConnectorVersion[]>([]);
  const [connectorVersionID, setConnectorVersionID] = useState("");
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [authMode, setAuthMode] = useState<"none" | "reusable_session">("none");
  const [networkMode, setNetworkMode] = useState<"disabled" | "trusted_admin">("disabled");
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      const connectorResult = await listConnectors();
      const active = connectorResult.connectors.filter((connector) => connector.status === "active");
      const versionLists = await Promise.all(active.map((connector) => listConnectorVersions(connector.id)));
      const approved = versionLists.flatMap((result) => result.versions).filter((version) => version.review_status === "approved");
      if (!cancelled) {
        setVersions(approved);
        setConnectorVersionID(approved[0]?.id ?? "");
      }
    }
    load().catch(() => setError("无法加载 approved Connector Version。")).finally(() => setLoading(false));
    return () => {
      cancelled = true;
    };
  }, []);

  const selectedVersion = useMemo(() => versions.find((version) => version.id === connectorVersionID), [versions, connectorVersionID]);
  const requiredSecrets = selectedVersion ? manifestSecrets(selectedVersion) : [];

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      const result = await createSource({
        connector_version_id: connectorVersionID,
        name,
        description,
        trigger_type: "manual",
        auth_mode: authMode,
        execution_mode: "unattended",
        network_mode: networkMode,
        config: {}
      });
      navigate(`/admin/sources/${result.source.id}`);
    } catch {
      setError("Source 创建失败。请确认版本已 approved 且 Alpha 仅支持 manual + unattended。");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="新建 Source" title="基于 approved Connector Version 创建配置实例" />
      <section className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
        Source 只保存配置和 Secret Reference。不要把 Secret 写进 Connector ZIP 或 manifest。本阶段不创建 Program、Episode，也不执行 Connector。
      </section>
      {loading ? <LoadingState title="正在加载可用 Connector Version" /> : null}
      {error ? <ErrorState title={error} /> : null}
      <form className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle" onSubmit={submit}>
        <Select label="Approved Connector Version" value={connectorVersionID} onChange={(event) => setConnectorVersionID(event.target.value)} options={versions.map((version) => version.id)} />
        <Input label="Source 名称" value={name} onChange={(event) => setName(event.target.value)} />
        <Input label="描述" value={description} onChange={(event) => setDescription(event.target.value)} />
        <Select label="Auth Mode" value={authMode} onChange={(event) => setAuthMode(event.target.value as "none" | "reusable_session")} options={["none", "reusable_session"]} />
        <Select label="Network Mode" value={networkMode} onChange={(event) => setNetworkMode(event.target.value as "disabled" | "trusted_admin")} options={["disabled", "trusted_admin"]} />
        {requiredSecrets.length > 0 ? <p className="text-sm text-secondary">Required Secret: {requiredSecrets.join(", ")}</p> : <p className="text-sm text-secondary">该 manifest 未声明 required Secret。</p>}
        <Button type="submit" disabled={submitting || versions.length === 0}>创建 Draft Source</Button>
      </form>
    </div>
  );
}
