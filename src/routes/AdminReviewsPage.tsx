import { useEffect, useState } from "react";
import { CheckCircle2, XCircle } from "lucide-react";
import { Link } from "react-router-dom";
import type { ApiError } from "../api/client";
import { approveReview, listReviews, rejectReview } from "../api/adminContent";
import type { ReviewItem } from "../api/adminContent";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, SuccessFeedback } from "../components/StateBlocks";

export function AdminReviewsPage() {
  const [reviews, setReviews] = useState<ReviewItem[]>([]);
  const [reasonById, setReasonById] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function reload() {
    const result = await listReviews();
    setReviews(result.reviews);
  }

  useEffect(() => {
    reload().catch(() => setError("审核队列暂不可用。")).finally(() => setLoading(false));
  }, []);

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

  if (loading) return <LoadingState title="正在加载审核队列" />;
  if (error && reviews.length === 0) return <ErrorState title={error} />;

  return (
    <div className="grid gap-6">
      <PageHeader eyebrow="审核队列" title="处理 Program 与 Episode 的审核项" />
      {success ? <SuccessFeedback message={success} /> : null}
      {error ? <ErrorState title={error} /> : null}
      {reviews.length === 0 ? <EmptyState title="当前没有审核项" /> : (
        <div className="grid gap-4">
          {reviews.map((review) => (
            <article key={review.id} className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
              <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div className="min-w-0">
                  <div className="flex flex-wrap gap-2">
                    <Badge tone={review.status === "pending" ? "warning" : review.status === "approved" ? "success" : "danger"}>{review.status}</Badge>
                    <Badge>{review.target_type}</Badge>
                    <Badge>{review.review_kind}</Badge>
                  </div>
                  <Link to={`/admin/review/${review.id}`} className="mt-3 block break-words font-semibold hover:text-action">{review.id}</Link>
                  <p className="mt-2 text-sm text-secondary">Target {review.target_id} · Created {formatDate(review.created_at)}</p>
                </div>
                <div className="grid gap-2 sm:min-w-80">
                  <textarea className="min-h-20 rounded-md border border-border bg-surface px-3 py-2 text-sm text-primary" placeholder="拒绝原因" value={reasonById[review.id] ?? ""} onChange={(event) => setReasonById((current) => ({ ...current, [review.id]: event.target.value }))} />
                  <div className="flex flex-wrap gap-2">
                    <Button icon={<CheckCircle2 className="h-4 w-4" />} disabled={review.status !== "pending" || busy} onClick={() => action(() => approveReview(review.id), "审核已通过。")}>通过</Button>
                    <Button variant="danger" icon={<XCircle className="h-4 w-4" />} disabled={review.status !== "pending" || busy} onClick={() => action(() => rejectReview(review.id, reasonById[review.id] ?? ""), "审核已拒绝。")}>拒绝</Button>
                  </div>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}

function formatDate(value?: string) {
  if (!value) return "not set";
  return new Date(value).toLocaleString();
}
