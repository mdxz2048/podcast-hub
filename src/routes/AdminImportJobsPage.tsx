import { SlidersHorizontal } from "lucide-react";
import { Link, useSearchParams } from "react-router-dom";
import { Badge } from "../components/Badge";
import { Button } from "../components/Button";
import { ImportJobCard } from "../components/ImportJobCard";
import { SearchBar, Select } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { jobs, programs, sources } from "../mock/data";
import { authModeLabel, ingestionTypeLabel, jobStatusLabel, triggerTypeLabel } from "../utils/labels";

export function AdminImportJobsPage() {
  const [params] = useSearchParams();
  const status = params.get("status") ?? "全部状态";
  const visible = jobs.filter((job) => status === "全部状态" || jobStatusLabel[job.status] === status);

  return (
    <div>
      <PageHeader eyebrow="导入任务" title="追踪每次导入、认证阻塞和审核入口">
        <Button variant="secondary" icon={<SlidersHorizontal className="h-4 w-4" />}>筛选</Button>
      </PageHeader>
      <div className="mb-5 grid gap-3 rounded-lg border border-border bg-surface p-4 md:grid-cols-[1fr_220px]">
        <SearchBar placeholder="搜索任务、节目或来源" />
        <Select label="状态" defaultValue={status} options={["全部状态", "运行中", "等待授权", "等待人工上传", "等待审核", "已完成", "失败", "已取消"]} />
      </div>
      <div className="grid gap-4 xl:grid-cols-[1fr_1fr]">
        {visible.map((job) => {
          const program = programs.find((item) => item.id === job.programId);
          const source = sources.find((item) => item.id === job.sourceId);
          return (
            <Link key={job.id} to={`/admin/import-jobs/${job.id}`}>
              <div className="grid gap-3">
                <ImportJobCard job={job} />
                <div className="rounded-lg border border-border bg-subtle p-3 text-sm text-secondary">
                  <div className="flex flex-wrap gap-2">
                    <Badge>{program?.title ?? "未知节目"}</Badge>
                    <Badge>{source?.name ?? "未知来源"}</Badge>
                  </div>
                  <p className="mt-2">{ingestionTypeLabel[job.ingestionType]} / {triggerTypeLabel[job.triggerType]} / {authModeLabel[job.authMode]}</p>
                </div>
              </div>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
