import { AlertTriangle, CheckCircle2, Eye, PauseCircle, RotateCcw, XCircle } from "lucide-react";
import { Link } from "react-router-dom";
import { useSearchParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { Dialog } from "../components/Dialog";
import { Drawer } from "../components/Drawer";
import { EmptyState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import { episodes, programs, reviewItems, sources } from "../mock/data";
import { useMockState } from "../mock/MockState";
import { reviewStatusLabel } from "../utils/labels";

export function AdminReviewsPage() {
  const [params] = useSearchParams();
  const { showToast } = useMockState();
  const state = params.get("state");
  if (state === "denied") return <PermissionDeniedState />;

  const visible = state === "empty" ? [] : reviewItems;
  const activeReview = reviewItems.find((item) => item.id === (params.get("drawer") ?? reviewItems[0]?.id));
  const dialogOpen = params.get("dialog") === "reject";

  function action(title: string) {
    showToast({ tone: "success", title, message: "审核操作已写入模拟状态，没有调用真实后端。" });
  }

  return (
    <div className="grid gap-6">
      <header>
        <p className="mb-2 text-xs font-semibold uppercase text-muted">审核队列</p>
        <h1 className="text-3xl font-semibold leading-tight md:text-4xl">处理待审核单集和发布风险</h1>
        <p className="mt-3 max-w-3xl text-secondary">审核前必须确认元数据、来源、授权状态和重复风险。未审核内容不得进入正式 RSS。</p>
      </header>
      {params.get("toast") === "success" ? <SuccessFeedback message="审核通过的模拟反馈已显示。" /> : null}
      {visible.length === 0 ? <EmptyState title="当前没有待审核单集" /> : (
        <div className="grid gap-4">
          {visible.map((item) => {
            const episode = episodes.find((entry) => entry.id === item.episodeId);
            const program = programs.find((entry) => entry.id === item.programId);
            const source = sources.find((entry) => entry.id === item.sourceId);
            return (
              <article key={item.id} className="rounded-lg border border-border bg-surface p-5 shadow-subtle">
                <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                  <div className="min-w-0">
                    <div className="flex flex-wrap gap-2">
                      <Badge tone={item.status === "pending_review" ? "info" : item.status === "duplicate_risk" ? "warning" : "danger"}>{reviewStatusLabel[item.status]}</Badge>
                      <Badge tone={item.rightsState === "clear" ? "success" : "warning"}>{item.rightsState === "clear" ? "授权已确认" : "授权待确认"}</Badge>
                      {item.duplicateRisk !== "none" ? <Badge tone="warning">重复风险</Badge> : null}
                    </div>
                    <h2 className="mt-3 text-lg font-semibold">{episode?.title ?? "未知单集"}</h2>
                    <p className="mt-2 text-sm text-secondary">{program?.title} · {source?.name} · 元数据完整度 {item.metadataCompleteness}% · {item.publishDate}</p>
                    <p className="mt-2 text-sm text-secondary">{item.suggestion}</p>
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Link to={`/admin/reviews?drawer=${item.id}`}>
                      <Button variant="secondary" icon={<Eye className="h-4 w-4" />}>查看详情</Button>
                    </Link>
                    <Button icon={<CheckCircle2 className="h-4 w-4" />} onClick={() => action("审核已通过")}>通过</Button>
                    <Link to={`/admin/reviews?dialog=reject&drawer=${item.id}`}>
                      <Button variant="danger" icon={<XCircle className="h-4 w-4" />}>拒绝</Button>
                    </Link>
                    <Button variant="secondary" icon={<RotateCcw className="h-4 w-4" />} onClick={() => action("已退回补充资料")}>退回补充</Button>
                    <Button variant="secondary" icon={<PauseCircle className="h-4 w-4" />} onClick={() => action("已暂停发布")}>暂停发布</Button>
                  </div>
                </div>
              </article>
            );
          })}
        </div>
      )}
      <Drawer open={Boolean(params.get("drawer")) && Boolean(activeReview)} title="审核详情" onClose={() => window.history.back()}>
        {activeReview ? <ReviewDetail reviewId={activeReview.id} /> : null}
      </Drawer>
      <Dialog
        open={dialogOpen}
        title="拒绝审核项"
        description="确认拒绝该单集？拒绝后不会进入发布候选列表，真实版本会写入审核审计。"
        confirmLabel="确认拒绝"
        onCancel={() => window.history.back()}
        onConfirm={() => action("审核项已拒绝")}
      >
        <div className="flex items-start gap-2 rounded-md border border-danger/25 bg-danger/5 p-3 text-sm text-secondary">
          <AlertTriangle className="mt-0.5 h-4 w-4 text-danger" /> 请确认不是因为权限缺失导致的临时阻塞。
        </div>
      </Dialog>
    </div>
  );
}

function ReviewDetail({ reviewId }: { reviewId: string }) {
  const item = reviewItems.find((review) => review.id === reviewId);
  if (!item) return null;
  const episode = episodes.find((entry) => entry.id === item.episodeId);
  const program = programs.find((entry) => entry.id === item.programId);
  const source = sources.find((entry) => entry.id === item.sourceId);

  return (
    <div className="grid gap-4">
      <div>
        <p className="text-xs text-muted">关联节目</p>
        <p className="font-semibold">{program?.title}</p>
      </div>
      <div>
        <p className="text-xs text-muted">单集标题</p>
        <p className="font-semibold">{episode?.title}</p>
        <p className="mt-2 text-sm text-secondary">{episode?.summary}</p>
      </div>
      <div className="grid gap-3 rounded-lg border border-border bg-subtle p-4">
        <p>来源：{source?.name}</p>
        <p>导入任务：{item.jobId}</p>
        <p>文件状态：{item.fileState === "staged" ? "已进入隔离区" : item.fileState === "missing" ? "文件缺失" : "仅元数据"}</p>
        <p>重复风险：{item.duplicateRisk === "none" ? "无" : item.duplicateRisk === "possible" ? "可能重复" : "高风险重复"}</p>
      </div>
    </div>
  );
}
