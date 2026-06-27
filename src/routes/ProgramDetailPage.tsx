import { ArrowLeft, FolderPlus, Heart, Lock, Plus } from "lucide-react";
import type { CSSProperties } from "react";
import { useEffect, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { addProgram, createCollection, listCollections } from "../api/collections";
import type { UserCollectionView } from "../api/collections";
import { getProgram, listProgramEpisodes } from "../api/programs";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Drawer } from "../components/Drawer";
import { EpisodeRow } from "../components/EpisodeRow";
import { Input } from "../components/Form";
import { EmptyState, ErrorState, LoadingState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import type { Episode, Program } from "../types/domain";
import { accessStateLabel, programStatusLabel, rightsStateLabel } from "../utils/labels";

export function ProgramDetailPage() {
  const { programId } = useParams();
  const [params] = useSearchParams();
  const state = params.get("state");
  const drawerFromUrl = params.get("drawer") === "add";
  const [program, setProgram] = useState<Program | null>(null);
  const [programEpisodes, setProgramEpisodes] = useState<Episode[]>([]);
  const [collections, setCollections] = useState<UserCollectionView[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(drawerFromUrl);
  const [selectedCollectionId, setSelectedCollectionId] = useState("");
  const [newCollectionTitle, setNewCollectionTitle] = useState("");
  const [favorite, setFavorite] = useState(false);

  useEffect(() => {
    if (!programId) return;
    Promise.all([getProgram(programId), listProgramEpisodes(programId), listCollections()])
      .then(([nextProgram, nextEpisodes, nextCollections]) => {
        setProgram(nextProgram);
        setProgramEpisodes(nextEpisodes);
        setCollections(nextCollections);
        setSelectedCollectionId(nextCollections[0]?.id ?? "");
      })
      .catch(() => setError("节目不存在或当前账号无权访问。"))
      .finally(() => setLoading(false));
  }, [programId]);

  if (loading) return <LoadingState title="正在加载节目" />;
  if (error) return <ErrorState title={error} />;
  if (!program) return <ErrorState title="节目不存在" />;
  const activeProgram = program;
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
          <p className="mt-2 text-secondary">当前账号没有访问该节目的权限，单集不会出现在你的目录或合集里。</p>
        </section>
      </div>
    );
  }

  async function handleAdd() {
    try {
      const target = newCollectionTitle.trim()
        ? await createCollection({ title: newCollectionTitle, description: "" })
        : collections.find((collection) => collection.id === selectedCollectionId);
      if (target) {
        const updated = await addProgram(target.id, activeProgram.id);
        const nextCollections = collections.some((item) => item.id === updated.id)
          ? collections.map((item) => item.id === updated.id ? updated : item)
          : [updated, ...collections];
        setCollections(nextCollections);
        setSelectedCollectionId(updated.id);
        setSuccess("节目已加入合集。");
      }
      setDrawerOpen(false);
      setNewCollectionTitle("");
    } catch {
      setError("合集更新失败或当前账号无权加入该节目。");
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
      {favorite ? <SuccessFeedback message="收藏状态已更新。" /> : null}
      {success ? <SuccessFeedback message={success} /> : null}
      <section className="rounded-lg border border-success/20 bg-success/10 p-4 text-sm text-primary">
        <p className="font-semibold">授权后可播放</p>
        <p className="mt-1 text-secondary">当前账号有权访问该节目。页面只展示已发布且带已发布媒体的单集，不显示存储地址或下载入口。</p>
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
          <p className="text-sm text-secondary">选择一个已有合集，或创建新合集。</p>
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
