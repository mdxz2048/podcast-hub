import { expect, test } from "@playwright/test";
import { mkdirSync } from "node:fs";

const pages = [
  { name: "home", path: "/" },
  { name: "register", path: "/register?state=success" },
  { name: "login", path: "/login?state=error" },
  { name: "programs", path: "/programs?state=long" },
  { name: "admin-overview", path: "/admin?state=success" },
  { name: "admin-programs", path: "/admin/programs?state=long" },
  { name: "components-states", path: "/components" },
  { name: "permission-denied", path: "/programs?state=denied" },
  { name: "empty-state", path: "/admin/programs?state=empty" },
  { name: "loading-state", path: "/admin?state=loading" }
];

test.describe("M0.1 screenshots", () => {
  for (const item of pages) {
    test(`${item.name}`, async ({ page }, testInfo) => {
      await page.goto(item.path);
      await expect(page.locator("body")).toBeVisible();
      mkdirSync(`screenshots/${testInfo.project.name}`, { recursive: true });
      await page.screenshot({
        path: `screenshots/${testInfo.project.name}/${item.name}.png`,
        fullPage: true
      });
    });
  }
});

