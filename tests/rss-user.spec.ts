import { expect, test } from "@playwright/test";
import type { Page } from "@playwright/test";

async function mockAuthApi(page: Page) {
  await page.route("http://127.0.0.1:8080/auth/**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    if (request.method() === "GET" && url.pathname.endsWith("/auth/me")) {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({ user: { id: "user-1", email: "user@example.invalid", role: "user", status: "active" } })
      });
      return;
    }
    if (request.method() === "POST" && url.pathname.endsWith("/auth/logout")) {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ status: "logged_out" }) });
      return;
    }
    await route.fulfill({ status: 404, body: "not mocked" });
  });
}

async function mockRssFeedsApi(page: Page) {
  let feeds: Array<Record<string, string | undefined>> = [{ id: "feed-1", user_id: "user-1", name: "每日通勤订阅", token_prefix: "tok_pref", status: "active", created_at: "2026-01-01T00:00:00Z", last_used_at: undefined, expires_at: undefined }];
  await page.route("http://127.0.0.1:8080/me/rss-feeds**", async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    if (request.method() === "GET" && url.pathname === "/me/rss-feeds") {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ feeds }) });
      return;
    }
    if (request.method() === "POST" && url.pathname === "/me/rss-feeds") {
      const body = JSON.parse(request.postData() ?? "{}") as { name?: string };
      const feed = { id: "feed-2", user_id: "user-1", name: body.name || "Beta Feed", token_prefix: "new_pref", status: "active", created_at: "2026-01-01T00:10:00Z" };
      feeds = [feed, ...feeds];
      await route.fulfill({ status: 201, contentType: "application/json", body: JSON.stringify({ feed, token: "new-token-secret", feed_url: "https://rss.example.invalid/rss/private/new-token-secret.xml" }) });
      return;
    }
    if (request.method() === "POST" && url.pathname === "/me/rss-feeds/feed-1/rotate") {
      feeds = feeds.map((feed) => feed.id === "feed-1" ? { ...feed, token_prefix: "rot_pref", rotated_at: "2026-01-01T00:20:00Z" } : feed);
      const feed = feeds.find((item) => item.id === "feed-1");
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ feed, token: "rotated-token-secret", feed_url: "https://rss.example.invalid/rss/private/rotated-token-secret.xml" }) });
      return;
    }
    if (request.method() === "POST" && url.pathname === "/me/rss-feeds/feed-1/revoke") {
      feeds = feeds.map((feed) => feed.id === "feed-1" ? { ...feed, status: "revoked", revoked_at: "2026-01-01T00:30:00Z" } : feed);
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify({ feed: feeds.find((feed) => feed.id === "feed-1") }) });
      return;
    }
    await route.fulfill({ status: 404, body: "not mocked" });
  });
}

test("user can create rotate and revoke real rss feeds with one-time token display", async ({ page }) => {
  await mockAuthApi(page);
  await mockRssFeedsApi(page);
  await page.goto("/rss");

  await expect(page.getByRole("heading", { name: "管理你的私有 RSS Feed" })).toBeVisible();
  await page.getByLabel("新建 Feed 名称").fill("Beta Feed");
  await page.getByRole("button", { name: "创建 Feed" }).click();

  await expect(page.getByText("一次性私有 RSS 链接")).toBeVisible();
  await expect(page.locator("textarea").first()).toHaveValue(/https:\/\/rss\.example\.invalid\/rss\/private\//);
  await page.getByRole("button", { name: "关闭明文展示" }).click();
  await expect(page.getByText("一次性私有 RSS 链接")).toHaveCount(0);
  await expect(page.getByText("new-token-secret")).toHaveCount(0);

  const feedCard = page.locator("article").filter({ hasText: "每日通勤订阅" }).first();
  await feedCard.getByRole("button", { name: "轮换" }).click();
  await expect(page.getByText("一次性私有 RSS 链接")).toBeVisible();

  await feedCard.getByRole("button", { name: "撤销" }).click();
  await page.getByRole("button", { name: "确认撤销" }).click();
  await expect(feedCard.getByText("Revoked")).toBeVisible();
  const storage = await page.evaluate(() => ({ local: localStorage.length, session: sessionStorage.length }));
  expect(storage).toEqual({ local: 0, session: 0 });
});
