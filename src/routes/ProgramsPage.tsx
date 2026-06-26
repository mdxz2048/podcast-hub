import { SlidersHorizontal } from "lucide-react";
import { useMemo, useState } from "react";
import { Button } from "../components/Button";
import { SearchBar, Select } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { ProgramCard } from "../components/ProgramCard";
import { EmptyState, ErrorState, LoadingState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import { programs } from "../mock/data";
import { useViewState } from "./state";

export function ProgramsPage() {
  const state = useViewState();
  const [query, setQuery] = useState("");
  const [scope, setScope] = useState("全部范围");
  const visiblePrograms = useMemo(() => {
    const base = state === "long" ? programs : programs.slice(0, 3);
    return base.filter((program) => {
      const matchesQuery = `${program.title} ${program.description}`.toLowerCase().includes(query.toLowerCase());
      const matchesScope = scope === "全部范围"
        || (scope === "指定用户" && program.publicationState === "selected_users")
        || (scope === "公开" && program.publicationState === "public")
        || (scope === "私有" && program.publicationState === "private");
      return matchesQuery && matchesScope;
    });
  }, [query, scope, state]);

  return (
    <>
      <PageHeader eyebrow="已授权节目" title="浏览你可以访问的节目">
        <Button variant="secondary" icon={<SlidersHorizontal className="h-4 w-4" />}>筛选</Button>
      </PageHeader>
      {state === "loading" ? <LoadingState title="正在加载已授权节目" /> : null}
      {state === "empty" ? <EmptyState title="暂无已授权节目" action="刷新" /> : null}
      {state === "error" ? <ErrorState title="节目库暂不可用" /> : null}
      {state === "denied" ? <PermissionDeniedState /> : null}
      {state === "success" ? <SuccessFeedback message="合集更新已保存到模拟状态。" /> : null}
      {!["loading", "empty", "error", "denied"].includes(state) ? (
        <div className="grid gap-6">
          <div className="grid gap-3 rounded-lg border border-border bg-surface p-4 md:grid-cols-[1fr_220px]">
            <SearchBar placeholder="搜索节目" value={query} onChange={(event) => setQuery(event.target.value)} />
            <Select label="范围" value={scope} onChange={(event) => setScope(event.target.value)} options={["全部范围", "指定用户", "公开", "私有"]} />
          </div>
          {visiblePrograms.length === 0 ? <EmptyState title="没有符合条件的节目" /> : (
            <div className="grid gap-5">
              {visiblePrograms.map((program) => (
                <ProgramCard key={program.id} program={program} />
              ))}
            </div>
          )}
        </div>
      ) : null}
    </>
  );
}
