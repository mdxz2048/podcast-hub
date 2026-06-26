import { FormEvent, useEffect, useState } from "react";
import { KeyRound, Plus } from "lucide-react";
import { createTextSecret, listSecrets, revokeSecret } from "../api/sources";
import type { SecretRecord } from "../api/sources";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminSecretsPage() {
  const [secrets, setSecrets] = useState<SecretRecord[]>([]);
  const [name, setName] = useState("");
  const [value, setValue] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function reload() {
    const result = await listSecrets();
    setSecrets(result.secrets);
  }

  useEffect(() => {
    reload().catch(() => setError("无法加载 Secret 列表。")).finally(() => setLoading(false));
  }, []);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    await createTextSecret({ name, value });
    setName("");
    setValue("");
    setSuccess("Secret 已加密保存。");
    await reload();
  }

  async function revoke(secretId: string) {
    await revokeSecret(secretId);
    setSuccess("Secret 已撤销。");
    await reload();
  }

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="Secret Reference" title="Secret 元数据" />
      <section className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
        Secret 值只在写入时加密保存，API 不提供读取值。不要把 Secret 写进 Connector ZIP 或 manifest。
      </section>
      {success ? <SuccessFeedback message={success} /> : null}
      {loading ? <LoadingState title="正在加载 Secret" /> : null}
      {error ? <ErrorState title={error} /> : null}
      <form className="grid gap-4 rounded-lg border border-border bg-surface p-5 shadow-subtle md:grid-cols-[1fr_1fr_auto]" onSubmit={submit}>
        <Input label="名称" value={name} onChange={(event) => setName(event.target.value)} />
        <Input label="Secret 值" value={value} onChange={(event) => setValue(event.target.value)} type="password" />
        <Button className="self-end" type="submit" icon={<Plus className="h-4 w-4" />}>新增</Button>
      </form>
      {!loading && secrets.length === 0 ? <EmptyState title="暂无 Secret" /> : null}
      <div className="grid gap-3">
        {secrets.map((secret) => (
          <div key={secret.id} className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <KeyRound className="h-5 w-5 text-secondary" />
                <div>
                  <h2 className="font-semibold">{secret.name}</h2>
                  <p className="text-sm text-secondary">{secret.secret_type} · {secret.encryption_version} · bindings {secret.binding_count}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Badge>{secret.revoked_at ? "revoked" : "active"}</Badge>
                {!secret.revoked_at ? <Button type="button" variant="danger" onClick={() => revoke(secret.id)}>Revoke</Button> : null}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
