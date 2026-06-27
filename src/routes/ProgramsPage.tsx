import { SlidersHorizontal } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { listPrograms } from "../api/programs";
import { Button } from "../components/Button";
import { SearchBar, Select } from "../components/Form";
import { PageHeader } from "../components/PageShell";
import { ProgramCard } from "../components/ProgramCard";
import { EmptyState, ErrorState, LoadingState, PermissionDeniedState, SuccessFeedback } from "../components/StateBlocks";
import type { Program } from "../types/domain";
import { useViewState } from "./state";

export function ProgramsPage() {
  const state = useViewState();
  const [programs, setPrograms] = useState<Program[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [query, setQuery] = useState("");
  const [scope, setScope] = useState("全部范围");

  useEffect(() => {
    listPrograms()
      .then((items) => {
        setPrograms(items);
        setError(false);
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false));
  }, []);

  const visiblePrograms = useMemo(() => {
    return programs.filter((program) => {
      const matchesQuery = `${program.title} ${program.description}`.toLowerCase().includes(query.toLowerCase());
      const matchesScope = scope === "全部范围"
        || (scope === "指定用户" && program.publicationState === "selected_users")
        || (scope === "公开" && program.publicationState === "public")
        || (scope === "私有" && program.publicationState === "private");
      return matchesQuery && matchesScope;
    });
  }, [programs, query, scope]);

  return (
    <>
      <PageHeader eyebrow="已授权节目" title="浏览你可以访问的节目">
        <Button variant="secondary" icon={<SlidersHorizontal className="h-4 w-4" />}>筛选</Button>
      </PageHeader>
      {loading || state === "loading" ? <LoadingState title="正在加载已授权节目" /> : null}
      {!loading && !error && programs.length === 0 ? <EmptyState title="暂无已授权节目" action="刷新" /> : null}
      {error || state === "error" ? <ErrorState title="节目库暂不可用" /> : null}
      {state === "denied" ? <PermissionDeniedState /> : null}
      {state === "success" ? <SuccessFeedback message="合集更新已保存。" /> : null}
      {!loading && !error && state !== "denied" && programs.length > 0 ? (
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
