import type { Collection, Episode, ImportJob, Program, Source, User } from "../types/domain";

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
    author: "Podcast Hub 编辑部",
    category: "研究简报",
    language: "中文",
    updateFrequency: "每周更新",
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
    author: "口述史项目组",
    category: "深度访谈",
    language: "中文",
    updateFrequency: "不定期更新",
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
    author: "RSS 汇编台",
    category: "资讯摘要",
    language: "中文",
    updateFrequency: "每日更新",
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
    author: "制作团队",
    category: "幕后记录",
    language: "中文",
    updateFrequency: "筹备中",
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

export const episodes: Episode[] = [
  {
    id: "episode_editorial_001",
    programId: "program_editorial_signal",
    title: "从授权源到个人 RSS 的内容路径",
    publishedAt: "2026-06-24",
    duration: "32:18",
    summary: "梳理内容接入、审核和 RSS 发布之间的职责边界。"
  },
  {
    id: "episode_editorial_002",
    programId: "program_editorial_signal",
    title: "编辑部如何处理来源备注和权利说明",
    publishedAt: "2026-06-17",
    duration: "28:04",
    summary: "讨论每个单集进入发布前应保留哪些来源和授权信息。"
  },
  {
    id: "episode_editorial_003",
    programId: "program_editorial_signal",
    title: "可信订阅体验中的安全提示",
    publishedAt: "2026-06-10",
    duration: "24:42",
    summary: "解释私有 RSS 链接为什么需要明确的撤销和权限提示。"
  },
  {
    id: "episode_field_001",
    programId: "program_field_archive",
    title: "一次非常长的田野访谈标题，用于检查移动端、卡片和 RSS 预览中的换行表现",
    publishedAt: "2026-06-20",
    duration: "01:12:09",
    summary: "长标题边界样例，摘要也包含较长的上下文描述，用来验证文本不会挤压操作按钮或造成横向滚动。"
  },
  {
    id: "episode_field_002",
    programId: "program_field_archive",
    title: "档案整理中的人工校注",
    publishedAt: "2026-06-03",
    duration: "48:33",
    summary: "记录访谈材料进入节目库前的校注、去重和审核步骤。"
  },
  {
    id: "episode_rss_001",
    programId: "program_rss_digest",
    title: "原生 RSS 导入器的规范化样例",
    publishedAt: "2026-06-25",
    duration: "18:20",
    summary: "展示平台内建 RSS 导入器如何把授权公开源转为待审核单集。"
  },
  {
    id: "episode_rss_002",
    programId: "program_rss_digest",
    title: "公开源也需要发布前审核",
    publishedAt: "2026-06-19",
    duration: "21:47",
    summary: "说明公开 RSS 并不等于可绕过平台审核和授权检查。"
  },
  {
    id: "episode_rss_003",
    programId: "program_rss_digest",
    title: "订阅客户端里的元数据一致性",
    publishedAt: "2026-06-12",
    duration: "16:05",
    summary: "关注标题、摘要、封面和发布时间在外部客户端中的表现。"
  }
];

export const collections: Collection[] = [
  {
    id: "collection_research_weekly",
    title: "研究周报",
    description: "把研究简报和 RSS 摘要合并成一个个人订阅源。",
    programIds: ["program_editorial_signal", "program_rss_digest"],
    accessScope: "private",
    rssTokenState: "active",
    lastUpdatedAt: "今天 10:12",
    rules: {
      sortOrder: "newest",
      perProgramLimit: 2,
      totalLimit: 5
    }
  },
  {
    id: "collection_interviews",
    title: "访谈资料夹",
    description: "收集已授权的深度访谈节目，用于外部客户端订阅。",
    programIds: ["program_field_archive"],
    accessScope: "private",
    rssTokenState: "active",
    lastUpdatedAt: "昨天 17:30",
    rules: {
      sortOrder: "newest",
      perProgramLimit: 3,
      totalLimit: 6
    }
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
