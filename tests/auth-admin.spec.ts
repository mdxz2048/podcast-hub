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
      if (body.email?.startsWith("admin")) {
        sessionUser = buildUser("admin");
      } else {
        sessionUser = buildUser("user");
      }
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ status: "authenticated", user: sessionUser })
      });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/auth/logout")) {
      sessionUser = null;
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ status: "logged_out" })
      });
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
