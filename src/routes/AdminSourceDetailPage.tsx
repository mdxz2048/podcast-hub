import { FormEvent, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { bindSourceSecret, enableSource, getSource, listSecrets } from "../api/sources";
import type { ConnectorSourceDetail, SecretRecord } from "../api/sources";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Select } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminSourceDetailPage() {
  const { sourceId = "" } = useParams();
  const [detail, setDetail] = useState<ConnectorSourceDetail | null>(null);
  const [secrets, setSecrets] = useState<SecretRecord[]>([]);
  const [selectedSecret, setSelectedSecret] = useState("");
  const [secretName, setSecretName] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function reload() {
    const [sourceResult, secretResult] = await Promise.all([getSource(sourceId), listSecrets()]);
    setDetail(sourceResult);
    setSecrets(secretResult.secrets);
    setSecretName(sourceResult.missing_secrets[0] ?? sourceResult.required_secrets[0] ?? "");
    setSelectedSecret(secretResult.secrets.find((secret) => !secret.revoked_at)?.id ?? "");
  }

  useEffect(() => {
    reload().catch(() => setError("无法加载 Source 详情。")).finally(() => setLoading(false));
  }, [sourceId]);

  async function bind(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const result = await bindSourceSecret(sourceId, { secret_name: secretName, secret_record_id: selectedSecret });
    setDetail(result);
    setSuccess("Secret Reference 已绑定。");
  }

  async function enable() {
    try {
      const result = await enableSource(sourceId);
      setDetail(result);
      setSuccess("Source 已启用。");
    } catch {
      setError("Source 仍缺少 required Secret，或绑定 Secret 已撤销。");
    }
  }

  if (loading) return <LoadingState title="正在加载 Source 详情" />;
  if (error && !detail) return <ErrorState title={error} />;
  if (!detail) return <ErrorState title="Source 不存在" />;

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="Source Detail" title={detail.source.name} />
      {success ? <SuccessFeedback message={success} /> : null}
      {error ? <ErrorState title={error} /> : null}
      <section className="rounded-lg border border-border bg-surface p-6 shadow-subtle">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className="text-sm text-secondary">{detail.source.description || "未填写描述"}</p>
            <p className="mt-2 text-sm text-secondary">Version: {detail.source.connector_version_id}</p>
          </div>
          <Badge>{detail.source.status}</Badge>
        </div>
        <div className="mt-5 grid gap-3 text-sm text-secondary md:grid-cols-4">
          <span>Trigger: {detail.source.trigger_type}</span>
          <span>Auth: {detail.source.auth_mode}</span>
          <span>Execution: {detail.source.execution_mode}</span>
          <span>Network: {detail.source.network_mode}</span>
        </div>
        <Button className="mt-5" type="button" onClick={enable} disabled={detail.missing_secrets.length > 0}>Enable Source</Button>
      </section>
      <section className="rounded-lg border border-border bg-surface p-6 shadow-subtle">
        <h2 className="font-semibold">Secret References</h2>
        <p className="mt-1 text-sm text-secondary">只显示绑定状态，不显示 Secret 值。</p>
        <div className="mt-4 grid gap-2">
          {detail.required_secrets.map((name) => {
            const binding = detail.secret_bindings.find((item) => item.secret_name === name);
            return (
              <div key={name} className="flex items-center justify-between rounded-md border border-border p-3 text-sm">
                <span>{name}</span>
                <Badge>{binding ? "bound" : "missing"}</Badge>
              </div>
            );
          })}
        </div>
        {detail.required_secrets.length > 0 ? (
          <form className="mt-5 grid gap-4 md:grid-cols-[1fr_1fr_auto]" onSubmit={bind}>
            <Select label="Secret Name" value={secretName} onChange={(event) => setSecretName(event.target.value)} options={detail.required_secrets} />
            <Select label="Secret Record" value={selectedSecret} onChange={(event) => setSelectedSecret(event.target.value)} options={secrets.filter((secret) => !secret.revoked_at).map((secret) => secret.id)} />
            <Button className="self-end" type="submit" disabled={!selectedSecret || !secretName}>绑定</Button>
          </form>
        ) : null}
      </section>
    </div>
  );
}
