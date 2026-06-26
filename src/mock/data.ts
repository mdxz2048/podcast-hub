import type { AdminUser, Collection, Connector, Episode, ImportJob, Program, ReviewItem, Source, User } from "../types/domain";

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
    nextRunAt: "已暂停",
    connectorId: "connector_partner_archive",
    connectorVersion: "1.4.2"
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
    lastJobStatus: "waiting_auth",
    connectorId: "connector_qr_portal",
    connectorVersion: "0.9.0"
  },
  {
    id: "source_manual_upload",
    programId: "program_empty",
    name: "人工上传入口",
    ingestionType: "manual_upload",
    triggerType: "manual",
    authMode: "none",
    executionMode: "unattended",
    status: "ready",
    lastJobStatus: "waiting_manual_upload"
  }
];

export const connectors: Connector[] = [
  {
    id: "connector_native_rss",
    name: "平台内建 Native RSS Importer",
    kind: "native_rss",
    version: "builtin",
    status: "native_builtin",
    supportedIngestionTypes: ["native_rss"],
    supportedTriggerTypes: ["manual", "scheduled"],
    authModes: ["none"],
    executionModes: ["unattended"],
    entrypoint: "platform/native-rss-importer",
    dependencyLock: "平台内建，无外部锁文件",
    networkPolicy: ["来源 RSS 域名"],
    resourceLimits: { memoryMb: 256, timeoutSeconds: 120, maxDownloadMb: 64 },
    versionHistory: [{ version: "builtin", status: "native_builtin", date: "M0 静态" }],
    boundSourceIds: ["source_native_rss"],
    lastJobStatus: "completed",
    nextAction: "继续进入审核队列"
  },
  {
    id: "connector_partner_archive",
    name: "合作方档案 Python Connector",
    kind: "python_connector",
    version: "1.4.2",
    status: "approved",
    supportedIngestionTypes: ["connector"],
    supportedTriggerTypes: ["manual", "scheduled"],
    authModes: ["reusable_session"],
    executionModes: ["unattended"],
    entrypoint: "src/main.py",
    dependencyLock: "requirements.lock",
    networkPolicy: ["archive.example.invalid"],
    resourceLimits: { memoryMb: 512, timeoutSeconds: 600, maxDownloadMb: 512 },
    versionHistory: [
      { version: "1.4.2", status: "approved", date: "2026-06-18" },
      { version: "1.3.0", status: "revoked", date: "2026-05-22" }
    ],
    boundSourceIds: ["source_connector_session"],
    lastJobStatus: "waiting_auth",
    nextAction: "刷新可复用会话后恢复周期任务"
  },
  {
    id: "connector_qr_portal",
    name: "扫码授权研究门户 Connector",
    kind: "python_connector",
    version: "0.9.0",
    status: "pending_review",
    supportedIngestionTypes: ["connector"],
    supportedTriggerTypes: ["manual"],
    authModes: ["qr_each_run"],
    executionModes: ["interactive"],
    entrypoint: "connector.py",
    dependencyLock: "uv.lock",
    networkPolicy: ["research.example.invalid"],
    resourceLimits: { memoryMb: 768, timeoutSeconds: 480, maxDownloadMb: 256 },
    versionHistory: [
      { version: "0.9.0", status: "pending_review", date: "2026-06-21" },
      { version: "0.8.1", status: "validation_failed", date: "2026-06-02" }
    ],
    boundSourceIds: ["source_qr_each_run"],
    lastJobStatus: "failed",
    nextAction: "确认二维码交互状态，不允许设置周期任务"
  },
  {
    id: "connector_manual_upload",
    name: "人工导入工作流",
    kind: "manual_import",
    version: "workflow",
    status: "native_builtin",
    supportedIngestionTypes: ["manual_upload"],
    supportedTriggerTypes: ["manual"],
    authModes: ["none"],
    executionModes: ["unattended"],
    entrypoint: "platform/manual-upload-workflow",
    dependencyLock: "平台内建，无 Connector ZIP",
    networkPolicy: [],
    resourceLimits: { memoryMb: 0, timeoutSeconds: 0, maxDownloadMb: 0 },
    versionHistory: [{ version: "workflow", status: "native_builtin", date: "M0 静态" }],
    boundSourceIds: [],
    lastJobStatus: "waiting_manual_upload",
    nextAction: "管理员补充文件和元数据后进入审核"
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
    status: "waiting_auth",
    startedAt: "09:15",
    errorCategory: "auth_required",
    nextAction: "刷新可复用会话",
    connectorId: "connector_partner_archive",
    connectorVersion: "1.4.2",
    progress: 35,
    logEvents: [
      "09:15 job.started source=source_connector_session",
      "09:16 auth.session_expired secret=redacted",
      "09:16 job.waiting_auth operator_action_required=true"
    ],
    timeline: [
      { label: "任务创建", at: "09:15", tone: "success" },
      { label: "会话失效", at: "09:16", tone: "warning" },
      { label: "等待授权", at: "现在", tone: "warning" }
    ]
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
    nextAction: "审核 3 个新单集",
    connectorId: "connector_native_rss",
    connectorVersion: "builtin",
    progress: 100,
    outputEpisodeIds: ["episode_rss_001", "episode_rss_002"],
    logEvents: ["08:00 job.started", "08:01 output.validated episodes=3", "08:01 review.items_created count=3"],
    timeline: [
      { label: "任务创建", at: "08:00", tone: "success" },
      { label: "产物校验", at: "08:01", tone: "success" },
      { label: "进入审核", at: "08:01", tone: "info" }
    ]
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
    nextAction: "创建新的手动任务",
    connectorId: "connector_qr_portal",
    connectorVersion: "0.9.0",
    progress: 62,
    logEvents: [
      "21:11 job.started",
      "21:12 qr.placeholder_created real_qr=false",
      "21:16 qr.expired",
      "21:16 job.failed reason=qr_expired"
    ],
    timeline: [
      { label: "任务创建", at: "昨天 21:11", tone: "success" },
      { label: "等待扫码授权", at: "昨天 21:12", tone: "warning" },
      { label: "二维码过期", at: "昨天 21:16", tone: "danger" }
    ]
  },
  {
    id: "job_mock_1004",
    programId: "program_editorial_signal",
    sourceId: "source_connector_session",
    ingestionType: "connector",
    triggerType: "manual",
    authMode: "reusable_session",
    executionMode: "unattended",
    status: "running",
    startedAt: "10:04",
    errorCategory: undefined,
    nextAction: "等待产物写入隔离区",
    connectorId: "connector_partner_archive",
    connectorVersion: "1.4.2",
    progress: 72,
    logEvents: ["10:04 job.started", "10:05 connector.event fetching_index", "10:07 connector.event writing_staging_output"],
    timeline: [
      { label: "任务创建", at: "10:04", tone: "success" },
      { label: "Connector 运行中", at: "10:07", tone: "info" }
    ]
  },
  {
    id: "job_mock_1005",
    programId: "program_empty",
    sourceId: "source_manual_upload",
    ingestionType: "manual_upload",
    triggerType: "manual",
    authMode: "none",
    executionMode: "unattended",
    status: "waiting_manual_upload",
    startedAt: "10:22",
    nextAction: "等待管理员补充音频和元数据",
    progress: 10,
    logEvents: ["10:22 manual_import.created", "10:22 waiting_manual_upload"],
    timeline: [
      { label: "人工导入创建", at: "10:22", tone: "success" },
      { label: "等待文件", at: "现在", tone: "warning" }
    ]
  },
  {
    id: "job_mock_1006",
    programId: "program_rss_digest",
    sourceId: "source_native_rss",
    ingestionType: "native_rss",
    triggerType: "scheduled",
    authMode: "none",
    executionMode: "unattended",
    status: "review_pending",
    startedAt: "07:40",
    finishedAt: "07:42",
    nextAction: "打开审核队列处理 2 个单集",
    connectorId: "connector_native_rss",
    connectorVersion: "builtin",
    progress: 100,
    outputEpisodeIds: ["episode_rss_001", "episode_rss_002"],
    logEvents: ["07:40 native_rss.fetch", "07:42 review.pending count=2"],
    timeline: [
      { label: "导入完成", at: "07:42", tone: "success" },
      { label: "等待审核", at: "现在", tone: "info" }
    ]
  },
  {
    id: "job_mock_1007",
    programId: "program_field_archive",
    sourceId: "source_qr_each_run",
    ingestionType: "connector",
    triggerType: "manual",
    authMode: "qr_each_run",
    executionMode: "interactive",
    status: "cancelled",
    startedAt: "昨天 19:03",
    finishedAt: "昨天 19:06",
    errorCategory: "operator_cancelled",
    nextAction: "如需继续，请重新创建手动任务",
    connectorId: "connector_qr_portal",
    connectorVersion: "0.9.0",
    progress: 20,
    logEvents: ["19:03 job.started", "19:05 operator.cancel_requested", "19:06 job.cancelled"],
    timeline: [
      { label: "任务创建", at: "昨天 19:03", tone: "success" },
      { label: "操作员取消", at: "昨天 19:06", tone: "neutral" }
    ]
  }
];

