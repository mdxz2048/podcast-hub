import { expect, test } from "@playwright/test";
import type { Page } from "@playwright/test";

type Role = "user" | "admin";
type AuthUser = {
  id: string;
  email: string;
  role: Role;
  status: "active";
  display_name?: string;
};

function buildUser(role: Role): AuthUser {
  return {
    id: role === "admin" ? "admin-1" : "user-1",
    email: role === "admin" ? "admin@example.invalid" : "user@example.invalid",
    role,
    status: "active"
  };
}

async function mockAuthApi(page: Page, initialUser: AuthUser | null = null) {
  let sessionUser = initialUser;
  await page.route("http://127.0.0.1:8080/auth/**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    if (request.method() === "GET" && url.pathname.endsWith("/auth/me")) {
      if (sessionUser) {
        await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ user: sessionUser }) });
        return;
      }
      await route.fulfill({
        status: 401,
        contentType: "application/json",
        body: JSON.stringify({ error: { code: "not_authenticated", message: "当前未登录。" } })
      });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/auth/login")) {
      const body = JSON.parse(request.postData() ?? "{}") as { email?: string };
      sessionUser = body.email?.startsWith("admin") ? buildUser("admin") : buildUser("user");
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ status: "authenticated", user: sessionUser })
      });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/auth/logout")) {
      sessionUser = null;
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ status: "logged_out" }) });
      return;
    }
    await route.fulfill({ status: 404, body: "not mocked" });
  });
}

