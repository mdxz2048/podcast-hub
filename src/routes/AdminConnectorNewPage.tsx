import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { uploadConnectorPackage } from "../api/connectors";
import type { ValidationSummary } from "../api/connectors";
import type { ApiError } from "../api/client";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminConnectorNewPage() {
  const navigate = useNavigate();
  const [connectorID, setConnectorID] = useState("");
  const [version, setVersion] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [validation, setValidation] = useState<ValidationSummary | null>(null);
  const [createdConnectorId, setCreatedConnectorId] = useState<string | null>(null);

  async function submitUpload(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setValidation(null);
    if (!file) {
      setError("请选择一个 Connector ZIP 文件。");
      return;
    }
    setIsSubmitting(true);
    try {
      const result = await uploadConnectorPackage({
        connector_id: connectorID.trim(),
        version: version.trim(),
        file
      });
      setValidation(result.validation_summary);
      setCreatedConnectorId(result.connector.id);
      if (result.validation_summary.is_valid) {
        navigate(`/admin/connectors/${result.connector.id}`);
      }
    } catch (e) {
      const apiError = e as ApiError;
      setError(apiError.message || "上传失败，请稍后重试。");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="Connector 上传" title="上传并登记 Python Connector ZIP（仅静态校验）" />
      <section className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
        平台仅做上传登记与安全校验，不会执行 Connector，不会调用 Python，不会下载媒体。请勿上传 Secret、Cookie、Session、媒体文件或 Dockerfile。
      </section>
      {isSubmitting ? <LoadingState title="正在上传并校验 Connector ZIP" /> : null}
      {error ? <ErrorState title={error} /> : null}
      {validation?.is_valid ? <SuccessFeedback message="上传成功，版本已登记为 pending_review，等待管理员审核。" /> : null}
      <form className="grid gap-5 rounded-lg border border-border bg-surface p-6 shadow-subtle" onSubmit={submitUpload}>
        <Input label="Connector ID（slug）" value={connectorID} onChange={(event) => setConnectorID(event.target.value)} placeholder="example-connector" />
        <Input label="Version（semver）" value={version} onChange={(event) => setVersion(event.target.value)} placeholder="1.0.0" />
        <label className="grid gap-2 text-sm">
          <span className="font-medium text-primary">Connector ZIP 文件</span>
          <input
            type="file"
            accept=".zip,application/zip"
            onChange={(event) => setFile(event.target.files?.[0] ?? null)}
            className="rounded-md border border-border bg-subtle p-2"
          />
          <span className="text-xs text-secondary">{file ? `已选择：${file.name}` : "仅支持 Python Connector ZIP"}</span>
        </label>
        <div className="flex flex-wrap gap-2">
          <Button type="submit" disabled={isSubmitting}>上传并校验</Button>
          {createdConnectorId ? <Button type="button" variant="secondary" onClick={() => navigate(`/admin/connectors/${createdConnectorId}`)}>查看详情</Button> : null}
        </div>
      </form>
      {validation && !validation.is_valid ? (
        <section className="rounded-lg border border-danger/30 bg-danger/5 p-5">
          <h2 className="text-lg font-semibold">校验失败</h2>
          <ul className="mt-3 list-disc space-y-2 pl-5 text-sm text-secondary">
            {validation.issues.map((issue) => (
              <li key={`${issue.code}-${issue.path ?? ""}`}>{issue.message}{issue.path ? `（${issue.path}）` : ""}</li>
            ))}
          </ul>
        </section>
      ) : null}
    </div>
  );
}