export const reviewItems: ReviewItem[] = [
  {
    id: "review_mock_001",
    programId: "program_rss_digest",
    sourceId: "source_native_rss",
    jobId: "job_mock_1006",
    episodeId: "episode_rss_001",
    status: "pending_review",
    metadataCompleteness: 96,
    rightsState: "clear",
    duplicateRisk: "none",
    fileState: "staged",
    publishDate: "2026-06-25",
    suggestion: "元数据完整，可进入发布前确认。"
  },
  {
    id: "review_mock_002",
    programId: "program_field_archive",
    sourceId: "source_qr_each_run",
    jobId: "job_mock_1003",
    episodeId: "episode_field_001",
    status: "duplicate_risk",
    metadataCompleteness: 82,
    rightsState: "needs_confirmation",
    duplicateRisk: "high",
    fileState: "metadata_only",
    publishDate: "2026-06-20",
    suggestion: "长标题访谈疑似重复，需要比对来源备注。"
  },
  {
    id: "review_mock_003",
    programId: "program_editorial_signal",
    sourceId: "source_connector_session",
    jobId: "job_mock_1004",
    episodeId: "episode_editorial_001",
    status: "on_hold",
    metadataCompleteness: 76,
    rightsState: "hold",
    duplicateRisk: "possible",
    fileState: "staged",
    publishDate: "2026-06-24",
    suggestion: "权利备注未确认，建议暂停发布。"
  }
];