async function mockConnectorAdminApi(page: Page) {
  const connectors = [{ id: "c1", slug: "example-connector", name: "Example Connector", description: "for test", status: "active", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" }];
  const versions = [{ id: "v1", connector_id: "c1", version: "1.0.0", review_status: "pending_review", runtime_profile: "python-basic", entrypoint: "src/connector.py", manifest: { id: "example-connector", version: "1.0.0" }, package_sha256: "abc", package_size_bytes: 100, validation_summary_json: "{\"is_valid\":true,\"issues\":[]}", created_at: "2026-01-01T00:00:00Z" }];
  const approvedVersions = [{ ...versions[0], id: "v-approved", review_status: "approved", manifest: { id: "example-connector", version: "1.0.0", secrets: [{ name: "session_file" }] } }];
  const sourceDetail = {
    source: { id: "s1", connector_version_id: "v-approved", name: "Test Source", description: "source", status: "draft", trigger_type: "manual", auth_mode: "reusable_session", execution_mode: "unattended", config_json: "{}", network_mode: "trusted_admin", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" },
    secret_bindings: [],
    required_secrets: ["session_file"],
    missing_secrets: ["session_file"]
  };
  const runnableSourceDetail = {
    source: { id: "s-runnable", connector_version_id: "v-approved", name: "Runnable Source", description: "ready", status: "active", trigger_type: "manual", auth_mode: "none", execution_mode: "unattended", config_json: "{}", network_mode: "trusted_admin", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" },
    secret_bindings: [],
    required_secrets: [],
    missing_secrets: []
  };
  const secrets = [{ id: "secret-1", name: "Session file", secret_type: "file", encryption_version: "aes-gcm-v1", created_at: "2026-01-01T00:00:00Z", binding_count: 0 }];
  const importJob = { id: "job-1", connector_source_id: "s1", connector_version_id: "v-approved", status: "queued", trigger_type: "manual", auth_mode: "reusable_session", execution_mode: "unattended", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" };
  const completedJob = { ...importJob, id: "job-completed", status: "completed", finished_at: "2026-01-01T00:10:00Z" };
  const failedJob = { ...importJob, id: "job-failed", status: "failed", failure_code: "fixture_error", failure_message_redacted: "fixture failed" };
  const stagingProgram = { id: "program-1", canonical_key: "source:s1:program-1", title: "候选节目", description: "等待审核的节目描述", author: "Fixture", language: "zh-CN", status: "review_pending", created_from_source_id: "s1", created_from_job_id: "job-completed", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:10:00Z" };
  const stagingEpisode = { id: "episode-1", program_id: "program-1", external_episode_id: "ep-1", title: "候选单集", description: "等待审核的单集描述", published_at: "2026-01-01T00:00:00Z", duration_seconds: 120, status: "review_pending", source_job_id: "job-completed", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:10:00Z" };
  let adminProgram = { ...stagingProgram };
  let adminEpisode = { ...stagingEpisode };
  let reviews = [
    { id: "review-program", target_type: "program", target_id: "program-1", review_kind: "metadata", status: "pending", review_note: "", created_at: "2026-01-01T00:10:00Z" },
    { id: "review-episode", target_type: "episode", target_id: "episode-1", review_kind: "metadata", status: "pending", review_note: "", created_at: "2026-01-01T00:10:00Z" }
  ];
  let grants = [{ id: "grant-1", user_id: "user-1", program_id: "program-1", status: "active", reason: "beta", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" }];
  let rssFeeds = [{ id: "feed-1", user_id: "user-1", user_email_hint: "u***", name: "Private Feed", token_prefix: "tok_pref", status: "active", created_at: "2026-01-01T00:00:00Z", last_used_at: "2026-01-01T00:05:00Z" }];
  await page.route("http://127.0.0.1:8080/admin/**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    if (request.method() === "GET" && url.pathname.endsWith("/admin/connectors")) {
      const body = JSON.stringify({ connectors: url.searchParams.get("empty") === "1" ? [] : connectors });
      await route.fulfill({ status: 200, contentType: "application/json", body });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/system/status")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ api: "ok", runner: { mode: "disabled", can_run_jobs: false, code: "runner_disabled", reason: "RUNNER_MODE=disabled; queued Import Jobs will not execute until a separate Runner is started." } })
      });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/connectors/c1")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ connector: connectors[0] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/connectors/c1/versions")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ versions: [...versions, ...approvedVersions] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/connector-versions/v1")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ version: versions[0], validation_summary: { is_valid: true, issues: [] } }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/connectors/upload")) {
      const contentType = (await request.headerValue("content-type")) ?? "";
      if (!contentType.includes("multipart/form-data")) {
        await route.fulfill({ status: 400, contentType: "application/json", body: JSON.stringify({ error: { code: "invalid_upload", message: "bad form" } }) });
        return;
      }
      if ((request.postData() ?? "").includes("invalid")) {
        await route.fulfill({
          status: 201,
          contentType: "application/json",
          body: JSON.stringify({
            connector: connectors[0],
            version: versions[0],
            validation_summary: { is_valid: false, issues: [{ code: "manifest_invalid", message: "manifest 失败", path: "manifest.yaml" }] }
          })
        });
        return;
      }
      await route.fulfill({
        status: 201,
        contentType: "application/json",
        body: JSON.stringify({ connector: connectors[0], version: versions[0], validation_summary: { is_valid: true, issues: [] } })
      });
      return;
    }
    if (request.method() === "POST" && url.pathname.includes("/admin/connector-versions/") && (url.pathname.endsWith("/approve") || url.pathname.endsWith("/reject") || url.pathname.endsWith("/disable"))) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ version: versions[0] }) });
      return;
    }
    if (request.method() === "POST" && (url.pathname.endsWith("/admin/connectors/c1/enable") || url.pathname.endsWith("/admin/connectors/c1/disable"))) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ connector: connectors[0] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/sources")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ sources: [] }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/sources")) {
      await route.fulfill({ status: 201, contentType: "application/json", body: JSON.stringify(sourceDetail) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/sources/s1")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify(sourceDetail) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/sources/s-runnable")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify(runnableSourceDetail) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/sources/s-runnable/import-jobs")) {
      await route.fulfill({ status: 201, contentType: "application/json", body: JSON.stringify({ job: { ...importJob, connector_source_id: "s-runnable" } }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/sources/s1/enable")) {
      await route.fulfill({ status: 409, contentType: "application/json", body: JSON.stringify({ error: { code: "missing_required_secrets", message: "missing" } }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/secrets")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ secrets }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/secrets/text")) {
      await route.fulfill({ status: 201, contentType: "application/json", body: JSON.stringify({ secret: secrets[0] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/review")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ reviews }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/review/review-program")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ review: reviews[0] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/review/review-episode")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ review: reviews[1] }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/review/review-program/approve")) {
      reviews = reviews.map((review) => review.id === "review-program" ? { ...review, status: "approved", reviewed_at: "2026-01-01T00:20:00Z" } : review);
      adminProgram = { ...adminProgram, status: "approved", updated_at: "2026-01-01T00:20:00Z" };
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ review: reviews[0] }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/review/review-episode/approve")) {
      reviews = reviews.map((review) => review.id === "review-episode" ? { ...review, status: "approved", reviewed_at: "2026-01-01T00:22:00Z" } : review);
      adminEpisode = { ...adminEpisode, status: "approved", updated_at: "2026-01-01T00:22:00Z" };
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ review: reviews[1] }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/review/review-program/reject")) {
      const body = JSON.parse(request.postData() ?? "{}") as { reason?: string };
      if (!body.reason?.trim()) {
        await route.fulfill({ status: 400, contentType: "application/json", body: JSON.stringify({ error: { code: "review_reason_required", message: "拒绝审核必须提供原因。" } }) });
        return;
      }
      reviews = reviews.map((review) => review.id === "review-program" ? { ...review, status: "rejected", review_note: body.reason ?? "", reviewed_at: "2026-01-01T00:21:00Z" } : review);
      adminProgram = { ...adminProgram, status: "rejected" };
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ review: reviews[0] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/programs")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ programs: [adminProgram] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/rss-feeds")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ feeds: rssFeeds }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/rss-feeds/feed-1/revoke")) {
      rssFeeds = rssFeeds.map((feed) => feed.id === "feed-1" ? { ...feed, status: "revoked", revoked_at: "2026-01-01T00:20:00Z" } : feed);
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ feed: rssFeeds[0] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/programs/program-1")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ program: adminProgram, episodes: [adminEpisode] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/programs/program-1/access-grants")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ grants }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/programs/program-1/access-grants")) {
      const body = JSON.parse(request.postData() ?? "{}") as { email?: string; reason?: string };
      if (!body.email?.includes("@")) {
        await route.fulfill({ status: 409, contentType: "application/json", body: JSON.stringify({ error: { code: "user_not_eligible", message: "User is not eligible for access." } }) });
        return;
      }
      const grant = { id: "grant-2", user_id: "user-2", program_id: "program-1", status: "active", reason: body.reason ?? "", created_at: "2026-01-01T00:30:00Z", updated_at: "2026-01-01T00:30:00Z" };
      grants = [grant, ...grants];
      await route.fulfill({ status: 201, contentType: "application/json", body: JSON.stringify({ grant }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/program-access/grant-1/revoke")) {
      grants = grants.map((grant) => grant.id === "grant-1" ? { ...grant, status: "revoked", revoked_at: "2026-01-01T00:31:00Z" } : grant);
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ grant: grants.find((grant) => grant.id === "grant-1") }) });
      return;
    }
    if (request.method() === "PATCH" && url.pathname.endsWith("/admin/programs/program-1")) {
      const body = JSON.parse(request.postData() ?? "{}") as { title?: string; description?: string };
      adminProgram = { ...adminProgram, title: body.title ?? adminProgram.title, description: body.description ?? adminProgram.description };
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ program: adminProgram }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/programs/program-1/publish")) {
      if (adminProgram.status !== "approved") {
        await route.fulfill({ status: 409, contentType: "application/json", body: JSON.stringify({ error: { code: "publish_precondition_failed", message: "发布前置条件未满足。" } }) });
        return;
      }
      adminProgram = { ...adminProgram, status: "published", updated_at: "2026-01-01T00:23:00Z" };
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ program: adminProgram }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/programs/program-1/archive")) {
      adminProgram = { ...adminProgram, status: "archived" };
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ program: adminProgram }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/programs/program-1/submit-review")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ review: reviews[0] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/episodes/episode-1")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ episode: adminEpisode }) });
      return;
    }
    if (request.method() === "PATCH" && url.pathname.endsWith("/admin/episodes/episode-1")) {
      const body = JSON.parse(request.postData() ?? "{}") as { title?: string; description?: string };
      adminEpisode = { ...adminEpisode, title: body.title ?? adminEpisode.title, description: body.description ?? adminEpisode.description };
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ episode: adminEpisode }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/episodes/episode-1/publish")) {
      if (adminEpisode.status !== "approved" || adminProgram.status !== "published") {
        await route.fulfill({ status: 409, contentType: "application/json", body: JSON.stringify({ error: { code: "publish_precondition_failed", message: "发布前置条件未满足。" } }) });
        return;
      }
      adminEpisode = { ...adminEpisode, status: "published" };
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ episode: adminEpisode }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/episodes/episode-1/archive")) {
      adminEpisode = { ...adminEpisode, status: "archived" };
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ episode: adminEpisode }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/episodes/episode-1/submit-review")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ review: reviews[1] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ jobs: [] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-1")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ job: importJob }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-1/events")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ events: [{ id: "evt-1", import_job_id: "job-1", event_type: "job.queued", level: "info", message_redacted: "queued token=[redacted]", metadata_redacted: "{\"authorization\":\"[redacted]\"}", created_at: "2026-01-01T00:00:00Z" }] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-1/artifacts")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ artifacts: [{ id: "artifact-1", import_job_id: "job-1", artifact_type: "episode_metadata", relative_path: "episodes/episode-001.json", size_bytes: 42, sha256: "0".repeat(64), created_at: "2026-01-01T00:00:00Z" }] }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/import-jobs/job-1/cancel")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ job: { ...importJob, status: "cancelled" } }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-1/intake-status")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ intake_run: null }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-completed")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ job: completedJob }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-completed/events")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ events: [] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-completed/artifacts")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ artifacts: [{ id: "bundle-1", import_job_id: "job-completed", artifact_type: "metadata_bundle", relative_path: "bundle.json", size_bytes: 92, sha256: "1".repeat(64), created_at: "2026-01-01T00:00:00Z" }] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-completed/intake-status")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ intake_run: null }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/import-jobs/job-completed/intake")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ intake_run: { id: "intake-1", import_job_id: "job-completed", status: "succeeded", validation_issues_redacted: "[]", program_id: "program-1", created_at: "2026-01-01T00:10:00Z", updated_at: "2026-01-01T00:10:00Z" }, program: stagingProgram }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-invalid")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ job: { ...completedJob, id: "job-invalid" } }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-invalid/events")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ events: [] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-invalid/artifacts")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ artifacts: [{ id: "bundle-bad", import_job_id: "job-invalid", artifact_type: "metadata_bundle", relative_path: "bundle.json", size_bytes: 92, sha256: "2".repeat(64), created_at: "2026-01-01T00:00:00Z" }] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-invalid/intake-status")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ intake_run: { id: "intake-bad", import_job_id: "job-invalid", status: "failed", validation_issues_redacted: "[\"metadata_bundle schema is invalid\"]", created_at: "2026-01-01T00:10:00Z", updated_at: "2026-01-01T00:10:00Z" } }) });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/admin/import-jobs/job-invalid/intake")) {
      await route.fulfill({ status: 422, contentType: "application/json", body: JSON.stringify({ error: { code: "metadata_bundle_invalid", message: "metadata bundle 校验失败。", validation_issues: ["metadata_bundle schema is invalid"] } }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-failed")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ job: failedJob }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-failed/events")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ events: [] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-failed/artifacts")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ artifacts: [] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/import-jobs/job-failed/intake-status")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ intake_run: null }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/staging/programs")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ programs: url.searchParams.get("empty") === "1" ? [] : [stagingProgram] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/staging/programs/program-1")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ program: stagingProgram }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/staging/episodes")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ episodes: url.searchParams.get("empty") === "1" ? [] : [stagingEpisode] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname.endsWith("/admin/staging/episodes/episode-1")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ episode: stagingEpisode }) });
      return;
    }
    await route.fulfill({ status: 404, body: "not mocked" });
  });
}

