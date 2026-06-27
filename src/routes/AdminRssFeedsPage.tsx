import { Radio, ShieldOff } from "lucide-react";
import { useEffect, useState } from "react";
import { adminRevokeRssFeed, listAdminRssFeeds } from "../api/rssFeeds";
import type { AdminRssFeed } from "../api/rssFeeds";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Dialog } from "../components/Dialog";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminRssFeedsPage() {
  const [feeds, setFeeds] = useState<AdminRssFeed[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [revokeFeedId, setRevokeFeedId] = useState<string | null>(null);

  useEffect(() => {
    listAdminRssFeeds()
      .then((items) => {
        setFeeds(items);
        setError(null);
      })
      .catch(() => setError("RSS Feed 列表暂不可用。"))
      .finally(() => setLoading(false));
  }, []);

  async function revoke(feedId: string) {
    setBusy(true);
    setError(null);
    try {
      const feed = await adminRevokeRssFeed(feedId);
      setFeeds((current) => current.map((item) => item.id === feedId ? { ...item, ...feed } : item));
      setSuccess("RSS Feed 已撤销。");
    } catch {
      setError("撤销 RSS Feed 失败。");
    } finally {
      setBusy(false);
      setRevokeFeedId(null);
    }
  }

  if (loading) return <LoadingState title="正在加载 RSS Feed" />;

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="Private RSS" title="RSS Feed 安全元数据" />
      {success ? <SuccessFeedback message={success} /> : null}
      {error ? <ErrorState title={error} /> : null}
      {feeds.length === 0 ? <EmptyState title="暂无 RSS Feed" /> : (
        <div className="grid gap-4">
          {feeds.map((feed) => (
            <article key={feed.id} className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
              <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap gap-2">
                    <Badge tone={feed.status === "active" ? "success" : feed.status === "expired" ? "warning" : "danger"}>{feed.status}</Badge>
                    <Badge>{feed.token_prefix}</Badge>
                  </div>
                  <h2 className="mt-3 text-lg font-semibold">{feed.name}</h2>
                  <p className="mt-1 break-all text-sm text-secondary">Owner: {feed.user_email_hint}</p>
                  <div className="mt-3 grid gap-1 text-sm text-secondary md:grid-cols-3">
                    <p>Created: {formatDate(feed.created_at)}</p>
                    <p>Last used: {formatDate(feed.last_used_at)}</p>
                    <p>Revoked: {formatDate(feed.revoked_at)}</p>
                  </div>
                </div>
                <Button variant="danger" icon={<ShieldOff className="h-4 w-4" />} disabled={busy || feed.status !== "active"} onClick={() => setRevokeFeedId(feed.id)}>撤销</Button>
              </div>
            </article>
          ))}
        </div>
      )}
      <section className="rounded-lg border border-border bg-surface p-5 text-sm text-secondary">
        <div className="flex items-center gap-2 font-semibold text-primary">
          <Radio className="h-5 w-5" /> 管理员不可读取 Token 明文
        </div>
        <p className="mt-2">此页只显示 Feed 名称、prefix、状态、owner 安全展示和时间元数据。</p>
      </section>
      <Dialog
        open={revokeFeedId !== null}
        title="撤销 RSS Feed"
        description="管理员撤销后，该 Feed 的 XML 和 enclosure URL 都会失效。"
        confirmLabel="确认撤销"
        onCancel={() => setRevokeFeedId(null)}
        onConfirm={() => revokeFeedId ? void revoke(revokeFeedId) : undefined}
      />
    </div>
  );
}

function formatDate(value?: string) {
  if (!value) return "not set";
  return new Date(value).toLocaleString();
}
