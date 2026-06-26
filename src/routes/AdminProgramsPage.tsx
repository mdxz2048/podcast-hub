import { Plus } from "lucide-react";
import { Link } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { SearchBar, Select } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { EmptyState, ErrorState, LoadingState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import { programs } from "../mock/data";
import { programStatusLabel } from "../utils/labels";
import { useViewState } from "./state";

export function AdminProgramsPage() {
  const state = useViewState();

  if (state === "loading") return <LoadingState title="正在加载节目列表" />;
  if (state === "empty") return <EmptyState title="暂无节目" action="创建节目" />;
  if (state === "error") return <ErrorState title="节目管理暂不可用" />;
  if (state === "denied") return <PermissionDeniedState />;

  return (
    <>
      <PageHeader eyebrow="节目" title="管理节目身份、内容来源、审核状态和发布准备度">
        <Button icon={<Plus className="h-4 w-4" />}>创建节目</Button>
      </PageHeader>
      {state === "success" ? <SuccessFeedback message="节目草稿已保存到模拟状态。" /> : null}
      <div className="mb-5 grid gap-3 rounded-lg border border-border bg-surface p-4 md:grid-cols-[1fr_220px_220px]">
        <SearchBar placeholder="搜索节目" />
        <Select label="状态" options={["全部状态", "已启用", "权利暂缓", "草稿"]} />
        <Select label="发布范围" options={["全部范围", "公开", "指定用户", "私有"]} />
      </div>
      <div className="overflow-hidden rounded-lg border border-border bg-surface shadow-subtle">
        <div className="hidden grid-cols-[1.4fr_140px_120px_120px_1fr] gap-4 border-b border-border px-4 py-3 text-xs font-semibold uppercase text-muted lg:grid">
          <span>节目</span>
          <span>状态</span>
          <span>来源</span>
          <span>单集</span>
          <span>下一步操作</span>
        </div>
        {programs.map((program) => (
          <article key={program.id} className="grid gap-3 border-b border-border px-4 py-4 last:border-b-0 lg:grid-cols-[1.4fr_140px_120px_120px_1fr] lg:items-center">
            <div className="min-w-0">
              <Link to={`/admin/programs/${program.id}`} className="break-words font-semibold leading-tight hover:text-action">{program.title}</Link>
              <p className="mt-1 line-clamp-2 text-sm text-secondary">{program.description}</p>
            </div>
            <Badge tone={program.status === "rights_hold" ? "danger" : program.status === "draft" ? "warning" : "success"}>{programStatusLabel[program.status]}</Badge>
            <span className="text-sm text-secondary">{program.sourceCount} 个来源</span>
            <span className="text-sm text-secondary">{program.episodeCount} 个单集</span>
            <Link to={`/admin/programs/${program.id}`} className="text-sm font-medium text-primary hover:text-action">{program.status === "rights_hold" ? "处理权利暂缓" : "检查发布状态"}</Link>
          </article>
        ))}
      </div>
    </>
  );
}