async function mockUserCatalogApi(page: Page) {
  const program = { id: "program-1", title: "Program Title", description: "Authorized program", author: "Author", language: "zh-CN", status: "published", episode_count: 1, updated_at: "2026-01-01T00:00:00Z" };
  const episode = { id: "episode-1", program_id: "program-1", title: "Hello Episode", description: "Published episode", published_at: "2026-01-01T00:00:00Z", duration_seconds: 120, status: "published", media_status: "published" };
  let collections = [{ id: "collection-1", title: "Commute", description: "Daily", programs: [program], created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" }];
  await page.route("http://127.0.0.1:8080/programs**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    if (request.method() === "GET" && url.pathname === "/programs") {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ programs: [program] }) });
      return;
    }
    if (request.method() === "GET" && url.pathname === "/programs/program-1") {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ program }) });
      return;
    }
    if (request.method() === "GET" && url.pathname === "/programs/program-1/episodes") {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ episodes: [episode] }) });
      return;
    }
    await route.fulfill({ status: 404, contentType: "application/json", body: JSON.stringify({ error: { code: "resource_not_found", message: "Resource not found." } }) });
  });
  await page.route("http://127.0.0.1:8080/episodes/**", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ episode }) });
  });
  await page.route("http://127.0.0.1:8080/me/collections**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    if (request.method() === "GET" && url.pathname === "/me/collections") {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ collections }) });
      return;
    }
    if (request.method() === "POST" && url.pathname === "/me/collections") {
      const body = JSON.parse(request.postData() ?? "{}") as { title?: string; description?: string };
      const collection = { id: "collection-2", title: body.title ?? "New", description: body.description ?? "", programs: [], created_at: "2026-01-01T00:10:00Z", updated_at: "2026-01-01T00:10:00Z" };
      collections = [collection, ...collections];
      await route.fulfill({ status: 201, contentType: "application/json", body: JSON.stringify({ collection }) });
      return;
    }
    if (request.method() === "PATCH" && url.pathname === "/me/collections/collection-1") {
      const body = JSON.parse(request.postData() ?? "{}") as { title?: string; description?: string };
      collections = collections.map((collection) => collection.id === "collection-1" ? { ...collection, title: body.title ?? collection.title, description: body.description ?? collection.description } : collection);
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ collection: collections.find((collection) => collection.id === "collection-1") }) });
      return;
    }
    if (request.method() === "POST" && url.pathname === "/me/collections/collection-1/programs") {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ collection: collections[0] }) });
      return;
    }
    if (request.method() === "DELETE" && url.pathname === "/me/collections/collection-1/programs/program-1") {
      collections = collections.map((collection) => collection.id === "collection-1" ? { ...collection, programs: [] } : collection);
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ collection: collections.find((collection) => collection.id === "collection-1") }) });
      return;
    }
    await route.fulfill({ status: 404, body: "not mocked" });
  });
}

