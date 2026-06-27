import { ArrowLeft, FolderPlus, Heart, Lock, Plus } from "lucide-react";
import type { CSSProperties } from "react";
import { useMemo, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Drawer } from "../components/Drawer";
import { EpisodeRow } from "../components/EpisodeRow";
import { Input } from "../components/Form";
import { EmptyState, ErrorState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import { programs } from "../mock/data";
import { useMockState } from "../mock/MockState";
import { episodesForProgram, findProgram } from "../mock/selectors";
import { accessStateLabel, programStatusLabel, rightsStateLabel } from "../utils/labels";

export function ProgramDetailPage() {
  const { programId } = useParams();
  const [params] = useSearchParams();
  const program = findProgram(programId);
  const state = params.get("state");
  const drawerFromUrl = params.get("drawer") === "add";
  const { collections, addProgramToCollection, createCollection, showToast } = useMockState();
  const [drawerOpen, setDrawerOpen] = useState(drawerFromUrl);
  const [selectedCollectionId, setSelectedCollectionId] = useState(collections[0]?.id ?? "");
  const [newCollectionTitle, setNewCollectionTitle] = useState("");
  const [favorite, setFavorite] = useState(false);

  const displayProgram = useMemo(() => {
    if (state === "long") return programs.find((item) => item.id === "program_field_archive") ?? program;
    return program;
  }, [program, state]);

  if (!displayProgram) return <ErrorState title="节目不存在" />;
  const activeProgram = displayProgram;
  if (state === "denied") return <PermissionDeniedState />;
  if (state === "unavailable") return <ErrorState title="节目暂不可用" />;
  if (activeProgram.accessState === "blocked") {
    return (
      <div className="grid gap-5">
        <Link to="/programs" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
          <ArrowLeft className="h-4 w-4" /> 返回节目库
        </Link>
        <section className="rounded-lg border border-warning/30 bg-warning/5 p-8">
          <Lock className="mb-4 h-7 w-7 text-warning" />
          <h1 className="text-2xl font-semibold">访问受限</h1>
          <p className="mt-2 text-secondary">当前账号没有访问「{activeProgram.title}」的权限，单集不会出现在你的 RSS 合集中。</p>
        </section>
      </div>
    );
  }

  const programEpisodes = episodesForProgram(activeProgram.id);

  function handleAdd() {
    const targetId = newCollectionTitle.trim()
      ? createCollection(newCollectionTitle, activeProgram.id).id
      : selectedCollectionId;
    if (targetId) {
      addProgramToCollection(targetId, activeProgram.id);
      showToast({ tone: "success", title: "已加入合集", message: `「${activeProgram.title}」已加入选中的模拟合集。` });
      setDrawerOpen(false);
      setNewCollectionTitle("");
    }
  }

  return (
    <div className="grid gap-6">
      <Link to="/programs" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回节目库
      </Link>
      <section className="grid gap-6 rounded-lg border border-border bg-surface p-5 shadow-subtle lg:grid-cols-[220px_1fr]">
        <div
          className="cover-art relative aspect-square rounded-lg"
          style={{ "--cover-a": activeProgram.coverTone[0], "--cover-b": activeProgram.coverTone[1] } as CSSProperties}
          aria-hidden="true"
        />
        <div className="min-w-0">
          <div className="flex flex-wrap gap-2">
            <Badge tone={activeProgram.accessState === "public" ? "info" : "success"}>{accessStateLabel[activeProgram.accessState]}</Badge>
            <Badge tone={activeProgram.rightsState === "clear" ? "success" : "warning"}>{rightsStateLabel[activeProgram.rightsState]}</Badge>
            <Badge>{programStatusLabel[activeProgram.status]}</Badge>
          </div>
          <h1 className="mt-4 text-3xl font-semibold leading-tight md:text-4xl">{activeProgram.title}</h1>
          <p className="mt-3 max-w-3xl text-secondary">{activeProgram.description}</p>
          <dl className="mt-5 grid gap-3 text-sm text-secondary sm:grid-cols-2 lg:grid-cols-4">
            <Meta label="作者" value={activeProgram.author} />
            <Meta label="分类" value={activeProgram.category} />
            <Meta label="语言" value={activeProgram.language} />
            <Meta label="更新频率" value={activeProgram.updateFrequency} />
          </dl>
          <div className="mt-6 flex flex-wrap gap-2">
            <Button variant="secondary" icon={<Heart className="h-4 w-4" />} onClick={() => setFavorite((value) => !value)}>
              {favorite ? "已收藏" : "收藏节目"}
            </Button>
            <Button icon={<FolderPlus className="h-4 w-4" />} onClick={() => setDrawerOpen(true)}>加入合集</Button>
          </div>
        </div>
      </section>
      {favorite ? <SuccessFeedback message="收藏状态已更新到本地内存，刷新后会恢复初始状态。" /> : null}
      <section className="rounded-lg border border-success/20 bg-success/10 p-4 text-sm text-primary">
        <p className="font-semibold">授权后可播放</p>
        <p className="mt-1 text-secondary">真实版本会在已登录、节目授权有效且单集已发布时提供私有播放能力。当前原型不会显示真实媒体链接，也不会提供下载按钮。</p>
      </section>
      <section>
        <h2 className="mb-4 text-xl font-semibold">最近单集</h2>
        {programEpisodes.length === 0 ? <EmptyState title="这个节目还没有可预览单集" /> : (
          <div className="grid gap-3">
            {programEpisodes.map((episode) => <EpisodeRow key={episode.id} episode={episode} />)}
          </div>
        )}
      </section>
      <Drawer open={drawerOpen} title="加入合集" onClose={() => setDrawerOpen(false)}>
        <div className="grid gap-5">
          <p className="text-sm text-secondary">选择一个已有合集，或创建新合集。操作只更新当前前端内存状态。</p>
          <div className="grid gap-3">
            {collections.map((collection) => (
              <label key={collection.id} className="flex items-start gap-3 rounded-lg border border-border p-4">
                <input
                  className="mt-1"
                  type="radio"
                  name="collection"
                  checked={selectedCollectionId === collection.id}
                  onChange={() => setSelectedCollectionId(collection.id)}
                />
                <span>
                  <span className="block font-semibold">{collection.title}</span>
                  <span className="text-sm text-secondary">{collection.programIds.length} 个节目 · {collection.lastUpdatedAt}</span>
                </span>
              </label>
            ))}
          </div>
          <Input label="新建合集名称" placeholder="例如：通勤订阅" value={newCollectionTitle} onChange={(event) => setNewCollectionTitle(event.target.value)} />
          <Button icon={<Plus className="h-4 w-4" />} onClick={handleAdd}>确认加入</Button>
        </div>
      </Drawer>
    </div>
  );
}

function Meta({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs text-muted">{label}</dt>
      <dd className="mt-1 font-medium text-primary">{value}</dd>
    </div>
  );
}
