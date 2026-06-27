import { Plus, Radio } from "lucide-react";
import type { CSSProperties } from "react";
import { Link } from "react-router-dom";
import type { Program } from "../types/domain";
import { programStatusLabel, publicationStateLabel } from "../utils/labels";
import { Badge } from "./Badge";
import { Button } from "./Button";

function statusTone(program: Program) {
  if (program.status === "rights_hold") return "danger" as const;
  if (program.status === "draft") return "warning" as const;
  return "success" as const;
}

export function ProgramCard({ program }: { program: Program }) {
  return (
    <article className="grid overflow-hidden rounded-lg border border-border bg-surface shadow-subtle md:grid-cols-[148px_1fr]">
      <div
        className="cover-art relative min-h-40 md:min-h-full"
        style={{ "--cover-a": program.coverTone[0], "--cover-b": program.coverTone[1] } as CSSProperties}
        aria-hidden="true"
      />
      <div className="grid gap-4 p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <Link to={`/programs/${program.id}`} className="break-words text-lg font-semibold leading-tight hover:text-action">
              {program.title}
            </Link>
            <p className="mt-2 line-clamp-3 text-sm text-secondary">{program.description}</p>
          </div>
          <Badge tone={statusTone(program)}>{programStatusLabel[program.status]}</Badge>
        </div>
        <div className="flex flex-wrap gap-2 text-xs text-muted">
          <span>{program.episodeCount} 个单集</span>
          <span>{program.language}</span>
          <span>{publicationStateLabel[program.publicationState]}</span>
          <span>更新于 {program.lastUpdated}</span>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link to={`/programs/${program.id}`}>
            <Button variant="secondary">查看详情</Button>
          </Link>
          <Button icon={<Radio className="h-4 w-4" />}>订阅</Button>
          <Button variant="secondary" icon={<Plus className="h-4 w-4" />}>加入合集</Button>
        </div>
      </div>
    </article>
  );
}