test("admin login redirects to /admin", async ({ page }) => {
  await mockAuthApi(page);
  await page.goto("/login");
  await page.getByLabel("邮箱").fill("admin@example.invalid");
  await page.getByLabel("密码").fill("Password123!");
  await page.getByRole("button", { name: "登录" }).click();
  await expect(page).toHaveURL(/\/admin$/);
});

test("user login stays on user path", async ({ page }) => {
  await mockAuthApi(page);
  await page.goto("/login");
  await page.getByLabel("邮箱").fill("user@example.invalid");
  await page.getByLabel("密码").fill("Password123!");
  await page.getByRole("button", { name: "登录" }).click();
  await expect(page).toHaveURL(/\/programs$/);
});

test("unauthenticated access to /admin redirects to /login", async ({ page }) => {
  await mockAuthApi(page);
  await page.goto("/admin");
  await expect(page).toHaveURL(/\/login$/);
});

test("user role gets permission denied on /admin", async ({ page }) => {
  await mockAuthApi(page, buildUser("user"));
  await page.goto("/admin");
  await expect(page.getByText("需要权限")).toBeVisible();
});

test("logout clears admin session state", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await page.goto("/admin");
  await page.getByRole("button", { name: "退出登录" }).click();
  await expect(page).toHaveURL(/\/login$/);
  await page.goto("/admin");
  await expect(page).toHaveURL(/\/login$/);
});

