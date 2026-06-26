import type { AuthMode, ExecutionMode, ImportJob, Program, Source, SourceIngestionType, TriggerType } from "../types/domain";

export const programStatusLabel: Record<Program["status"], string> = {
  active: "已启用",
  draft: "草稿",
  rights_hold: "权利暂缓",
  archived: "已归档"
};

export const publicationStateLabel: Record<Program["publicationState"], string> = {
  public: "公开 RSS",
  selected_users: "指定用户",
  private: "私有",
  paused: "已暂停"
};

export const sourceStatusLabel: Record<Source["status"], string> = {
  ready: "就绪",
  auth_required: "需要授权",
  schedule_paused: "调度暂停",
  disabled: "已停用"
};

export const jobStatusLabel: Record<ImportJob["status"], string> = {
  queued: "排队中",
  running: "运行中",
  waiting_for_auth: "等待授权",
  completed: "已完成",
  completed_with_warnings: "完成但有警告",
  failed: "失败",
  timed_out: "已超时"
};

export const ingestionTypeLabel: Record<SourceIngestionType, string> = {
  native_rss: "原生 RSS",
  connector: "Connector",
  manual_upload: "手动上传"
};

export const triggerTypeLabel: Record<TriggerType, string> = {
  manual: "手动",
  scheduled: "周期"
};

export const authModeLabel: Record<AuthMode, string> = {
  none: "无授权",
  reusable_session: "可复用会话",
  qr_each_run: "每次扫码"
};

export const executionModeLabel: Record<ExecutionMode, string> = {
  unattended: "无人值守",
  interactive: "交互式"
};

export const errorCategoryLabel: Record<string, string> = {
  auth_required: "需要重新授权",
  qr_expired: "二维码已过期"
};
