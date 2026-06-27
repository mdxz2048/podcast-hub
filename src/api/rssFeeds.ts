import { apiRequest, readCookie } from "./client";

export interface RssFeed {
  id: string;
  user_id: string;
  name: string;
  token_prefix: string;
  status: "active" | "revoked" | "expired";
  created_at: string;
  last_used_at?: string;
  rotated_at?: string;
  revoked_at?: string;
  expires_at?: string;
}

export interface AdminRssFeed extends RssFeed {
  user_email_hint: string;
}

export interface OneTimeFeedToken {
  feed: RssFeed;
  token: string;
  feed_url: string;
}

function csrfHeaders() {
  const csrf = readCookie("podcast_hub_csrf");
  return csrf ? { "X-CSRF-Token": csrf } : undefined;
}

export async function listRssFeeds() {
  const result = await apiRequest<{ feeds: RssFeed[] }>("/me/rss-feeds");
  return result.feeds;
}

export async function createRssFeed(name: string) {
  return apiRequest<OneTimeFeedToken>("/me/rss-feeds", {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({ name })
  });
}

export async function rotateRssFeed(feedId: string) {
  return apiRequest<OneTimeFeedToken>(`/me/rss-feeds/${encodeURIComponent(feedId)}/rotate`, {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({})
  });
}

export async function revokeRssFeed(feedId: string) {
  const result = await apiRequest<{ feed: RssFeed }>(`/me/rss-feeds/${encodeURIComponent(feedId)}/revoke`, {
    method: "POST",
    headers: csrfHeaders()
  });
  return result.feed;
}

export async function deleteRssFeed(feedId: string) {
  await apiRequest<void>(`/me/rss-feeds/${encodeURIComponent(feedId)}`, { method: "DELETE", headers: csrfHeaders() });
}

export async function listAdminRssFeeds() {
  const result = await apiRequest<{ feeds: AdminRssFeed[] }>("/admin/rss-feeds");
  return result.feeds;
}

export async function adminRevokeRssFeed(feedId: string) {
  const result = await apiRequest<{ feed: RssFeed }>(`/admin/rss-feeds/${encodeURIComponent(feedId)}/revoke`, {
    method: "POST",
    headers: csrfHeaders()
  });
  return result.feed;
}