test("auth state is not persisted in local/session storage", async ({ page }) => {
  await mockAuthApi(page);
  await page.goto("/login");
  await page.getByLabel("邮箱").fill("admin@example.invalid");
  await page.getByLabel("密码").fill("Password123!");
  await page.getByRole("button", { name: "登录" }).click();
  const storageCounts = await page.evaluate(() => ({
    local: window.localStorage.length,
    session: window.sessionStorage.length
  }));
  expect(storageCounts.local).toBe(0);
  expect(storageCounts.session).toBe(0);
});

test("user catalog and program detail use real API data", async ({ page }) => {
  await mockAuthApi(page, buildUser("user"));
  await mockUserCatalogApi(page);

  await page.goto("/programs");
  await expect(page.getByText("Program Title")).toBeVisible();
  await expect(page.getByText("Source")).toHaveCount(0);
  await expect(page.getByText("Import Job")).toHaveCount(0);

  await page.goto("/programs/program-1");
  await expect(page.getByRole("heading", { name: "Program Title" })).toBeVisible();
  await expect(page.getByText("Hello Episode")).toBeVisible();
  await expect(page.getByRole("button", { name: "下载" })).toHaveCount(0);
});

test("user collections create edit and remove authorized programs", async ({ page }) => {
  await mockAuthApi(page, buildUser("user"));
  await mockUserCatalogApi(page);

  await page.goto("/collections");
  await expect(page.getByText("Commute")).toBeVisible();
  await page.getByLabel("新建合集").fill("Focus");
  await page.getByRole("button", { name: "创建合集" }).click();
  await expect(page.getByText("新建合集已保存。")).toBeVisible();

  await page.goto("/collections/collection-1");
  await expect(page.getByText("Program Title")).toBeVisible();
  await page.getByLabel("合集名称").fill("Commute Updated");
  await page.getByRole("button", { name: "保存" }).click();
  await expect(page.getByText("合集设置已保存。")).toBeVisible();
  await page.getByRole("button", { name: "移除" }).click();
  await expect(page.getByText("合集里还没有节目")).toBeVisible();
  await expect(page.getByText("storage_key")).toHaveCount(0);
});

