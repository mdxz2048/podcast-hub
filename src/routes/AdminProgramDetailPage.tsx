import { useEffect, useState } from "react";
import { ArrowLeft, Archive, CheckCircle2, Save } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import type { ApiError } from "../api/client";
import { archiveProgram, getAdminProgram, patchAdminProgram, publishProgram, submitProgramReview } from "../api/adminContent";
import type { AdminEpisode, AdminProgram } from "../api/adminContent";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Input } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminProgramDetailPage() {
  const { programId = "" } = useParams();
  const [program, setProgram] = useState<AdminProgram | null>(null);
  const [episodes, setEpisodes] = useState<AdminEpisode[]>([]);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function reload() {
    const result = await getAdminProgram(programId);
    setProgram(result.program);
    setEpisodes(result.episodes);
    setTitle(result.program.title);
    setDescription(result.program.description);
  }

  useEffect(() => {
    reload().catch(() => setError("节目不存在或暂不可用。")).finally(() => setLoading(false));
  }, [programId]);

  async function run(action: () => Promise<{ program?: AdminProgram; review?: unknown }>, message: string) {
    setBusy(true);
    setError(null);
    try {
      const result = await action();
      if (result.program) setProgram(result.program);
      setSuccess(message);
    } catch (err) {
      setError((err as ApiError).message);
    } finally {
      setBusy(false);
    }
  }

  if (loading) return <LoadingState title="正在加载节目" />;
  if (error && !program) return <ErrorState title={error} />;
  if (!program) return <EmptyState title="节目不存在" />;

  const canPublish = program.status === "approved";
  const publishReason = canPublish ? "前置条件会由后端再次校验。" : "只有 approved Program 可以发布。";

  return (
    <div className="grid gap-6">
      <Link to="/admin/programs" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回节目列表
      </Link>
      <PageHeader eyebrow="Program" title={program.title}>
        <Button variant="secondary" icon={<CheckCircle2 className="h-4 w-4" />} onClick={() => run(() => submitProgramReview(program.id), "已提交审核。")} disabled={busy}>提交审核</Button>
        <Button icon={<CheckCircle2 className="h-4 w-4" />} onClick={() => run(() => publishProgram(program.id), "Program 已发布。")} disabled={!canPublish || busy}>发布</Button>
        <Button variant="danger" icon={<Archive className="h-4 w-4" />} onClick={() => run(() => archiveProgram(program.id), "Program 已归档。")} disabled={busy}>归档</Button>
      </PageHeader>
      {success ? <SuccessFeedback message={success} /> : null}
      {error ? <ErrorState title={error} /> : null}
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <div className="flex flex-wrap gap-2">
          <Badge tone={program.status === "published" ? "success" : program.status === "archived" || program.status === "rejected" ? "danger" : "warning"}>{program.status}</Badge>
          <Badge>{program.language}</Badge>
        </div>
        <p className="mt-4 text-sm text-secondary">{publishReason}</p>
        <dl className="mt-4 grid gap-3 text-sm text-secondary md:grid-cols-2">
          <Info label="Source" value={program.created_from_source_id} />
          <Info label="Import Job" value={program.created_from_job_id} />
          <Info label="Updated" value={formatDate(program.updated_at)} />
          <Info label="Author" value={program.author} />
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
          <Button className="w-fit" icon={<Save className="h-4 w-4" />} disabled={busy} onClick={() => run(() => patchAdminProgram(program.id, { title, description }), "元数据已保存并写入审计。")}>保存元数据</Button>
        </div>
      </section>
      <section className="rounded-lg border border-border bg-surface p-5">
        <h2 className="text-lg font-semibold">Episodes</h2>
        {episodes.length === 0 ? <EmptyState title="暂无 Episode" /> : (
          <div className="mt-4 grid gap-3">
            {episodes.map((episode) => (
              <Link key={episode.id} to={`/admin/episodes/${episode.id}`} className="rounded-md border border-border p-4 hover:border-strong">
                <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                  <div>
                    <h3 className="font-semibold">{episode.title}</h3>
                    <p className="mt-1 text-sm text-secondary">{episode.description}</p>
                  </div>
                  <Badge tone={episode.status === "published" ? "success" : "warning"}>{episode.status}</Badge>
                </div>
              </Link>
            ))}
          </div>
        )}
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