export const adminUsers: AdminUser[] = [
  {
    id: "admin_user_owner",
    email: "owner@example.invalid",
    displayName: "系统负责人",
    role: "admin",
    status: "active",
    responsibilityLabels: ["system_owner", "operator", "reviewer"],
    accessibleProgramCount: 4,
    privateRssState: "active",
    lastActivity: "今天 10:30",
    accessSummary: "拥有系统设置、节目、Connector、审核和发布权限。"
  },
  {
    id: "admin_user_reviewer",
    email: "reviewer@example.invalid",
    displayName: "审核员",
    role: "admin",
    status: "active",
    responsibilityLabels: ["reviewer"],
    accessibleProgramCount: 3,
    privateRssState: "active",
    lastActivity: "昨天 18:20",
    accessSummary: "可查看任务和处理审核队列，不管理系统设置。"
  },
  {
    id: "normal_user_active",
    email: "listener@example.invalid",
    displayName: "普通用户",
    role: "user",
    status: "active",
    responsibilityLabels: [],
    accessibleProgramCount: 2,
    privateRssState: "active",
    lastActivity: "今天 08:12",
    accessSummary: "可访问研究周报和原生 RSS 摘要。"
  },
  {
    id: "normal_user_pending",
    email: "pending@example.invalid",
    displayName: "待验证用户",
    role: "user",
    status: "pending_verification",
    responsibilityLabels: [],
    accessibleProgramCount: 0,
    privateRssState: "paused",
    lastActivity: "6月22日",
    accessSummary: "邮箱未验证，不应获得活跃会话。"
  },
  {
    id: "normal_user_suspended",
    email: "suspended@example.invalid",
    displayName: "已暂停用户",
    role: "user",
    status: "suspended",
    responsibilityLabels: [],
    accessibleProgramCount: 1,
    privateRssState: "revoked",
    lastActivity: "6月18日",
    accessSummary: "账号已暂停，私有 RSS 应立即失效。"
  }
];
