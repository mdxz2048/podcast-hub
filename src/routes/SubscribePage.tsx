import { Copy, RefreshCcw, ShieldAlert } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Dialog } from "../components/Dialog";
import { ErrorState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import { useMockState } from "../mock/MockState";
import { previewEpisodesForCollection } from "../mock/selectors";

export function SubscribePage() {
  const { collectionId } = useParams();
  const [params] = useSearchParams();
  const state = params.get("state");
  const dialogFromUrl = params.get("dialog") === "reset";
  const toastFromUrl = params.get("toast") === "success";
  const { collections, resetRssToken, showToast } = useMockState();
  const [dialogOpen, setDialogOpen] = useState(dialogFromUrl);
  const collection = collections.find((item) => item.id === collectionId);

  useEffect(() => {
    if (toastFromUrl) {
      showToast({ tone: "success", title: "已复制", message: "模拟 RSS 地址已复制到剪贴板。" });
    }
  }, [showToast, toastFromUrl]);

  if (state === "denied") return <PermissionDeniedState />;
  if (!collection) return <ErrorState title="合集不存在" />;
  const activeCollection = collection;
  if (state === "revoked" || activeCollection.rssTokenState === "revoked") return <ErrorState title="订阅地址已失效" />;

  const rssUrl = `https://rss.example.invalid/users/mock-user/collections/${activeCollection.id}/feed.xml?token=mock-token-redacted`;
  const previewEpisodes = previewEpisodesForCollection(activeCollection);

  async function copyUrl() {
    try {
      if (!navigator.clipboard) throw new Error("clipboard unavailable");
      await navigator.clipboard.writeText(rssUrl);
      showToast({ tone: "success", title: "已复制", message: "模拟 RSS 地址已复制。请勿在真实场景公开私有链接。" });
    } catch {
      showToast({ tone: "danger", title: "复制失败", message: "浏览器拒绝剪贴板访问，请手动选择地址。" });
    }
  }

  function confirmReset() {
    resetRssToken(activeCollection.id);
    setDialogOpen(false);
    showToast({ tone: "success", title: "地址已重置", message: "这是模拟反馈，未生成真实 RSS Token。" });
  }

  return (
    <div className="mx-auto grid max-w-5xl gap-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div>
          <p className="mb-2 text-xs font-semibold uppercase text-muted">RSS 订阅</p>
          <h1 className="text-3xl font-semibold leading-tight md:text-4xl">{activeCollection.title}</h1>
          <p className="mt-3 max-w-2xl text-secondary">复制这个模拟地址到外部播客客户端即可体验订阅路径。M0.2A 不生成真实 RSS XML。</p>
        </div>
        <Link to={`/collections/${activeCollection.id}`}>
          <Button variant="secondary">返回编辑器</Button>
        </Link>
      </header>
      {toastFromUrl ? <SuccessFeedback message="截图状态：复制成功 Toast 已显示。" /> : null}
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <div className="flex flex-wrap items-center gap-2">
          <Badge tone="success">RSS 已启用</Badge>
          <Badge>{previewEpisodes.length} 个预览单集</Badge>
        </div>
        <label className="mt-5 grid gap-2 text-sm font-medium text-primary">
          模拟 RSS 地址
          <textarea className="min-h-24 rounded-md border border-border bg-subtle px-3 py-2 font-mono text-sm text-primary" readOnly value={rssUrl} />
        </label>
        <div className="mt-4 flex flex-wrap gap-2">
          <Button icon={<Copy className="h-4 w-4" />} onClick={copyUrl}>复制地址</Button>
          <Button variant="danger" icon={<RefreshCcw className="h-4 w-4" />} onClick={() => setDialogOpen(true)}>重置 RSS 地址</Button>
        </div>
      </section>
      <section className="grid gap-4 md:grid-cols-2">
        <div className="rounded-lg border border-border bg-surface p-5">
          <h2 className="font-semibold">外部客户端订阅说明</h2>
          <p className="mt-2 text-sm text-secondary">在支持自定义 RSS 的播客客户端中选择添加订阅源，粘贴上方地址。不同客户端的入口名称可能略有不同。</p>
        </div>
        <div className="rounded-lg border border-warning/30 bg-warning/5 p-5">
          <div className="flex items-center gap-2 font-semibold">
            <ShieldAlert className="h-5 w-5 text-warning" /> 私有链接安全提示
          </div>
          <p className="mt-2 text-sm text-secondary">私有 RSS 地址等同于访问凭据。真实版本中重置 Token 后旧地址会立即失效。</p>
        </div>
      </section>
      <Dialog
        open={dialogOpen}
        title="重置 RSS 地址"
        description={`确认后「${activeCollection.title}」的旧模拟订阅地址会被视为失效。真实版本中该操作需要审计。`}
        confirmLabel="确认重置"
        onCancel={() => setDialogOpen(false)}
        onConfirm={confirmReset}
      />
    </div>
  );
}
