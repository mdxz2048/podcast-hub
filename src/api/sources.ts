import { apiRequest, readCookie } from "./client";
import type { AdminConnectorVersion } from "./connectors";

export type SourceStatus = "draft" | "active" | "disabled";
export type SecretType = "text" | "file";

export interface ConnectorSource {
  id: string;
  connector_version_id: string;
  name: string;
  description: string;
  status: SourceStatus;
  trigger_type: "manual";
  auth_mode: "none" | "reusable_session";
  execution_mode: "unattended";
  config_json: string;
  network_mode: "disabled" | "trusted_admin";
  created_at: string;
  updated_at: string;
}

export interface SecretRecord {
  id: string;
  name: string;
  secret_type: SecretType;
  encryption_version: string;
  created_at: string;
  revoked_at?: string;
  binding_count: number;
}

export interface SourceSecretBinding {
  id: string;
  connector_source_id: string;
  secret_name: string;
  secret_record_id: string;
  created_at: string;
}

export interface ConnectorSourceDetail {
  source: ConnectorSource;
  secret_bindings: SourceSecretBinding[];
  required_secrets: string[];
  missing_secrets: string[];
}

function csrfHeaders(extra?: HeadersInit) {
  const csrf = readCookie("podcast_hub_csrf");
  return { ...(extra ?? {}), ...(csrf ? { "X-CSRF-Token": csrf } : {}) };
}

export async function listSources() {
  return apiRequest<{ sources: ConnectorSource[] }>("/admin/sources");
}

export async function createSource(payload: {
  connector_version_id: string;
  name: string;
  description: string;
  trigger_type: "manual";
  auth_mode: "none" | "reusable_session";
  execution_mode: "unattended";
  network_mode: "disabled" | "trusted_admin";
  config: Record<string, unknown>;
}) {
  return apiRequest<ConnectorSourceDetail>("/admin/sources", {
    method: "POST",
    headers: csrfHeaders({ "Content-Type": "application/json" }),
    body: JSON.stringify(payload)
  });
}

export async function getSource(sourceId: string) {
  return apiRequest<ConnectorSourceDetail>(`/admin/sources/${encodeURIComponent(sourceId)}`);
}

export async function enableSource(sourceId: string) {
  return apiRequest<ConnectorSourceDetail>(`/admin/sources/${encodeURIComponent(sourceId)}/enable`, { method: "POST", headers: csrfHeaders() });
}

export async function disableSource(sourceId: string) {
  return apiRequest<ConnectorSourceDetail>(`/admin/sources/${encodeURIComponent(sourceId)}/disable`, { method: "POST", headers: csrfHeaders() });
}

export async function listSecrets() {
  return apiRequest<{ secrets: SecretRecord[] }>("/admin/secrets");
}

export async function createTextSecret(payload: { name: string; value: string }) {
  return apiRequest<{ secret: SecretRecord }>("/admin/secrets/text", {
    method: "POST",
    headers: csrfHeaders({ "Content-Type": "application/json" }),
    body: JSON.stringify(payload)
  });
}

export async function revokeSecret(secretId: string) {
  return apiRequest<{ secret: SecretRecord }>(`/admin/secrets/${encodeURIComponent(secretId)}/revoke`, { method: "POST", headers: csrfHeaders() });
}

export async function bindSourceSecret(sourceId: string, payload: { secret_name: string; secret_record_id: string }) {
  return apiRequest<ConnectorSourceDetail>(`/admin/sources/${encodeURIComponent(sourceId)}/secret-bindings`, {
    method: "POST",
    headers: csrfHeaders({ "Content-Type": "application/json" }),
    body: JSON.stringify(payload)
  });
}

export function manifestSecrets(version: AdminConnectorVersion): string[] {
  const manifest = version.manifest as { secrets?: Array<{ name?: string }> };
  return (manifest.secrets ?? []).map((item) => item.name ?? "").filter(Boolean);
}
