export type UserRole = "user" | "admin";
export type UserStatus = "pending_verification" | "active" | "suspended" | "deleted";

export type ProgramStatus = "active" | "draft" | "rights_hold" | "archived";
export type AccessState = "authorized" | "public" | "blocked";
export type SourceIngestionType = "native_rss" | "connector" | "manual_upload";
export type TriggerType = "manual" | "scheduled";
export type AuthMode = "none" | "reusable_session" | "qr_each_run";
export type ExecutionMode = "unattended" | "interactive";
export type JobStatus =
  | "queued"
  | "running"
  | "waiting_for_auth"
  | "completed"
  | "completed_with_warnings"
  | "failed"
  | "timed_out";

export interface User {
  id: string;
  email: string;
  displayName: string;
  role: UserRole;
  status: UserStatus;
  responsibilityLabels?: Array<"system_owner" | "operator" | "reviewer">;
}

export interface Program {
  id: string;
  title: string;
  description: string;
  coverTone: [string, string];
  status: ProgramStatus;
  rightsState: "clear" | "hold" | "needs_note";
  publicationState: "public" | "selected_users" | "private" | "paused";
  episodeCount: number;
  sourceCount: number;
  accessState: AccessState;
  lastUpdated: string;
}

export interface Source {
  id: string;
  programId: string;
  name: string;
  ingestionType: SourceIngestionType;
  triggerType: TriggerType;
  authMode: AuthMode;
  executionMode: ExecutionMode;
  status: "ready" | "auth_required" | "schedule_paused" | "disabled";
  lastJobStatus: JobStatus;
  nextRunAt?: string;
}

export interface ImportJob {
  id: string;
  programId: string;
  sourceId: string;
  ingestionType: SourceIngestionType;
  triggerType: TriggerType;
  authMode: AuthMode;
  executionMode: ExecutionMode;
  status: JobStatus;
  startedAt: string;
  finishedAt?: string;
  errorCategory?: string;
  nextAction: string;
}

