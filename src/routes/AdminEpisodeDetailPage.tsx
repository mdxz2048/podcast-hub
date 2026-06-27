import { useEffect, useState } from "react";
import { Archive, ArrowLeft, CheckCircle2, Save } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import type { ApiError } from "../api/client";
import { archiveEpisode, getAdminEpisode, patchAdminEpisode, publishEpisode, submitEpisodeReview } from "../api/adminContent";
import type { AdminEpisode } from "../api/adminContent";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminEpisodeDetailPage() {
  const { episodeId = "" } = useParams();
  const [episode, setEpisode] = useState<AdminEpisode | null>(null);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  useEffect(() => {
    getAdminEpisode(episodeId)
      .then((result) => {
        setEpisode(result.episode);
        setTitle(result.episode.title);
        setDescription(result.episode.description);
      })
      .catch(() => setError("Episode 不存在或暂不可用。"))
      .finally(() => setLoading(false));
  }, [episodeId]);

  async function run(action: () => Promise<{ episode?: AdminEpisode; review?: unknown }>, message: string) {
    setBusy(true);
    setError(null);
    try {
      const result = await action();
      if (result.episode) setEpisode(result.episode);
      setSuccess(message);
    } catch (err) {
      setError((err as ApiError).message);
    } finally {
      setBusy(false);
    }
  }

  if (loading) return <LoadingState title="正在加载 Episode" />;
  if (error && !episode) return <ErrorState title={error} />;
  if (!episode) return <EmptyState title="Episode 不存在" />;

  const canPublish = episode.status === "approved";

  return (
    <div className="grid gap-6">
      <Link to={`/admin/programs/${episode.program_id}`} className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回 Program
      </Link>
      <PageHeader eyebrow="Episode" title={episode.title}>
        <Button variant="secondary" icon={<CheckCircle2 className="h-4 w-4" />} onClick={() => run(() => submitEpisodeReview(episode.id), "已提交审核。")} disabled={busy}>提交审核</Button>
        <Button icon={<CheckCircle2 className="h-4 w-4" />} onClick={() => run(() => publishEpisode(episode.id), "Episode 已发布。")} disabled={!canPublish || busy}>发布</Button>
        <Button variant="danger" icon={<Archive className="h-4 w-4" />} onClick={() => run(() => archiveEpisode(episode.id), "Episode 已归档。")} disabled={busy}>归档</Button>
      </PageHeader>
      {success ? <SuccessFeedback message={success} /> : null}
      {error ? <ErrorState title={error} /> : null}
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <div className="flex flex-wrap gap-2">
          <Badge tone={episode.status === "published" ? "success" : episode.status === "archived" || episode.status === "rejected" ? "danger" : "warning"}>{episode.status}</Badge>
          <Badge>{episode.duration_seconds} seconds</Badge>
        </div>
        <p className="mt-4 text-sm text-secondary">{canPublish ? "后端会确认 Program 已发布、Review 完成、MediaAsset 已 approved。" : "只有 approved Episode 可以发布。"}</p>
        <dl className="mt-4 grid gap-3 text-sm text-secondary md:grid-cols-2">
          <Info label="Program" value={episode.program_id} />
          <Info label="Import Job" value={episode.source_job_id} />
          <Info label="External ID" value={episode.external_episode_id} />
          <Info label="Published At" value={formatDate(episode.published_at)} />
        </dl>
      </section>
      <section className="rounded-lg border border-border bg-surface p-5">
        <h2 className="text-lg font-semibold">安全元数据</h2>
        <div className="mt-4 grid gap-4">
          <Input label="标题" value={title} onChange={(event) => setTitle(event.target.value)} />
          <label className="grid gap-2 text-sm font-medium text-primary">
            描述
            <textarea className="min-h-28 rounded-md border border-border bg-surface px-3 py-2 text-sm text-primary" value={description} onChange={(event) => setDescription(event.target.value)} />
          </label>
          <Button className="w-fit" icon={<Save className="h-4 w-4" />} disabled={busy} onClick={() => run(() => patchAdminEpisode(episode.id, { title, description }), "元数据已保存并写入审计。")}>保存元数据</Button>
        </div>
      </section>
      <section className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
        <p className="font-medium text-primary">发布边界</p>
        <p className="mt-1">该页不提供订阅源管理、用户订阅、外部媒体下载、Connector 执行或 staging 文件路径。</p>
      </section>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs uppercase text-muted">{label}</dt>
      <dd className="mt-1 break-words font-medium text-primary">{value}</dd>
    </div>
  );
}

function formatDate(value?: string) {
  if (!value) return "not set";
  return new Date(value).toLocaleString();
}
