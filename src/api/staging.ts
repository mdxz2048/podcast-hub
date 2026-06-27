import { apiRequest, readCookie } from "./client";
import type { ImportJob } from "./jobs";

export interface StagingProgram {
  id: string;
  canonical_key: string;
  title: string;
  description: string;
  author: string;
  language: string;
  status: "staging" | "review_pending" | "approved" | "published" | "archived" | "rejected";
  created_from_source_id: string;
  created_from_job_id: string;
  created_at: string;
  updated_at: string;
}

export interface StagingEpisode {
  id: string;
  program_id: string;
  external_episode_id: string;
  title: string;
  description: string;
  published_at: string;
  duration_seconds: number;
  status: "staging" | "review_pending" | "approved" | "rejected" | "published" | "archived";
  source_job_id: string;
  created_at: string;
  updated_at: string;
}

export interface IntakeRun {
  id: string;
  import_job_id: string;
  status: "succeeded" | "failed";
  validation_issues_redacted: string;
  program_id?: string;
  created_at: string;
  updated_at: string;
}

export interface IntakeResult {
  intake_run: IntakeRun;
  program?: StagingProgram;
  validation_issues?: string[];
}

function csrfHeaders() {
  const csrf = readCookie("podcast_hub_csrf");
  return csrf ? { "X-CSRF-Token": csrf } : undefined;
}

export async function runImportJobIntake(jobId: string) {
  return apiRequest<IntakeResult>(`/admin/import-jobs/${encodeURIComponent(jobId)}/intake`, { method: "POST", headers: csrfHeaders() });
}

export async function getImportJobIntakeStatus(jobId: string) {
  return apiRequest<{ intake_run: IntakeRun | null }>(`/admin/import-jobs/${encodeURIComponent(jobId)}/intake-status`);
}

export async function listStagingPrograms() {
  return apiRequest<{ programs: StagingProgram[] }>("/admin/staging/programs");
}

export async function getStagingProgram(programId: string) {
  return apiRequest<{ program: StagingProgram }>(`/admin/staging/programs/${encodeURIComponent(programId)}`);
}

export async function listStagingEpisodes() {
  return apiRequest<{ episodes: StagingEpisode[] }>("/admin/staging/episodes");
}

export async function getStagingEpisode(episodeId: string) {
  return apiRequest<{ episode: StagingEpisode }>(`/admin/staging/episodes/${encodeURIComponent(episodeId)}`);
}

export function canRunIntake(job: ImportJob, hasMetadataBundle: boolean, intake: IntakeRun | null) {
  if (intake?.status === "succeeded") return { allowed: false, reason: "已导入待审核区。" };
  if (job.status === "failed") return { allowed: false, reason: "失败任务不能导入待审核区。" };
  if (job.status === "cancelled") return { allowed: false, reason: "已取消任务不能导入待审核区。" };
  if (job.status !== "completed") return { allowed: false, reason: "任务完成后才能导入待审核区。" };
  if (!hasMetadataBundle) return { allowed: false, reason: "缺少已登记 metadata_bundle artifact。" };
  return { allowed: true, reason: "可导入待审核区。" };
}
