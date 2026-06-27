import { AlertTriangle, Copy, Plus, RefreshCcw, ShieldAlert, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Dialog } from "../components/Dialog";
import { Input } from "../components/Form";
import { EmptyState, SuccessFeedback } from "../components/StateBlocks";
import { useMockState } from "../mock/MockState";

export function RssFeedsPage() {
  const { rssFeeds, revealedRssToken, createRssFeed, rotateRssFeed, revokeRssFeed, deleteRssFeed, clearRevealedRssToken, showToast } = useMockState();
  const [draftName, setDraftName] = useState("");
  const [revokeFeedId, setRevokeFeedId] = useState<string | null>(null);
  const [deleteFeedId, setDeleteFeedId] = useState<string | null>(null);

  const activeCount = useMemo(() => rssFeeds.filter((feed) => feed.status === "active").length, [rssFeeds]);

  async function copyLink(url: string) {
    try {
      if (!navigator.clipboard) throw new Error("clipboard unavailable");
      await navigator.clipboard.writeText(url);
      showToast({ tone: "success", title: "链接已复制", message: "这是一次性显示的模拟私有 RSS 链接，请勿公开分享。" });
    } catch {
      showToast({ tone: "danger", title: "复制失败", message: "浏览器未允许剪贴板访问，请手动复制。" });
    }
  }

  function handleCreate() {
    const revealed = createRssFeed(draftName);
    setDraftName("");
    showToast({ tone: "success", title: "RSS Feed 已创建", message: `已生成 ${revealed.action === "created" ? "新的" : "更新后的"}模拟私有链接。` });
  }

  function handleRotate(feedId: string) {
    const revealed = rotateRssFeed(feedId);
    if (revealed) {
      showToast({ tone: "success", title: "RSS Feed 已轮换", message: "旧链接视为失效，新的模拟链接仅在当前界面显示一次。" });
    }
  }

  return (
    <div className="grid gap-6">
      <header className="grid gap-3 md:grid-cols-[1fr_auto] md:items-end">
        <div>
          <p className="mb-2 text-xs font-semibold uppercase text-muted">Private RSS</p>
          <h1 className="text-3xl font-semibold leading-tight md:text-4xl">管理你的私有 RSS Feed</h1>
          <p className="mt-3 max-w-3xl text-secondary">每个链接都等同于访问凭据。真实版本中，撤销或轮换后旧链接与媒体访问会立即失效。</p>
        </div>
        <div className="rounded-lg border border-border bg-surface px-4 py-3 text-sm text-secondary shadow-subtle">
          <p>Active feeds: <span className="font-semibold text-primary">{activeCount}</span></p>
          <p>Visible one-time token: <span className="font-semibold text-primary">{revealedRssToken ? "1" : "0"}</span></p>
        </div>
      </header>

      {revealedRssToken ? (
        <section className="grid gap-4 rounded-lg border border-success/20 bg-success/10 p-5">
          <SuccessFeedback message={revealedRssToken.action === "created" ? "创建成功后仅在当前界面显示一次明文链接。刷新页面后无法再次读取。" : "轮换成功后旧链接应立刻失效。当前原型只模拟这条规则。"} />
          <label className="grid gap-2 text-sm font-medium text-primary">
            一次性私有 RSS 链接
            <textarea className="min-h-24 rounded-md border border-border bg-surface px-3 py-2 font-mono text-sm text-primary" readOnly value={revealedRssToken.url} />
          </label>
          <div className="flex flex-wrap gap-2">
            <Button icon={<Copy className="h-4 w-4" />} onClick={() => void copyLink(revealedRssToken.url)}>复制链接</Button>
            <Button variant="secondary" onClick={clearRevealedRssToken}>关闭明文展示</Button>
          </div>
        </section>
      ) : null}

      <section className="grid gap-4 rounded-lg border border-border bg-surface p-5 shadow-subtle md:grid-cols-[1fr_auto] md:items-end">
        <Input
          label="新建 Feed 名称"
          placeholder="例如：每日通勤订阅"
          hint="当前为前端模拟页，不连接真实 Go API。"
          value={draftName}
          onChange={(event) => setDraftName(event.target.value)}
        />
        <Button icon={<Plus className="h-4 w-4" />} onClick={handleCreate}>创建 Feed</Button>
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
            <AlertTriangle className="h-5 w-5 text-danger" /> 当前原型限制
          </div>
          <p className="mt-2 text-sm text-secondary">当前页面不会把 token 写入 localStorage、sessionStorage、URL 查询参数或浏览器日志。刷新后明文展示会被清空。</p>
        </div>
      </section>

      <section className="grid gap-4">
        {rssFeeds.length === 0 ? <EmptyState title="你还没有私有 RSS Feed" action="创建第一个 Feed" /> : rssFeeds.map((feed) => (
          <article key={feed.id} className="grid gap-4 rounded-lg border border-border bg-surface p-5 shadow-subtle md:grid-cols-[1fr_auto] md:items-center">
            <div className="grid gap-3">
              <div className="flex flex-wrap items-center gap-2">
                <h2 className="text-lg font-semibold">{feed.name}</h2>
                <Badge tone={feed.status === "active" ? "success" : feed.status === "expired" ? "warning" : "danger"}>{feed.status === "active" ? "Active" : feed.status === "expired" ? "Expired" : "Revoked"}</Badge>
                <Badge>{feed.tokenPrefix}</Badge>
              </div>
              <div className="grid gap-1 text-sm text-secondary md:grid-cols-3">
                <p>创建时间：{feed.createdAt}</p>
                <p>最近使用：{feed.lastUsedAt ?? "尚未使用"}</p>
                <p>最近轮换：{feed.rotatedAt ?? "未轮换"}</p>
              </div>
            </div>
            <div className="flex flex-wrap gap-2 md:justify-end">
              <Button variant="secondary" icon={<RefreshCcw className="h-4 w-4" />} onClick={() => handleRotate(feed.id)}>轮换</Button>
              <Button variant="danger" onClick={() => setRevokeFeedId(feed.id)}>撤销</Button>
              <Button variant="ghost" icon={<Trash2 className="h-4 w-4" />} onClick={() => setDeleteFeedId(feed.id)}>删除</Button>
            </div>
          </article>
        ))}
      </section>

      <Dialog
        open={revokeFeedId !== null}
        title="撤销 RSS Feed"
        description="撤销后旧链接应立即失效。当前原型不会访问真实后端，但会模拟状态切换。"
        confirmLabel="确认撤销"
        onCancel={() => setRevokeFeedId(null)}
        onConfirm={() => {
          if (revokeFeedId) revokeRssFeed(revokeFeedId);
          setRevokeFeedId(null);
          showToast({ tone: "success", title: "Feed 已撤销", message: "旧链接在真实系统中应立即失效。" });
        }}
      />
      <Dialog
        open={deleteFeedId !== null}
        title="删除 RSS Feed"
        description="删除只影响当前前端模拟状态。真实版本应由后端审计并删除记录。"
        confirmLabel="确认删除"
        onCancel={() => setDeleteFeedId(null)}
        onConfirm={() => {
          if (deleteFeedId) deleteRssFeed(deleteFeedId);
          setDeleteFeedId(null);
          showToast({ tone: "success", title: "Feed 已删除", message: "该模拟 Feed 已从本地状态移除。" });
        }}
      />
    </div>
  );
}