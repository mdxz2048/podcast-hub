import { apiRequest, readCookie } from "./client";
import type { StagingEpisode, StagingProgram } from "./staging";

export type AdminProgram = StagingProgram;
export type AdminEpisode = StagingEpisode;

export interface ReviewItem {
  id: string;
  target_type: "program" | "episode";
  target_id: string;
  review_kind: string;
  status: "pending" | "approved" | "rejected" | "cancelled";
  requested_by_job_id?: string;
  reviewed_by?: string;
  review_note: string;
  created_at: string;
  reviewed_at?: string;
}

function csrfHeaders() {
  const csrf = readCookie("podcast_hub_csrf");
  return csrf ? { "X-CSRF-Token": csrf } : undefined;
}

export async function listAdminPrograms() {
  return apiRequest<{ programs: AdminProgram[] }>("/admin/programs");
}

export async function getAdminProgram(programId: string) {
  return apiRequest<{ program: AdminProgram; episodes: AdminEpisode[] }>(`/admin/programs/${encodeURIComponent(programId)}`);
}

export async function patchAdminProgram(programId: string, body: Partial<Pick<AdminProgram, "title" | "description" | "author" | "language">>) {
  return apiRequest<{ program: AdminProgram }>(`/admin/programs/${encodeURIComponent(programId)}`, { method: "PATCH", headers: csrfHeaders(), body: JSON.stringify(body) });
}

export async function submitProgramReview(programId: string) {
  return apiRequest<{ review: ReviewItem }>(`/admin/programs/${encodeURIComponent(programId)}/submit-review`, { method: "POST", headers: csrfHeaders() });
}

export async function publishProgram(programId: string) {
  return apiRequest<{ program: AdminProgram }>(`/admin/programs/${encodeURIComponent(programId)}/publish`, { method: "POST", headers: csrfHeaders() });
}

export async function archiveProgram(programId: string) {
  return apiRequest<{ program: AdminProgram }>(`/admin/programs/${encodeURIComponent(programId)}/archive`, { method: "POST", headers: csrfHeaders() });
}

export async function getAdminEpisode(episodeId: string) {
  return apiRequest<{ episode: AdminEpisode }>(`/admin/episodes/${encodeURIComponent(episodeId)}`);
}

export async function patchAdminEpisode(episodeId: string, body: Partial<Pick<AdminEpisode, "title" | "description" | "duration_seconds">>) {
  return apiRequest<{ episode: AdminEpisode }>(`/admin/episodes/${encodeURIComponent(episodeId)}`, { method: "PATCH", headers: csrfHeaders(), body: JSON.stringify(body) });
}

export async function submitEpisodeReview(episodeId: string) {
  return apiRequest<{ review: ReviewItem }>(`/admin/episodes/${encodeURIComponent(episodeId)}/submit-review`, { method: "POST", headers: csrfHeaders() });
}

export async function publishEpisode(episodeId: string) {
  return apiRequest<{ episode: AdminEpisode }>(`/admin/episodes/${encodeURIComponent(episodeId)}/publish`, { method: "POST", headers: csrfHeaders() });
}

export async function archiveEpisode(episodeId: string) {
  return apiRequest<{ episode: AdminEpisode }>(`/admin/episodes/${encodeURIComponent(episodeId)}/archive`, { method: "POST", headers: csrfHeaders() });
}

export async function listReviews() {
  return apiRequest<{ reviews: ReviewItem[] }>("/admin/review");
}

export async function getReview(reviewId: string) {
  return apiRequest<{ review: ReviewItem }>(`/admin/review/${encodeURIComponent(reviewId)}`);
}

export async function approveReview(reviewId: string) {
  return apiRequest<{ review: ReviewItem }>(`/admin/review/${encodeURIComponent(reviewId)}/approve`, { method: "POST", headers: csrfHeaders() });
}

export async function rejectReview(reviewId: string, reason: string) {
  return apiRequest<{ review: ReviewItem }>(`/admin/review/${encodeURIComponent(reviewId)}/reject`, { method: "POST", headers: csrfHeaders(), body: JSON.stringify({ reason }) });
}
