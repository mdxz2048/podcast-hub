import type { ImportJob, Program, Source, User } from "../types/domain";

export const currentUser: User = {
  id: "mock_user_admin",
  email: "owner@example.invalid",
  displayName: "M0 系统负责人",
  role: "admin",
  status: "active",
  responsibilityLabels: ["system_owner", "operator", "reviewer"]
};

export const programs: Program[] = [
  {
    id: "program_editorial_signal",
    title: "编辑信号",
    description: "面向内部研究简报和已清权访谈的精选音频刊物。",
    coverTone: ["#236058", "#a6671f"],
    status: "active",
    rightsState: "clear",
    publicationState: "selected_users",
    episodeCount: 42,
    sourceCount: 3,
    accessState: "authorized",
    lastUpdated: "今天 09:20"
  },
  {
    id: "program_field_archive",
    title: "田野档案：用于验证长标题换行的深度访谈节目",
    description: "为限定成员整理的口述史录音，包含审核备注和来源溯源信息。",
    coverTone: ["#305e91", "#2d7a59"],
    status: "rights_hold",
    rightsState: "hold",
    publicationState: "paused",
    episodeCount: 18,
    sourceCount: 5,
    accessState: "authorized",
    lastUpdated: "昨天 18:42"
  },
  {
    id: "program_rss_digest",
    title: "原生 RSS 摘要",
    description: "授权公开源会先经过平台内建 RSS 导入器规范化，再进入审核流程。",
    coverTone: ["#6f756d", "#236058"],
    status: "active",
    rightsState: "clear",
    publicationState: "public",
    episodeCount: 86,
    sourceCount: 1,
    accessState: "public",
    lastUpdated: "周一 11:05"
  },
  {
    id: "program_empty",
    title: "制作札记",
    description: "一个尚未发布单集的节目壳，用于检查空状态表现。",
    coverTone: ["#b1b9b0", "#a6671f"],
    status: "draft",
    rightsState: "needs_note",
    publicationState: "private",
    episodeCount: 0,
    sourceCount: 0,
    accessState: "blocked",
    lastUpdated: "6月20日"
  }
];

export const sources: Source[] = [
  {
    id: "source_native_rss",
    programId: "program_rss_digest",
    name: "授权公开 RSS",
    ingestionType: "native_rss",
    triggerType: "scheduled",
    authMode: "none",
    executionMode: "unattended",
    status: "ready",
    lastJobStatus: "completed",
    nextRunAt: "今天 22:00"
  },
  {
    id: "source_connector_session",
    programId: "program_editorial_signal",
    name: "合作方档案 Connector",
    ingestionType: "connector",
    triggerType: "scheduled",
    authMode: "reusable_session",
    executionMode: "unattended",
    status: "auth_required",
    lastJobStatus: "waiting_for_auth",
    nextRunAt: "已暂停"
  },
  {
    id: "source_qr_each_run",
    programId: "program_field_archive",
    name: "扫码授权研究门户",
    ingestionType: "connector",
    triggerType: "manual",
    authMode: "qr_each_run",
    executionMode: "interactive",
    status: "auth_required",
    lastJobStatus: "waiting_for_auth"
  }
];

export const jobs: ImportJob[] = [
  {
    id: "job_mock_1001",
    programId: "program_editorial_signal",
    sourceId: "source_connector_session",
    ingestionType: "connector",
    triggerType: "scheduled",
    authMode: "reusable_session",
    executionMode: "unattended",
    status: "waiting_for_auth",
    startedAt: "09:15",
    errorCategory: "auth_required",
    nextAction: "刷新可复用会话"
  },
  {
    id: "job_mock_1002",
    programId: "program_rss_digest",
    sourceId: "source_native_rss",
    ingestionType: "native_rss",
    triggerType: "scheduled",
    authMode: "none",
    executionMode: "unattended",
    status: "completed",
    startedAt: "08:00",
    finishedAt: "08:01",
    nextAction: "审核 3 个新单集"
  },
  {
    id: "job_mock_1003",
    programId: "program_field_archive",
    sourceId: "source_qr_each_run",
    ingestionType: "connector",
    triggerType: "manual",
    authMode: "qr_each_run",
    executionMode: "interactive",
    status: "failed",
    startedAt: "昨天 21:11",
    finishedAt: "昨天 21:16",
    errorCategory: "qr_expired",
    nextAction: "创建新的手动任务"
  }
];
