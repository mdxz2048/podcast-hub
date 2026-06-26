import { apiRequest, readCookie } from "./client";

export type ImportJobStatus = "queued" | "running" | "completed" | "failed" | "cancelled";

export interface ImportJob {
  id: string;
  connector_source_id: string;
  connector_version_id: string;
  status: ImportJobStatus;
  trigger_type: "manual";
  auth_mode: "none" | "reusable_session";
  execution_mode: "unattended";
  failure_code?: string;
  failure_message_redacted?: string;
  started_at?: string;
  finished_at?: string;
  timeout_at?: string;
  cancellation_requested_at?: string;
  created_at: string;
  updated_at: string;
}

export interface ImportJobEvent {
  id: string;
  import_job_id: string;
  event_type: string;
  level: string;
  message_redacted: string;
  metadata_redacted: string;
  created_at: string;
}

export interface ImportJobArtifact {
  id: string;
  import_job_id: string;
  artifact_type: string;
  relative_path: string;
  size_bytes: number;
  sha256: string;
  created_at: string;
}

function csrfHeaders() {
  const csrf = readCookie("podcast_hub_csrf");
  return csrf ? { "X-CSRF-Token": csrf } : undefined;
}

export async function listImportJobs() {
  return apiRequest<{ jobs: ImportJob[] }>("/admin/import-jobs");
}

export async function createImportJob(sourceId: string) {
  return apiRequest<{ job: ImportJob }>(`/admin/sources/${encodeURIComponent(sourceId)}/import-jobs`, { method: "POST", headers: csrfHeaders() });
}

export async function getImportJob(jobId: string) {
  return apiRequest<{ job: ImportJob }>(`/admin/import-jobs/${encodeURIComponent(jobId)}`);
}

export async function listImportJobEvents(jobId: string) {
  return apiRequest<{ events: ImportJobEvent[] }>(`/admin/import-jobs/${encodeURIComponent(jobId)}/events`);
}

export async function listImportJobArtifacts(jobId: string) {
  return apiRequest<{ artifacts: ImportJobArtifact[] }>(`/admin/import-jobs/${encodeURIComponent(jobId)}/artifacts`);
}

export async function cancelImportJob(jobId: string) {
  return apiRequest<{ job: ImportJob }>(`/admin/import-jobs/${encodeURIComponent(jobId)}/cancel`, { method: "POST", headers: csrfHeaders() });
}
