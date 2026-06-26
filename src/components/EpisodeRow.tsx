import { CalendarDays, Clock3 } from "lucide-react";
import type { Episode } from "../types/domain";

export function EpisodeRow({ episode, programTitle }: { episode: Episode; programTitle?: string }) {
  return (
    <article className="rounded-lg border border-border bg-surface p-4">
      {programTitle ? <p className="mb-1 text-xs text-muted">{programTitle}</p> : null}
      <h3 className="font-semibold leading-snug">{episode.title}</h3>
      <p className="mt-2 line-clamp-2 text-sm text-secondary">{episode.summary}</p>
      <div className="mt-3 flex flex-wrap gap-4 text-xs text-muted">
        <span className="flex items-center gap-1"><CalendarDays className="h-4 w-4" />{episode.publishedAt}</span>
        <span className="flex items-center gap-1"><Clock3 className="h-4 w-4" />{episode.duration}</span>
      </div>
    </article>
  );
}
