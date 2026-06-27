import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Badge } from "../components/Badge";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState } from "../components/StateBlocks";
import { listStagingEpisodes, listStagingPrograms } from "../api/staging";
import type { StagingEpisode, StagingProgram } from "../api/staging";

export function AdminStagingPage() {
  const [programs, setPrograms] = useState<StagingProgram[]>([]);
  const [episodes, setEpisodes] = useState<StagingEpisode[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([listStagingPrograms(), listStagingEpisodes()])
      .then(([programResult, episodeResult]) => {
        setPrograms(programResult.programs);
        setEpisodes(episodeResult.episodes);
      })
      .catch(() => setError("无法加载待审核区。"))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingState title="正在加载待审核区" />;
  if (error) return <ErrorState title={error} />;

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="Staging" title="待审核区" />
      {programs.length === 0 && episodes.length === 0 ? <EmptyState title="待审核区暂无内容" /> : null}
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <h2 className="text-lg font-semibold">Program Candidates</h2>
        {programs.length === 0 ? <EmptyState title="暂无待审核 Program" /> : (
          <div className="mt-4 grid gap-3">
            {programs.map((program) => (
              <Link key={program.id} to={`/admin/staging/programs/${program.id}`} className="rounded-md border border-border p-4 hover:border-strong">
                <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                  <div className="min-w-0">
                    <h3 className="break-words font-semibold">{program.title}</h3>
                    <p className="mt-1 line-clamp-2 text-sm text-secondary">{program.description}</p>
                    <p className="mt-2 text-xs text-muted">Source {program.created_from_source_id} · Job {program.created_from_job_id}</p>
                  </div>
                  <Badge tone={program.status === "review_pending" ? "warning" : "info"}>{program.status}</Badge>
                </div>
              </Link>
            ))}
          </div>
        )}
      </section>
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <h2 className="text-lg font-semibold">Episode Candidates</h2>
        {episodes.length === 0 ? <EmptyState title="暂无待审核 Episode" /> : (
          <div className="mt-4 grid gap-3">
            {episodes.map((episode) => (
              <Link key={episode.id} to={`/admin/staging/episodes/${episode.id}`} className="rounded-md border border-border p-4 hover:border-strong">
                <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                  <div className="min-w-0">
                    <h3 className="break-words font-semibold">{episode.title}</h3>
                    <p className="mt-1 line-clamp-2 text-sm text-secondary">{episode.description}</p>
                    <p className="mt-2 text-xs text-muted">Program {episode.program_id} · Job {episode.source_job_id}</p>
                  </div>
                  <Badge tone={episode.status === "review_pending" ? "warning" : "info"}>{episode.status}</Badge>
                </div>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
