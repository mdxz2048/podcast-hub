import type { Collection, Episode, Program } from "../types/domain";
import { episodes, programs } from "./data";

export function findProgram(programId: string | undefined) {
  return programs.find((program) => program.id === programId);
}

export function episodesForProgram(programId: string) {
  return episodes.filter((episode) => episode.programId === programId);
}

export function programsForCollection(collection: Collection, allPrograms: Program[] = programs) {
  return collection.programIds
    .map((programId) => allPrograms.find((program) => program.id === programId))
    .filter((program): program is Program => Boolean(program));
}

export function previewEpisodesForCollection(collection: Collection): Episode[] {
  const limitedByProgram = collection.programIds.flatMap((programId) => {
    const list = episodesForProgram(programId);
    const sorted = sortEpisodes(list, collection.rules.sortOrder);
    return sorted.slice(0, collection.rules.perProgramLimit);
  });

  return sortEpisodes(limitedByProgram, collection.rules.sortOrder).slice(0, collection.rules.totalLimit);
}

export function sortEpisodes(list: Episode[], order: Collection["rules"]["sortOrder"]) {
  return [...list].sort((a, b) => {
    const diff = new Date(b.publishedAt).getTime() - new Date(a.publishedAt).getTime();
    return order === "newest" ? diff : -diff;
  });
}

export function programTitle(programId: string) {
  return programs.find((program) => program.id === programId)?.title ?? "未知节目";
}
