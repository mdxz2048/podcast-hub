import { apiRequest, readCookie } from "./client";

export type ConnectorStatus = "active" | "disabled";
export type ReviewStatus = "pending_review" | "approved" | "rejected" | "disabled";

export interface AdminConnector {
  id: string;
  slug: string;
  name: string;
  description: string;
  status: ConnectorStatus;
  created_at: string;
  updated_at: string;
}

export interface AdminConnectorVersion {
  id: string;
  connector_id: string;
  version: string;
  review_status: ReviewStatus;
  runtime_profile: string;
  entrypoint: string;
  manifest: Record<string, unknown>;
  package_sha256: string;
  package_size_bytes: number;
  validation_summary_json: string;
  uploaded_by?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  created_at: string;
}

export interface ValidationIssue {
  code: string;
  message: string;
  path?: string;
}

export interface ValidationSummary {
  is_valid: boolean;
  issues: ValidationIssue[];
}

function csrfHeaders() {
  const csrf = readCookie("podcast_hub_csrf");
  return csrf ? { "X-CSRF-Token": csrf } : undefined;
}

export async function listConnectors() {
  return apiRequest<{ connectors: AdminConnector[] }>("/admin/connectors");
}

export async function getConnector(connectorId: string) {
  return apiRequest<{ connector: AdminConnector }>(`/admin/connectors/${encodeURIComponent(connectorId)}`);
}

export async function listConnectorVersions(connectorId: string) {
  return apiRequest<{ versions: AdminConnectorVersion[] }>(`/admin/connectors/${encodeURIComponent(connectorId)}/versions`);
}

export async function getConnectorVersion(versionId: string) {
  return apiRequest<{ version: AdminConnectorVersion; validation_summary: ValidationSummary }>(`/admin/connector-versions/${encodeURIComponent(versionId)}`);
}

export async function uploadConnectorPackage(payload: { connector_id: string; version: string; file: File }) {
  const form = new FormData();
  form.append("connector_id", payload.connector_id);
  form.append("version", payload.version);
  form.append("package", payload.file);
  return apiRequest<{ connector: AdminConnector; version: AdminConnectorVersion; validation_summary: ValidationSummary; note: string }>("/admin/connectors/upload", {
    method: "POST",
    body: form,
    headers: csrfHeaders()
  });
}

export async function approveConnectorVersion(versionId: string) {
  return apiRequest<{ version: AdminConnectorVersion }>(`/admin/connector-versions/${encodeURIComponent(versionId)}/approve`, { method: "POST", headers: csrfHeaders() });
}

export async function rejectConnectorVersion(versionId: string) {
  return apiRequest<{ version: AdminConnectorVersion }>(`/admin/connector-versions/${encodeURIComponent(versionId)}/reject`, { method: "POST", headers: csrfHeaders() });
}

export async function disableConnectorVersion(versionId: string) {
  return apiRequest<{ version: AdminConnectorVersion }>(`/admin/connector-versions/${encodeURIComponent(versionId)}/disable`, { method: "POST", headers: csrfHeaders() });
}

export async function enableConnector(connectorId: string) {
  return apiRequest<{ connector: AdminConnector }>(`/admin/connectors/${encodeURIComponent(connectorId)}/enable`, { method: "POST", headers: csrfHeaders() });
}

export async function disableConnector(connectorId: string) {
  return apiRequest<{ connector: AdminConnector }>(`/admin/connectors/${encodeURIComponent(connectorId)}/disable`, { method: "POST", headers: csrfHeaders() });
}

