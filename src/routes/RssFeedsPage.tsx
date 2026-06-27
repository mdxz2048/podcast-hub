import { AlertTriangle, Copy, Plus, RefreshCcw, ShieldAlert, Trash2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { createRssFeed, deleteRssFeed, listRssFeeds, revokeRssFeed, rotateRssFeed } from "../api/rssFeeds";
import type { OneTimeFeedToken, RssFeed } from "../api/rssFeeds";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Dialog } from "../components/Dialog";
import { Input } from "../components/Form";
import { EmptyState, ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function RssFeedsPage() {
  const [feeds, setFeeds] = useState<RssFeed[]>([]);
  const [revealed, setRevealed] = useState<OneTimeFeedToken | null>(null);
  const [draftName, setDraftName] = useState("");
  const [revokeFeedId, setRevokeFeedId] = useState<string | null>(null);
  const [deleteFeedId, setDeleteFeedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    listRssFeeds()
      .then((items) => {
        setFeeds(items);
        setError(null);
      })
      .catch(() => setError("RSS Feed 暂不可用。"))
      .finally(() => setLoading(false));
  }, []);

  const activeCount = useMemo(() => feeds.filter((feed) => feed.status === "active").length, [feeds]);

  async function copyLink(url: string) {
    try {
      if (!navigator.clipboard) throw new Error("clipboard unavailable");
      await navigator.clipboard.writeText(url);
      setSuccess("链接已复制。请像密码一样保管这个一次性展示的私有 RSS URL。");
    } catch {
      setError("浏览器未允许剪贴板访问，请手动复制。");
    }
  }

  async function handleCreate() {
    setBusy(true);
    setError(null);
    try {
      const result = await createRssFeed(draftName);
      setFeeds((current) => [result.feed, ...current.filter((feed) => feed.id !== result.feed.id)]);
      setRevealed(result);
      setDraftName("");
      setSuccess("RSS Feed 已创建，明文链接仅显示一次。");
    } catch {
      setError("创建 RSS Feed 失败。");
    } finally {
      setBusy(false);
    }
  }

  async function handleRotate(feedId: string) {
    setBusy(true);
    setError(null);
    try {
      const result = await rotateRssFeed(feedId);
      setFeeds((current) => current.map((feed) => feed.id === feedId ? result.feed : feed));
      setRevealed(result);
      setSuccess("RSS Feed 已轮换，旧链接会立即失效。");
    } catch {
      setError("轮换 RSS Feed 失败。");
    } finally {
      setBusy(false);
    }
  }

  async function handleRevoke(feedId: string) {
    setBusy(true);
    setError(null);
    try {
      const feed = await revokeRssFeed(feedId);
      setFeeds((current) => current.map((item) => item.id === feedId ? feed : item));
      if (revealed?.feed.id === feedId) setRevealed(null);
      setSuccess("Feed 已撤销，旧链接不再有效。");
    } catch {
      setError("撤销 RSS Feed 失败。");
    } finally {
      setBusy(false);
      setRevokeFeedId(null);
    }
  }

  async function handleDelete(feedId: string) {
    setBusy(true);
    setError(null);
    try {
      await deleteRssFeed(feedId);
      setFeeds((current) => current.filter((feed) => feed.id !== feedId));
      if (revealed?.feed.id === feedId) setRevealed(null);
      setSuccess("Feed 已删除。");
    } catch {
      setError("删除 RSS Feed 失败。");
    } finally {
      setBusy(false);
      setDeleteFeedId(null);
    }
  }

  if (loading) return <LoadingState title="正在加载 RSS Feed" />;

  return (
    <div className="grid gap-6">
      <header className="grid gap-3 md:grid-cols-[1fr_auto] md:items-end">
        <div>
          <p className="mb-2 text-xs font-semibold uppercase text-muted">Private RSS</p>
          <h1 className="text-3xl font-semibold leading-tight md:text-4xl">管理你的私有 RSS Feed</h1>
          <p className="mt-3 max-w-3xl text-secondary">私有 RSS URL 等同于访问凭据。不要分享；怀疑泄露后立即轮换。</p>
        </div>
        <div className="rounded-lg border border-border bg-surface px-4 py-3 text-sm text-secondary shadow-subtle">
          <p>Active feeds: <span className="font-semibold text-primary">{activeCount}</span></p>
          <p>Visible one-time token: <span className="font-semibold text-primary">{revealed ? "1" : "0"}</span></p>
        </div>
      </header>

      {success ? <SuccessFeedback message={success} /> : null}
      {error ? <ErrorState title={error} /> : null}

      {revealed ? (
        <section className="grid gap-4 rounded-lg border border-success/20 bg-success/10 p-5">
          <SuccessFeedback message="创建或轮换成功后仅在当前界面显示一次明文链接。刷新页面后无法再次读取。" />
          <label className="grid gap-2 text-sm font-medium text-primary">
            一次性私有 RSS 链接
            <textarea className="min-h-24 rounded-md border border-border bg-surface px-3 py-2 font-mono text-sm text-primary" readOnly value={revealed.feed_url} />
          </label>
          <div className="flex flex-wrap gap-2">
            <Button icon={<Copy className="h-4 w-4" />} onClick={() => void copyLink(revealed.feed_url)}>复制链接</Button>
            <Button variant="secondary" onClick={() => setRevealed(null)}>关闭明文展示</Button>
          </div>
        </section>
      ) : null}

      <section className="grid gap-4 rounded-lg border border-border bg-surface p-5 shadow-subtle md:grid-cols-[1fr_auto] md:items-end">
        <Input
          label="新建 Feed 名称"
          placeholder="例如：每日通勤订阅"
          hint="Token 明文只会在创建或轮换成功后显示一次。"
          value={draftName}
          onChange={(event) => setDraftName(event.target.value)}
        />
        <Button icon={<Plus className="h-4 w-4" />} disabled={busy} onClick={handleCreate}>创建 Feed</Button>
      </section>

      <section className="grid gap-4 md:grid-cols-2">
        <div className="rounded-lg border border-warning/30 bg-warning/5 p-5">
          <div className="flex items-center gap-2 font-semibold">
            <ShieldAlert className="h-5 w-5 text-warning" /> 私有链接安全提示
          </div>
          <p className="mt-2 text-sm text-secondary">不要把私有 RSS 链接发到公开群组、博客或工单系统。若怀疑泄露，请立刻轮换。</p>
        </div>
        <div className="rounded-lg border border-border bg-surface p-5">
          <div className="flex items-center gap-2 font-semibold">
            <AlertTriangle className="h-5 w-5 text-danger" /> Token 展示规则
          </div>
          <p className="mt-2 text-sm text-secondary">列表永远只显示 token prefix；明文不会进入 localStorage、sessionStorage、路由状态或前端日志。</p>
        </div>
      </section>

      <section className="grid gap-4">
        {feeds.length === 0 ? <EmptyState title="你还没有私有 RSS Feed" action="创建第一个 Feed" /> : feeds.map((feed) => (
          <article key={feed.id} className="grid gap-4 rounded-lg border border-border bg-surface p-5 shadow-subtle md:grid-cols-[1fr_auto] md:items-center">
            <div className="grid gap-3">
              <div className="flex flex-wrap items-center gap-2">
                <h2 className="text-lg font-semibold">{feed.name}</h2>
                <Badge tone={feed.status === "active" ? "success" : feed.status === "expired" ? "warning" : "danger"}>{feed.status === "active" ? "Active" : feed.status === "expired" ? "Expired" : "Revoked"}</Badge>
                <Badge>{feed.token_prefix}</Badge>
              </div>
              <div className="grid gap-1 text-sm text-secondary md:grid-cols-3">
                <p>创建时间：{formatDate(feed.created_at)}</p>
                <p>最近使用：{formatDate(feed.last_used_at)}</p>
                <p>过期时间：{formatDate(feed.expires_at)}</p>
                <p>撤销时间：{formatDate(feed.revoked_at)}</p>
              </div>
            </div>
            <div className="flex flex-wrap gap-2 md:justify-end">
              <Button variant="secondary" icon={<RefreshCcw className="h-4 w-4" />} disabled={busy || feed.status !== "active"} onClick={() => void handleRotate(feed.id)}>轮换</Button>
              <Button variant="danger" disabled={busy || feed.status !== "active"} onClick={() => setRevokeFeedId(feed.id)}>撤销</Button>
              <Button variant="ghost" icon={<Trash2 className="h-4 w-4" />} disabled={busy} onClick={() => setDeleteFeedId(feed.id)}>删除</Button>
            </div>
          </article>
        ))}
      </section>

      <Dialog
        open={revokeFeedId !== null}
        title="撤销 RSS Feed"
        description="撤销后旧链接和 enclosure 访问会立即失效。"
        confirmLabel="确认撤销"
        onCancel={() => setRevokeFeedId(null)}
        onConfirm={() => revokeFeedId ? void handleRevoke(revokeFeedId) : undefined}
      />
      <Dialog
        open={deleteFeedId !== null}
        title="删除 RSS Feed"
        description="删除后该 Feed 记录会从当前账号列表移除。"
        confirmLabel="确认删除"
        onCancel={() => setDeleteFeedId(null)}
        onConfirm={() => deleteFeedId ? void handleDelete(deleteFeedId) : undefined}
      />
    </div>
  );
}

function formatDate(value?: string) {
  if (!value) return "未记录";
  return new Date(value).toLocaleString("zh-CN");
}
