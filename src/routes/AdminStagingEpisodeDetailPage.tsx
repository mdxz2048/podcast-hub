import { useEffect, useState } from "react";
import { ArrowLeft } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { getStagingEpisode } from "../api/staging";
import type { StagingEpisode } from "../api/staging";
import { Badge } from "../components/Badge";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState } from "../components/StateBlocks";

export function AdminStagingEpisodeDetailPage() {
  const { episodeId = "" } = useParams();
  const [episode, setEpisode] = useState<StagingEpisode | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getStagingEpisode(episodeId)
      .then((result) => setEpisode(result.episode))
      .catch(() => setError("无法加载 Episode Candidate。"))
      .finally(() => setLoading(false));
  }, [episodeId]);

  if (loading) return <LoadingState title="正在加载 Episode Candidate" />;
  if (error) return <ErrorState title={error} />;
  if (!episode) return <EmptyState title="Episode Candidate 不存在" />;

  return (
    <div className="grid gap-6">
      <Link to="/admin/staging" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回待审核区
      </Link>
      <PageHeader eyebrow="Episode Candidate" title={episode.title}>
        <Badge tone="warning">{episode.status}</Badge>
      </PageHeader>
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <dl className="grid gap-3 text-sm text-secondary md:grid-cols-2">
          <Info label="Program" value={episode.program_id} />
          <Info label="External Episode ID" value={episode.external_episode_id} />
          <Info label="Import Job" value={episode.source_job_id} />
          <Info label="Published At" value={formatDate(episode.published_at)} />
          <Info label="Duration" value={`${episode.duration_seconds} seconds`} />
          <Info label="Updated" value={formatDate(episode.updated_at)} />
        </dl>
        <p className="mt-5 whitespace-pre-wrap text-secondary">{episode.description}</p>
      </section>
      <section className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm text-secondary">
        <p className="font-medium text-primary">媒体仍为私有 staging metadata</p>
        <p className="mt-1">该页面不暴露文件路径、storage key、下载地址、订阅源链接或用户可见入口。</p>
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
