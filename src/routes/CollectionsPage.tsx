import { Plus, Radio } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { EmptyState, ErrorState, LoadingState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import { useMockState } from "../mock/MockState";
import { useViewState } from "./state";

export function CollectionsPage() {
  const state = useViewState();
  const { collections, createCollection, showToast } = useMockState();
  const [title, setTitle] = useState("");
  const [created, setCreated] = useState(false);

  if (state === "loading") return <LoadingState title="正在加载我的合集" />;
  if (state === "error") return <ErrorState title="合集列表暂不可用" />;
  if (state === "denied") return <PermissionDeniedState />;

  const visibleCollections = state === "empty" ? [] : collections;

  function handleCreate() {
    const collection = createCollection(title || "新的合集");
    setTitle("");
    setCreated(true);
    showToast({ tone: "success", title: "合集已创建", message: `「${collection.title}」已加入本地模拟列表。` });
  }

  return (
    <div className="grid gap-7">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div>
          <p className="mb-2 text-xs font-semibold uppercase text-muted">我的合集</p>
          <h1 className="text-3xl font-semibold leading-tight md:text-4xl">管理个人 RSS 合集</h1>
          <p className="mt-3 max-w-2xl text-secondary">把多个已授权节目组合成一个模拟订阅源，方便在外部播客客户端中统一订阅。</p>
        </div>
        <Link to="/programs">
          <Button variant="secondary">浏览节目</Button>
        </Link>
      </header>
      {created ? <SuccessFeedback message="新建合集只保存在当前前端内存中。" /> : null}
      <section className="grid gap-3 rounded-lg border border-border bg-surface p-4 md:grid-cols-[1fr_auto]">
        <Input label="新建合集" placeholder="例如：每日通勤" value={title} onChange={(event) => setTitle(event.target.value)} />
        <Button className="self-end" icon={<Plus className="h-4 w-4" />} onClick={handleCreate}>创建合集</Button>
      </section>
      {visibleCollections.length === 0 ? (
        <EmptyState title="还没有任何合集" action="先浏览节目" />
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {visibleCollections.map((collection) => (
            <article key={collection.id} className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <h2 className="text-lg font-semibold">{collection.title}</h2>
                  <p className="mt-2 line-clamp-2 text-sm text-secondary">{collection.description}</p>
                </div>
                <Badge tone={collection.rssTokenState === "active" ? "success" : "danger"}>
                  {collection.rssTokenState === "active" ? "RSS 已启用" : "RSS 已撤销"}
                </Badge>
              </div>
              <div className="mt-4 flex flex-wrap gap-3 text-sm text-muted">
                <span>{collection.programIds.length} 个节目</span>
                <span>{collection.accessScope === "private" ? "私有" : "指定用户"}</span>
                <span>更新于 {collection.lastUpdatedAt}</span>
              </div>
              <div className="mt-5 flex flex-wrap gap-2">
                <Link to={`/collections/${collection.id}`}>
                  <Button>编辑合集</Button>
                </Link>
                <Link to={`/collections/${collection.id}/subscribe`}>
                  <Button variant="secondary" icon={<Radio className="h-4 w-4" />}>订阅地址</Button>
                </Link>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}
