import { ArrowDown, ArrowUp, Plus, Save, Trash2 } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { addProgram, listCollections, removeProgram, updateCollection } from "../api/collections";
import type { UserCollectionView } from "../api/collections";
import { listPrograms } from "../api/programs";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { EmptyState, ErrorState, LoadingState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import type { Program } from "../types/domain";

export function CollectionEditorPage() {
  const { collectionId } = useParams();
  const [params] = useSearchParams();
  const state = params.get("state");
  const [collections, setCollections] = useState<UserCollectionView[]>([]);
  const [programs, setPrograms] = useState<Program[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const forcedEmpty = state === "empty";
  const collection = collections.find((item) => item.id === collectionId);

  useEffect(() => {
    Promise.all([listCollections(), listPrograms()])
      .then(([nextCollections, nextPrograms]) => {
        setCollections(nextCollections);
        setPrograms(nextPrograms);
        setLoadError(false);
      })
      .catch(() => setLoadError(true))
      .finally(() => setLoading(false));
  }, []);

  const baseDraft = useMemo(() => {
    if (!collection) return undefined;
    return forcedEmpty ? { ...collection, programIds: [] } : collection;
  }, [collection, forcedEmpty]);

  const [title, setTitle] = useState(baseDraft?.title ?? "");
  const [description, setDescription] = useState(baseDraft?.description ?? "");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (!baseDraft) return;
    setTitle(baseDraft.title);
    setDescription(baseDraft.description);
  }, [baseDraft?.id]);

  if (loading) return <LoadingState title="正在加载合集" />;
  if (loadError) return <ErrorState title="合集暂不可用" />;
  if (state === "denied") return <PermissionDeniedState />;
  if (!baseDraft) return <ErrorState title="合集不存在" />;
  const base = baseDraft;

  const selectedPrograms = base.programs;
  const availablePrograms = programs.filter((program) => !base.programIds.includes(program.id) && program.accessState !== "blocked");
  const unsaved = title !== base.title
    || description !== base.description;

  async function save() {
    const updated = await updateCollection(base.id, { title, description });
    setCollections((current) => current.map((item) => item.id === updated.id ? updated : item));
    setSaved(true);
  }

  async function addLocalProgram(programId: string) {
    const updated = await addProgram(base.id, programId);
    setCollections((current) => current.map((item) => item.id === updated.id ? updated : item));
    setSaved(true);
  }

  async function removeLocalProgram(programId: string) {
    const updated = await removeProgram(base.id, programId);
    setCollections((current) => current.map((item) => item.id === updated.id ? updated : item));
    setSaved(true);
  }

  return (
    <div className="grid gap-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div>
          <p className="mb-2 text-xs font-semibold uppercase text-muted">合集编辑器</p>
          <h1 className="text-3xl font-semibold leading-tight md:text-4xl">{base.title || "未命名合集"}</h1>
          <p className="mt-3 max-w-2xl text-secondary">修改合集内容。无权节目会自动从读取结果中隐藏，不泄露历史标题。</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {unsaved ? <Badge tone="warning">有未保存修改</Badge> : <Badge tone="success">已保存</Badge>}
          <Button icon={<Save className="h-4 w-4" />} onClick={save}>保存</Button>
        </div>
      </header>
      {saved ? <SuccessFeedback message="合集设置已保存。" /> : null}
      <div className="grid gap-6 xl:grid-cols-[1fr_0.9fr]">
        <section className="grid gap-4">
          <h2 className="text-lg font-semibold">已选择的节目</h2>
          {selectedPrograms.length === 0 ? <EmptyState title="合集里还没有节目" /> : (
            <div className="grid gap-3">
              {selectedPrograms.map((program, index) => (
                <article key={program.id} className="rounded-lg border border-border bg-surface p-4">
                  <h3 className="font-semibold leading-snug">{program.title}</h3>
                  <p className="mt-2 line-clamp-2 text-sm text-secondary">{program.description}</p>
                  <div className="mt-4 flex flex-wrap gap-2">
                    <Button variant="secondary" className="px-3" aria-label={`上移 ${program.title}`} disabled={index === 0}>
                      <ArrowUp className="h-4 w-4" />
                    </Button>
                    <Button variant="secondary" className="px-3" aria-label={`下移 ${program.title}`} disabled={index === selectedPrograms.length - 1}>
                      <ArrowDown className="h-4 w-4" />
                    </Button>
                    <Button variant="danger" icon={<Trash2 className="h-4 w-4" />} onClick={() => removeLocalProgram(program.id)}>移除</Button>
                  </div>
                </article>
              ))}
            </div>
          )}
          <div className="rounded-lg border border-border bg-surface p-4">
            <h3 className="font-semibold">添加节目</h3>
            <div className="mt-3 grid gap-2">
              {availablePrograms.length === 0 ? <p className="text-sm text-secondary">没有更多可加入的节目。</p> : availablePrograms.map((program) => (
                <Button key={program.id} variant="secondary" icon={<Plus className="h-4 w-4" />} onClick={() => addLocalProgram(program.id)}>
                  {program.title}
                </Button>
              ))}
            </div>
          </div>
        </section>
        <section className="grid content-start gap-4">
          <h2 className="text-lg font-semibold">合集设置</h2>
          <div className="grid gap-4 rounded-lg border border-border bg-surface p-4">
            <Input label="合集名称" value={title} onChange={(event) => setTitle(event.target.value)} />
            <label className="grid gap-2 text-sm font-medium text-primary">
              合集简介
              <textarea
                className="min-h-28 rounded-md border border-border bg-surface px-3 py-2 text-primary placeholder:text-muted"
                value={description}
                onChange={(event) => setDescription(event.target.value)}
              />
            </label>
          </div>
        </section>
      </div>
    </div>
  );
}
