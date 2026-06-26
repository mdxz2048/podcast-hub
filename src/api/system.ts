import { apiRequest } from "./client";

export interface RunnerStatus {
  mode: "disabled" | "docker_trusted_admin";
  can_run_jobs: boolean;
  reason: string;
}

export interface AdminSystemStatus {
  api: "ok";
  dependencies?: Record<string, string>;
  runner?: RunnerStatus;
}

export async function getAdminSystemStatus() {
  return apiRequest<AdminSystemStatus>("/admin/system/status");
}
