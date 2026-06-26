import { ArrowDown, ArrowUp, Plus, Save, Trash2 } from "lucide-react";
import { useMemo, useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { EpisodeRow } from "../components/EpisodeRow";
import { Input } from "../components/Form";
import { EmptyState, ErrorState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import { programs } from "../mock/data";
import { useMockState } from "../mock/MockState";
import { previewEpisodesForCollection, programsForCollection, programTitle } from "../mock/selectors";
import type { Collection } from "../types/domain";

export function CollectionEditorPage() {
  const { collectionId } = useParams();
  const [params] = useSearchParams();
  const state = params.get("state");
  const { collections, updateCollection, showToast } = useMockState();
  const collection = collections.find((item) => item.id === collectionId);
  const forcedEmpty = state === "empty";
  const forcedPreviewEmpty = state === "preview_empty";

  const baseDraft: Collection | undefined = useMemo(() => {
    if (!collection) return undefined;
    return forcedEmpty ? { ...collection, programIds: [] } : collection;
  }, [collection, forcedEmpty]);

  const [title, setTitle] = useState(baseDraft?.title ?? "");
  const [description, setDescription] = useState(baseDraft?.description ?? "");
  const [sortOrder, setSortOrder] = useState(baseDraft?.rules.sortOrder ?? "newest");
  const [perProgramLimit, setPerProgramLimit] = useState(baseDraft?.rules.perProgramLimit ?? 3);
  const [totalLimit, setTotalLimit] = useState(forcedPreviewEmpty ? 0 : baseDraft?.rules.totalLimit ?? 8);
  const [programIds, setProgramIds] = useState(baseDraft?.programIds ?? []);
  const [saved, setSaved] = useState(false);

  if (state === "denied") return <PermissionDeniedState />;
  if (!baseDraft) return <ErrorState title="合集不存在" />;
  const base = baseDraft;

  const draft: Collection = {
    ...base,
    programIds,
    title,
    description,
    rules: { sortOrder, perProgramLimit, totalLimit }
  };
  const selectedPrograms = programsForCollection(draft);
  const availablePrograms = programs.filter((program) => !draft.programIds.includes(program.id) && program.accessState !== "blocked");
  const previewEpisodes = forcedPreviewEmpty ? [] : previewEpisodesForCollection(draft);
  const unsaved = title !== base.title
    || description !== base.description
    || sortOrder !== base.rules.sortOrder
    || perProgramLimit !== base.rules.perProgramLimit
    || totalLimit !== base.rules.totalLimit
    || programIds.join("|") !== base.programIds.join("|");

  function save() {
    updateCollection(base.id, {
      title,
      description,
      programIds,
      rules: { sortOrder, perProgramLimit, totalLimit }
    });
    setSaved(true);
    showToast({ tone: "success", title: "合集已保存", message: "模拟设置已更新，未调用真实 API。" });
  }

  function addProgram(programId: string) {
    setProgramIds((current) => current.includes(programId) ? current : [...current, programId]);
  }

  function removeProgram(programId: string) {
    setProgramIds((current) => current.filter((id) => id !== programId));
  }

  function moveLocalProgram(programId: string, direction: "up" | "down") {
    setProgramIds((current) => {
      const index = current.indexOf(programId);
      const target = direction === "up" ? index - 1 : index + 1;
      if (index < 0 || target < 0 || target >= current.length) return current;
      const next = [...current];
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  }

  return (
    <div className="grid gap-6">
      <header className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div>
          <p className="mb-2 text-xs font-semibold uppercase text-muted">合集编辑器</p>
          <h1 className="text-3xl font-semibold leading-tight md:text-4xl">{draft.title || "未命名合集"}</h1>
          <p className="mt-3 max-w-2xl text-secondary">修改合集内容和模拟规则，右侧预览会即时显示将进入 RSS 的单集。</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {unsaved ? <Badge tone="warning">有未保存修改</Badge> : <Badge tone="success">已保存</Badge>}
          <Link to={`/collections/${base.id}/subscribe`}>
            <Button variant="secondary">查看订阅地址</Button>
          </Link>
          <Button icon={<Save className="h-4 w-4" />} onClick={save}>保存</Button>
        </div>
      </header>
      {saved ? <SuccessFeedback message="合集设置已保存到本地内存，刷新后恢复初始模拟数据。" /> : null}
      <div className="grid gap-6 xl:grid-cols-[0.9fr_1.1fr_0.9fr]">
        <section className="grid gap-4">
          <h2 className="text-lg font-semibold">已选择的节目</h2>
          {selectedPrograms.length === 0 ? <EmptyState title="合集里还没有节目" /> : (
            <div className="grid gap-3">
              {selectedPrograms.map((program, index) => (
                <article key={program.id} className="rounded-lg border border-border bg-surface p-4">
                  <h3 className="font-semibold leading-snug">{program.title}</h3>
                  <p className="mt-2 line-clamp-2 text-sm text-secondary">{program.description}</p>
                  <div className="mt-4 flex flex-wrap gap-2">
                    <Button variant="secondary" className="px-3" aria-label={`上移 ${program.title}`} onClick={() => moveLocalProgram(program.id, "up")} disabled={index === 0}>
                      <ArrowUp className="h-4 w-4" />
                    </Button>
                    <Button variant="secondary" className="px-3" aria-label={`下移 ${program.title}`} onClick={() => moveLocalProgram(program.id, "down")} disabled={index === selectedPrograms.length - 1}>
                      <ArrowDown className="h-4 w-4" />
                    </Button>
                    <Button variant="danger" icon={<Trash2 className="h-4 w-4" />} onClick={() => removeProgram(program.id)}>移除</Button>
                  </div>
                </article>
              ))}
            </div>
          )}
          <div className="rounded-lg border border-border bg-surface p-4">
            <h3 className="font-semibold">添加节目</h3>
            <div className="mt-3 grid gap-2">
              {availablePrograms.length === 0 ? <p className="text-sm text-secondary">没有更多可加入的节目。</p> : availablePrograms.map((program) => (
                <Button key={program.id} variant="secondary" icon={<Plus className="h-4 w-4" />} onClick={() => addProgram(program.id)}>
                  {program.title}
                </Button>
              ))}
            </div>
          </div>
        </section>
        <section className="grid gap-4">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h2 className="text-lg font-semibold">RSS 内容实时预览</h2>
            <div className="flex flex-wrap gap-2">
              <Badge tone="info">{previewEpisodes.length} 个单集</Badge>
              <Badge>{selectedPrograms.length} 个节目</Badge>
              <Badge tone="success">预计每日检查</Badge>
            </div>
          </div>
          {previewEpisodes.length === 0 ? <EmptyState title="当前规则不会产生 RSS 单集" /> : (
            <div className="grid gap-3">
              {previewEpisodes.map((episode) => (
                <EpisodeRow key={episode.id} episode={episode} programTitle={programTitle(episode.programId)} />
              ))}
            </div>
          )}
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
            <label className="grid gap-2 text-sm font-medium text-primary">
              排序方式
              <select className="min-h-11 rounded-md border border-border bg-surface px-3 text-sm" value={sortOrder} onChange={(event) => setSortOrder(event.target.value as Collection["rules"]["sortOrder"])}>
                <option value="newest">按发布时间从新到旧</option>
                <option value="oldest">按发布时间从旧到新</option>
              </select>
            </label>
            <Input label="每个节目保留单集数量" type="number" min={0} value={perProgramLimit} onChange={(event) => setPerProgramLimit(Number(event.target.value))} />
            <Input label="合集总单集数量" type="number" min={0} value={totalLimit} onChange={(event) => setTotalLimit(Number(event.target.value))} />
          </div>
        </section>
      </div>
    </div>
  );
}