test("admin connectors list shows empty state", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.route("http://127.0.0.1:8080/admin/connectors", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ connectors: [] }) });
  });
  await page.goto("/admin/connectors");
  await expect(page.getByText("暂无 Connector")).toBeVisible();
});

test("admin can upload zip and see pending review", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/connectors/new");
  await page.getByLabel("Connector ID（slug）").fill("example-connector");
  await page.getByLabel("Version（semver）").fill("1.0.0");
  await page.setInputFiles("input[type=file]", { name: "connector.zip", mimeType: "application/zip", buffer: Buffer.from("ok") });
  await page.getByRole("button", { name: "上传并校验" }).click();
  await expect(page).toHaveURL(/\/admin\/connectors\/c1$/);
});

test("invalid upload shows validation issues", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/connectors/new");
  await page.getByLabel("Connector ID（slug）").fill("example-connector");
  await page.getByLabel("Version（semver）").fill("1.0.0");
  await page.setInputFiles("input[type=file]", { name: "invalid.zip", mimeType: "application/zip", buffer: Buffer.from("invalid") });
  await page.getByRole("button", { name: "上传并校验" }).click();
  await expect(page.getByText("校验失败")).toBeVisible();
});

test("version actions are visible and runtime buttons are absent", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/connectors/c1");
  await expect(page.getByRole("button", { name: "Approve", exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "Reject", exact: true })).toBeVisible();
  await expect(page.getByRole("button", { name: "运行" })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "下载媒体" })).toHaveCount(0);
});

test("admin source pages use real API empty and secret metadata states", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/sources");
  await expect(page.getByText("暂无 Source")).toBeVisible();
  await page.goto("/admin/secrets");
  await expect(page.getByText("Session file")).toBeVisible();
  await expect(page.getByText("plain-secret-value")).toHaveCount(0);
});

