import { useEffect, useState } from "react";
import { ArrowLeft, CheckCircle2, XCircle } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import type { ApiError } from "../api/client";
import { approveReview, getReview, rejectReview } from "../api/adminContent";
import type { ReviewItem } from "../api/adminContent";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminReviewDetailPage() {
  const { reviewId = "" } = useParams();
  const [review, setReview] = useState<ReviewItem | null>(null);
  const [reason, setReason] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function reload() {
    const result = await getReview(reviewId);
    setReview(result.review);
  }

  useEffect(() => {
    reload().catch(() => setError("审核项不存在或暂不可用。")).finally(() => setLoading(false));
  }, [reviewId]);

  async function action(run: () => Promise<unknown>, message: string) {
    setBusy(true);
    setError(null);
    try {
      await run();
      await reload();
      setSuccess(message);
    } catch (err) {
      setError((err as ApiError).message);
    } finally {
      setBusy(false);
    }
  }

  if (loading) return <LoadingState title="正在加载审核项" />;
  if (error && !review) return <ErrorState title={error} />;
  if (!review) return <EmptyState title="审核项不存在" />;

  return (
    <div className="grid gap-6">
      <Link to="/admin/reviews" className="inline-flex items-center gap-2 text-sm text-secondary hover:text-action">
        <ArrowLeft className="h-4 w-4" /> 返回审核队列
      </Link>
      <PageHeader eyebrow="Review" title={review.id}>
        <Button icon={<CheckCircle2 className="h-4 w-4" />} disabled={review.status !== "pending" || busy} onClick={() => action(() => approveReview(review.id), "审核已通过。")}>通过</Button>
        <Button variant="danger" icon={<XCircle className="h-4 w-4" />} disabled={review.status !== "pending" || busy} onClick={() => action(() => rejectReview(review.id, reason), "审核已拒绝。")}>拒绝</Button>
      </PageHeader>
      {success ? <SuccessFeedback message={success} /> : null}
      {error ? <ErrorState title={error} /> : null}
      <section className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
        <div className="flex flex-wrap gap-2">
          <Badge tone={review.status === "pending" ? "warning" : review.status === "approved" ? "success" : "danger"}>{review.status}</Badge>
          <Badge>{review.target_type}</Badge>
          <Badge>{review.review_kind}</Badge>
        </div>
        <dl className="mt-4 grid gap-3 text-sm text-secondary md:grid-cols-2">
          <Info label="Target" value={review.target_id} />
          <Info label="Requested Job" value={review.requested_by_job_id ?? "not set"} />
          <Info label="Created" value={formatDate(review.created_at)} />
          <Info label="Reviewed" value={formatDate(review.reviewed_at)} />
        </dl>
      </section>
      <section className="rounded-lg border border-border bg-surface p-5">
        <label className="grid gap-2 text-sm font-medium text-primary">
          拒绝原因
          <textarea className="min-h-28 rounded-md border border-border bg-surface px-3 py-2 text-sm text-primary" value={reason} onChange={(event) => setReason(event.target.value)} />
        </label>
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
