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
  const secrets = [{ id: "secret-1", name: "Session file", secret_type: "file", encryption_version: "aes-gcm-v1", created_at: "2026-01-01T00:00:00Z", binding_count: 0 }];
  await page.route("http://127.0.0.1:8080/admin/**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    if (request.method() === "GET" && url.pathname.endsWith("/admin/connectors")) {
      const body = JSON.stringify({ connectors: url.searchParams.get("empty") === "1" ? [] : connectors });
      await route.fulfill({ status: 200, contentType: "application/json", body });
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
    if (request.method() === "POST" && (url.pathname.endsWith("/approve") || url.pathname.endsWith("/reject") || url.pathname.endsWith("/disable"))) {
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