test("admin can create source from approved connector version", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/sources/new");
  await page.getByLabel("Source 名称").fill("Test Source");
  await page.getByLabel("描述").fill("source");
  await page.getByLabel("Auth Mode").selectOption("reusable_session");
  await page.getByLabel("Network Mode").selectOption("trusted_admin");
  await page.getByRole("button", { name: "创建 Draft Source" }).click();
  await expect(page).toHaveURL(/\/admin\/sources\/s1$/);
  await expect(page.getByText("missing")).toBeVisible();
  await expect(page.getByRole("button", { name: "Enable Source" })).toBeDisabled();
  await expect(page.getByRole("button", { name: "运行" })).toHaveCount(0);
  await expect(page.getByRole("button", { name: "下载媒体" })).toHaveCount(0);
});

test("admin import jobs use real API metadata states", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/import-jobs");
  await expect(page.getByText("暂无 Import Job")).toBeVisible();
  await expect(page.getByText("Runner disabled")).toBeVisible();
  await page.goto("/admin/import-jobs/job-1");
  await expect(page.getByText("job.queued")).toBeVisible();
  await expect(page.getByText("任务完成后才能导入待审核区。")).toBeVisible();
  await expect(page.getByText("queued token=[redacted]")).toBeVisible();
  await expect(page.getByText("episodes/episode-001.json")).toBeVisible();
  await page.getByRole("button", { name: "取消任务" }).click();
  await expect(page.getByText("取消请求已记录。")).toBeVisible();
  await expect(page.getByText("发布")).toHaveCount(0);
  await expect(page.getByText("RSS")).toHaveCount(0);
  await expect(page.getByRole("button", { name: "下载媒体" })).toHaveCount(0);
});

test("completed import job can intake into staging", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/import-jobs/job-completed");
  await expect(page.getByText("可导入待审核区。")).toBeVisible();
  await page.getByRole("button", { name: "导入到待审核区" }).click();
  await expect(page.getByText("已导入待审核区：候选节目")).toBeVisible();
  await expect(page.getByText("发布")).toHaveCount(0);
  await expect(page.getByText("RSS")).toHaveCount(0);
  await expect(page.getByRole("button", { name: "下载媒体" })).toHaveCount(0);
});

test("failed import job cannot intake", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/import-jobs/job-failed");
  await expect(page.getByText("失败任务不能导入待审核区。")).toBeVisible();
  await expect(page.getByRole("button", { name: "导入到待审核区" })).toBeDisabled();
});

test("intake validation issues are visible without leaking paths", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/import-jobs/job-invalid");
  await page.getByRole("button", { name: "导入到待审核区" }).click();
  await expect(page.getByText("metadata_bundle schema is invalid")).toBeVisible();
  await expect(page.getByText("/Users/")).toHaveCount(0);
  await expect(page.getByText("storage_key")).toHaveCount(0);
});

test("admin staging uses real API empty and detail states", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/staging");
  await expect(page.getByText("候选节目")).toBeVisible();
  await expect(page.getByText("候选单集")).toBeVisible();
  await expect(page.getByText("发布")).toHaveCount(0);
  await expect(page.getByText("RSS")).toHaveCount(0);
  await expect(page.getByRole("button", { name: "下载媒体" })).toHaveCount(0);

  await page.goto("/admin/staging/programs/program-1");
  await expect(page.getByText("审核前状态")).toBeVisible();
  await expect(page.getByText("候选节目")).toBeVisible();

  await page.goto("/admin/staging/episodes/episode-1");
  await expect(page.getByText("媒体仍为私有 staging metadata")).toBeVisible();
  await expect(page.getByText("候选单集")).toBeVisible();
});

test("admin staging empty state is driven by API", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.route("http://127.0.0.1:8080/admin/staging/programs", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ programs: [] }) });
  });
  await page.route("http://127.0.0.1:8080/admin/staging/episodes", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ episodes: [] }) });
  });
  await page.goto("/admin/staging");
  await expect(page.getByText("待审核区暂无内容")).toBeVisible();
});

