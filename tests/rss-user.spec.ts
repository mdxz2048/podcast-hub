import { expect, test } from "@playwright/test";

async function mockAuthApi(page: import("@playwright/test").Page) {
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

test("user can create rotate and revoke mock rss feeds with one-time token display", async ({ page }) => {
  await mockAuthApi(page);
  await page.goto("/rss");

  await expect(page.getByRole("heading", { name: "管理你的私有 RSS Feed" })).toBeVisible();
  await page.getByLabel("新建 Feed 名称").fill("Beta Feed");
  await page.getByRole("button", { name: "创建 Feed" }).click();

  await expect(page.getByText("一次性私有 RSS 链接")).toBeVisible();
  await expect(page.locator("textarea").first()).toHaveValue(/https:\/\/rss\.example\.invalid\/rss\/private\//);
  await page.getByRole("button", { name: "关闭明文展示" }).click();
  await expect(page.getByText("一次性私有 RSS 链接")).toHaveCount(0);

  const feedCard = page.locator("article").filter({ hasText: "每日通勤订阅" }).first();
  await feedCard.getByRole("button", { name: "轮换" }).click();
  await expect(page.getByText("一次性私有 RSS 链接")).toBeVisible();

  await feedCard.getByRole("button", { name: "撤销" }).click();
  await page.getByRole("button", { name: "确认撤销" }).click();
  await expect(feedCard.getByText("Revoked")).toBeVisible();
});