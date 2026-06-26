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
  | "waiting_auth"
  | "waiting_for_auth"
  | "waiting_manual_upload"
  | "review_pending"
  | "completed"
  | "completed_with_warnings"
  | "failed"
  | "cancelled"
  | "timed_out";

export type ConnectorStatus = "native_builtin" | "approved" | "pending_review" | "validation_failed" | "disabled" | "revoked";
export type ReviewStatus = "pending_review" | "approved" | "rejected" | "needs_revision" | "duplicate_risk" | "on_hold";
export type RssState = "active" | "revoked" | "paused";

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
  author: string;
  category: string;
  language: string;
  updateFrequency: string;
  coverTone: [string, string];
  status: ProgramStatus;
  rightsState: "clear" | "hold" | "needs_note";
  publicationState: "public" | "selected_users" | "private" | "paused";
  episodeCount: number;
  sourceCount: number;
  accessState: AccessState;
  lastUpdated: string;
}

export interface Episode {
  id: string;
  programId: string;
  title: string;
  publishedAt: string;
  duration: string;
  summary: string;
}

export interface CollectionRules {
  sortOrder: "newest" | "oldest";
  perProgramLimit: number;
  totalLimit: number;
}

export interface Collection {
  id: string;
  title: string;
  description: string;
  programIds: string[];
  accessScope: "private" | "selected_users";
  rssTokenState: "active" | "revoked";
  lastUpdatedAt: string;
  rules: CollectionRules;
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
  connectorId?: string;
  connectorVersion?: string;
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
  connectorId?: string;
  connectorVersion?: string;
  progress?: number;
  logEvents?: string[];
  timeline?: Array<{ label: string; at: string; tone: "success" | "warning" | "danger" | "info" | "neutral" }>;
  outputEpisodeIds?: string[];
}

export interface Connector {
  id: string;
  name: string;
  kind: "native_rss" | "python_connector" | "manual_import";
  version: string;
  status: ConnectorStatus;
  supportedIngestionTypes: SourceIngestionType[];
  supportedTriggerTypes: TriggerType[];
  authModes: AuthMode[];
  executionModes: ExecutionMode[];
  entrypoint: string;
  dependencyLock: string;
  networkPolicy: string[];
  resourceLimits: {
    memoryMb: number;
    timeoutSeconds: number;
    maxDownloadMb: number;
  };
  versionHistory: Array<{ version: string; status: ConnectorStatus; date: string }>;
  boundSourceIds: string[];
  lastJobStatus: JobStatus;
  nextAction: string;
}

export interface ReviewItem {
  id: string;
  programId: string;
  sourceId: string;
  jobId: string;
  episodeId: string;
  status: ReviewStatus;
  metadataCompleteness: number;
  rightsState: "clear" | "needs_confirmation" | "hold";
  duplicateRisk: "none" | "possible" | "high";
  fileState: "staged" | "missing" | "metadata_only";
  publishDate: string;
  suggestion: string;
}

export interface AdminUser {
  id: string;
  email: string;
  displayName: string;
  role: UserRole;
  status: UserStatus;
  responsibilityLabels: Array<"system_owner" | "operator" | "reviewer">;
  accessibleProgramCount: number;
  privateRssState: RssState;
  lastActivity: string;
  accessSummary: string;
}