test("admin review workflow enforces reason and publish preconditions", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);

  await page.goto("/admin/reviews");
  await expect(page.getByText("review-program")).toBeVisible();
  await page.getByRole("button", { name: "拒绝" }).first().click();
  await expect(page.getByText("拒绝审核必须提供原因。")).toBeVisible();

  await page.goto("/admin/programs/program-1");
  await expect(page.getByText("只有 approved Program 可以发布。")).toBeVisible();
  await expect(page.getByRole("button", { name: "发布" })).toBeDisabled();

  await page.goto("/admin/reviews");
  await page.getByRole("button", { name: "通过" }).first().click();
  await expect(page.getByText("审核已通过。")).toBeVisible();

  await page.goto("/admin/programs/program-1");
  await expect(page.getByRole("button", { name: "发布" })).toBeEnabled();
  await page.getByRole("button", { name: "发布" }).click();
  await expect(page.getByText("Program 已发布。")).toBeVisible();
  await expect(page.getByText("RSS")).toHaveCount(0);
  await expect(page.getByRole("button", { name: "下载媒体" })).toHaveCount(0);
});

test("admin can grant and revoke program access through real API contract", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);

  await page.goto("/admin/programs/program-1");
  await expect(page.getByText("用户授权")).toBeVisible();
  await expect(page.getByText("User user-1")).toBeVisible();
  await page.getByLabel("用户邮箱").fill("other@example.invalid");
  await page.getByLabel("授权原因").fill("beta");
  await page.getByRole("button", { name: "授予访问" }).click();
  await expect(page.getByText("授权已写入。")).toBeVisible();
  await page.locator("article").filter({ hasText: "User user-1" }).getByRole("button", { name: "撤销" }).click();
  await expect(page.getByText("授权已撤销。")).toBeVisible();
  await expect(page.getByText("token")).toHaveCount(0);
});

test("admin can revoke rss feeds without plaintext token access", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);

  await page.goto("/admin/rss-feeds");
  await expect(page.getByText("Private Feed")).toBeVisible();
  await expect(page.getByText("tok_pref")).toBeVisible();
  await expect(page.getByText("new-token-secret")).toHaveCount(0);
  await page.getByRole("button", { name: "撤销" }).click();
  await page.getByRole("button", { name: "确认撤销" }).click();
  await expect(page.getByText("RSS Feed 已撤销。")).toBeVisible();
  await expect(page.getByText("revoked", { exact: true })).toBeVisible();
});

test("admin episode review publish and archive use real API", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);

  await page.goto("/admin/reviews");
  await page.getByRole("button", { name: "通过" }).first().click();
  await page.goto("/admin/programs/program-1");
  await page.getByRole("button", { name: "发布" }).click();
  await expect(page.getByText("Program 已发布。")).toBeVisible();

  await page.goto("/admin/episodes/episode-1");
  await expect(page.getByText("只有 approved Episode 可以发布。")).toBeVisible();
  await expect(page.getByRole("button", { name: "发布" })).toBeDisabled();

  await page.goto("/admin/review/review-episode");
  await page.getByRole("button", { name: "通过" }).click();
  await expect(page.getByText("审核已通过。")).toBeVisible();

  await page.goto("/admin/episodes/episode-1");
  await page.getByRole("button", { name: "发布" }).click();
  await expect(page.getByText("Episode 已发布。")).toBeVisible();
  await page.getByRole("button", { name: "归档" }).click();
  await expect(page.getByText("Episode 已归档。")).toBeVisible();
  await expect(page.getByText("RSS")).toHaveCount(0);
  await expect(page.getByRole("button", { name: "下载媒体" })).toHaveCount(0);
});

test("admin review empty state is driven by API", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.route("http://127.0.0.1:8080/admin/review", async (route) => {
    await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ reviews: [] }) });
  });
  await page.goto("/admin/reviews");
  await expect(page.getByText("当前没有审核项")).toBeVisible();
});

test("admin can create manual import job from runnable source", async ({ page }) => {
  await mockAuthApi(page, buildUser("admin"));
  await mockConnectorAdminApi(page);
  await page.goto("/admin/sources/s-runnable");
  await page.getByRole("button", { name: "创建手动任务" }).click();
  await expect(page.getByText("Import Job 已创建：job-1")).toBeVisible();
});
