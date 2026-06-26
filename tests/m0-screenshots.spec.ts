import { expect, test } from "@playwright/test";
import { mkdirSync } from "node:fs";

const pages = [
  { name: "home-normal", path: "/" },
  { name: "register-normal", path: "/register" },
  { name: "register-success-feedback", path: "/register?state=success" },
  { name: "register-verify-normal", path: "/register/verify" },
  { name: "register-verify-success-feedback", path: "/register/verify?state=success" },
  { name: "login-error", path: "/login?state=error" },
  { name: "login-focus-visible", path: "/login?state=focus" },
  { name: "forgot-password-normal", path: "/forgot-password" },
  { name: "forgot-password-success-feedback", path: "/forgot-password?state=success" },

  { name: "programs-normal", path: "/programs" },
  { name: "programs-long-text", path: "/programs?state=long" },
  { name: "programs-permission-denied", path: "/programs?state=denied" },
  { name: "program-detail-normal", path: "/programs/program_editorial_signal" },
  { name: "program-detail-restricted", path: "/programs/program_empty" },
  { name: "program-detail-long-text", path: "/programs/program_editorial_signal?state=long" },
  { name: "collections-normal", path: "/collections" },
  { name: "collections-empty", path: "/collections?state=empty" },
  { name: "collection-editor-normal", path: "/collections/collection_research_weekly" },
  { name: "collection-editor-empty", path: "/collections/collection_research_weekly?state=empty" },
  { name: "collection-editor-preview-empty", path: "/collections/collection_research_weekly?state=preview_empty" },
  { name: "collection-editor-permission-denied", path: "/collections/collection_research_weekly?state=denied" },
  { name: "rss-subscribe-normal", path: "/collections/collection_research_weekly/subscribe" },
  { name: "rss-subscribe-success-feedback", path: "/collections/collection_research_weekly/subscribe?toast=success" },
  { name: "rss-subscribe-revoked", path: "/collections/collection_research_weekly/subscribe?state=revoked" },

  { name: "admin-overview-normal", path: "/admin" },
  { name: "admin-overview-loading", path: "/admin?state=loading" },
  { name: "admin-overview-success-feedback", path: "/admin?state=success" },
  { name: "admin-programs-normal", path: "/admin/programs" },
  { name: "admin-programs-long-text", path: "/admin/programs?state=long" },
  { name: "admin-programs-empty", path: "/admin/programs?state=empty" },
  { name: "admin-program-detail-normal", path: "/admin/programs/program_editorial_signal" },
  { name: "admin-program-detail-auth", path: "/admin/programs/program_editorial_signal?state=auth" },
  { name: "admin-program-detail-draft", path: "/admin/programs/program_editorial_signal?state=draft" },
  { name: "admin-program-detail-permission-denied", path: "/admin/programs/program_editorial_signal?state=denied" },
  { name: "admin-connectors-normal", path: "/admin/connectors" },
  { name: "admin-connector-detail-normal", path: "/admin/connectors/connector_partner_archive" },
  { name: "admin-connector-detail-success-feedback", path: "/admin/connectors/connector_partner_archive?toast=success" },
  { name: "admin-connector-new-normal", path: "/admin/connectors/new" },
  { name: "admin-import-jobs-normal", path: "/admin/import-jobs" },
  { name: "admin-import-job-running", path: "/admin/import-jobs/job_mock_1004" },
  { name: "admin-import-job-waiting-auth", path: "/admin/import-jobs/job_mock_1003?state=waiting_auth" },
  { name: "admin-import-job-failed", path: "/admin/import-jobs/job_mock_1003" },
  { name: "admin-reviews-normal", path: "/admin/reviews" },
  { name: "admin-reviews-empty", path: "/admin/reviews?state=empty" },
  { name: "admin-users-normal", path: "/admin/users" },
  { name: "admin-users-long-email", path: "/admin/users?state=long" },
  { name: "admin-users-permission-denied", path: "/admin/users?state=denied" },
  { name: "components-states", path: "/components" }
];

const overlayPages = [
  { name: "program-detail-add-to-collection-drawer-overlay", path: "/programs/program_editorial_signal?drawer=add" },
  { name: "rss-subscribe-reset-dialog-overlay", path: "/collections/collection_research_weekly/subscribe?dialog=reset" },
  { name: "admin-review-drawer-overlay", path: "/admin/reviews?drawer=review_mock_002" },
  { name: "admin-review-reject-dialog-overlay", path: "/admin/reviews?dialog=reject&drawer=review_mock_002" },
  { name: "admin-job-cancel-dialog-overlay", path: "/admin/import-jobs/job_mock_1004?dialog=cancel" },
  { name: "admin-users-permission-drawer-overlay", path: "/admin/users?drawer=normal_user_suspended" },
  { name: "admin-import-job-waiting-auth-overlay", path: "/admin/import-jobs/job_mock_1003?state=waiting_auth" },
  { name: "admin-connector-success-toast-overlay", path: "/admin/connectors/connector_partner_archive?toast=success" }
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
