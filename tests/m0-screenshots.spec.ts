import { expect, test } from "@playwright/test";
import { mkdirSync } from "node:fs";

const pages = [
  { name: "home", path: "/" },
  { name: "register", path: "/register?state=success" },
  { name: "register-verify", path: "/register/verify?state=success" },
  { name: "login", path: "/login?state=error" },
  { name: "forgot-password", path: "/forgot-password?state=success" },
  { name: "programs", path: "/programs?state=long" },
  { name: "program-detail", path: "/programs/program_editorial_signal" },
  { name: "program-detail-restricted", path: "/programs/program_empty" },
  { name: "program-detail-long", path: "/programs/program_editorial_signal?state=long" },
  { name: "program-detail-drawer", path: "/programs/program_editorial_signal?drawer=add" },
  { name: "collections", path: "/collections" },
  { name: "collections-empty", path: "/collections?state=empty" },
  { name: "collection-editor", path: "/collections/collection_research_weekly" },
  { name: "collection-editor-empty", path: "/collections/collection_research_weekly?state=empty" },
  { name: "collection-editor-preview-empty", path: "/collections/collection_research_weekly?state=preview_empty" },
  { name: "collection-editor-denied", path: "/collections/collection_research_weekly?state=denied" },
  { name: "subscribe", path: "/collections/collection_research_weekly/subscribe" },
  { name: "subscribe-toast", path: "/collections/collection_research_weekly/subscribe?toast=success" },
  { name: "subscribe-dialog", path: "/collections/collection_research_weekly/subscribe?dialog=reset" },
  { name: "admin-overview", path: "/admin?state=success" },
  { name: "admin-programs", path: "/admin/programs?state=long" },
  { name: "admin-program-detail", path: "/admin/programs/program_editorial_signal" },
  { name: "admin-program-detail-auth", path: "/admin/programs/program_editorial_signal?state=auth" },
  { name: "admin-program-detail-draft", path: "/admin/programs/program_editorial_signal?state=draft" },
  { name: "admin-connectors", path: "/admin/connectors" },
  { name: "admin-connector-detail", path: "/admin/connectors/connector_partner_archive" },
  { name: "admin-connector-new", path: "/admin/connectors/new" },
  { name: "admin-import-jobs", path: "/admin/import-jobs" },
  { name: "admin-import-job-running", path: "/admin/import-jobs/job_mock_1004" },
  { name: "admin-import-job-waiting-auth", path: "/admin/import-jobs/job_mock_1003?state=waiting_auth" },
  { name: "admin-import-job-failed", path: "/admin/import-jobs/job_mock_1003" },
  { name: "admin-reviews", path: "/admin/reviews" },
  { name: "admin-reviews-empty", path: "/admin/reviews?state=empty" },
  { name: "admin-users", path: "/admin/users" },
  { name: "components-states", path: "/components" },
  { name: "permission-denied", path: "/programs?state=denied" },
  { name: "empty-state", path: "/admin/programs?state=empty" },
  { name: "loading-state", path: "/admin?state=loading" }
];

const overlayPages = [
  { name: "admin-review-drawer-viewport-overlay", path: "/admin/reviews?drawer=review_mock_002" },
  { name: "admin-job-cancel-dialog-viewport-overlay", path: "/admin/import-jobs/job_mock_1004?dialog=cancel" },
  { name: "admin-review-reject-dialog-viewport-overlay", path: "/admin/reviews?dialog=reject&drawer=review_mock_002" },
  { name: "admin-users-drawer-viewport-overlay", path: "/admin/users?drawer=normal_user_suspended" },
  { name: "admin-import-job-waiting-auth-viewport-overlay", path: "/admin/import-jobs/job_mock_1003?state=waiting_auth" },
  { name: "admin-connector-success-toast-viewport-overlay", path: "/admin/connectors/connector_partner_archive?toast=success" }
];

test.describe("M0 screenshots", () => {
  for (const item of pages) {
    test(`${item.name}`, async ({ page }, testInfo) => {
      await page.goto(item.path);
      await expect(page.locator("body")).toBeVisible();
      await page.waitForTimeout(150);
      mkdirSync(`screenshots/${testInfo.project.name}`, { recursive: true });
      await page.screenshot({
        path: `screenshots/${testInfo.project.name}/${item.name}.png`,
        fullPage: true
      });
    });
  }

  for (const item of overlayPages) {
    test(`${item.name}`, async ({ page }, testInfo) => {
      await page.goto(item.path);
      await expect(page.locator("body")).toBeVisible();
      await page.waitForTimeout(150);
      mkdirSync(`screenshots/${testInfo.project.name}`, { recursive: true });
      await page.screenshot({
        path: `screenshots/${testInfo.project.name}/${item.name}.png`,
        fullPage: false
      });
    });
  }
});
