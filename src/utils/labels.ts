import type { AdminUser, AuthMode, Connector, ExecutionMode, ImportJob, Program, ReviewItem, Source, SourceIngestionType, TriggerType } from "../types/domain";

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
  waiting_auth: "等待授权",
  waiting_for_auth: "等待授权",
  waiting_manual_upload: "等待人工上传",
  review_pending: "等待审核",
  completed: "已完成",
  completed_with_warnings: "完成但有警告",
  failed: "失败",
  cancelled: "已取消",
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
  qr_expired: "二维码已过期",
  operator_cancelled: "操作员取消"
};

export const accessStateLabel: Record<Program["accessState"], string> = {
  authorized: "已授权",
  public: "公开可访问",
  blocked: "访问受限"
};

export const rightsStateLabel: Record<Program["rightsState"], string> = {
  clear: "权利已确认",
  hold: "权利暂缓",
  needs_note: "需要补充说明"
};

export const rssTokenStateLabel = {
  active: "RSS 已启用",
  revoked: "RSS 已撤销",
  paused: "RSS 已暂停"
};

export const connectorStatusLabel: Record<Connector["status"], string> = {
  native_builtin: "平台内建",
  approved: "已审核",
  pending_review: "待审核",
  validation_failed: "校验失败",
  disabled: "已禁用",
  revoked: "已撤销"
};

export const connectorKindLabel: Record<Connector["kind"], string> = {
  native_rss: "平台内建 Native RSS Importer",
  python_connector: "Python Connector",
  manual_import: "人工导入工作流"
};

export const reviewStatusLabel: Record<ReviewItem["status"], string> = {
  pending_review: "待审核",
  approved: "已通过",
  rejected: "已拒绝",
  needs_revision: "退回补充",
  duplicate_risk: "重复风险",
  on_hold: "暂停发布"
};

export const userRoleLabel: Record<AdminUser["role"], string> = {
  user: "普通用户",
  admin: "管理员"
};

export const userStatusLabel: Record<AdminUser["status"], string> = {
  pending_verification: "待邮箱验证",
  active: "活跃",
  suspended: "已暂停",
  deleted: "已删除"
};

export const responsibilityLabel: Record<AdminUser["responsibilityLabels"][number], string> = {
  system_owner: "System Owner",
  operator: "Operator",
  reviewer: "Reviewer"
};
