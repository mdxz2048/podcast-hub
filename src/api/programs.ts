import { apiRequest } from "./client";
import type { Episode, Program } from "../types/domain";

export interface UserProgramApi {
  id: string;
  title: string;
  description: string;
  author: string;
  language: string;
  status: "published";
  episode_count: number;
  updated_at: string;
}

export interface UserEpisodeApi {
  id: string;
  program_id: string;
  title: string;
  description: string;
  published_at: string;
  duration_seconds: number;
  status: "published";
  media_status: "published";
}

const coverPairs: Array<[string, string]> = [
  ["#314158", "#B17C3A"],
  ["#245A59", "#C05B43"],
  ["#4C4A36", "#557A46"],
  ["#5A3F52", "#C4A15A"]
];

export async function listPrograms() {
  const result = await apiRequest<{ programs: UserProgramApi[] }>("/programs");
  return result.programs.map(toProgram);
}

export async function getProgram(programId: string) {
  const result = await apiRequest<{ program: UserProgramApi }>(`/programs/${encodeURIComponent(programId)}`);
  return toProgram(result.program);
}

export async function listProgramEpisodes(programId: string) {
  const result = await apiRequest<{ episodes: UserEpisodeApi[] }>(`/programs/${encodeURIComponent(programId)}/episodes`);
  return result.episodes.map(toEpisode);
}

export async function getEpisode(episodeId: string) {
  const result = await apiRequest<{ episode: UserEpisodeApi }>(`/episodes/${encodeURIComponent(episodeId)}`);
  return toEpisode(result.episode);
}

export function toProgram(item: UserProgramApi): Program {
  return {
    id: item.id,
    title: item.title,
    description: item.description,
    author: item.author,
    category: "已授权",
    language: item.language,
    updateFrequency: "按发布更新",
    coverTone: coverPairs[Math.abs(hashString(item.id)) % coverPairs.length],
    status: "active",
    rightsState: "clear",
    publicationState: "selected_users",
    episodeCount: item.episode_count,
    sourceCount: 0,
    accessState: "authorized",
    lastUpdated: formatDate(item.updated_at)
  };
}

export function toEpisode(item: UserEpisodeApi): Episode {
  return {
    id: item.id,
    programId: item.program_id,
    title: item.title,
    publishedAt: formatDate(item.published_at),
    duration: formatDuration(item.duration_seconds),
    summary: item.description
  };
}

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" });
}

function formatDuration(seconds: number) {
  const minutes = Math.floor(seconds / 60);
  const remaining = seconds % 60;
  return `${minutes}:${String(remaining).padStart(2, "0")}`;
}

function hashString(value: string) {
  let hash = 0;
  for (const char of value) hash = ((hash << 5) - hash + char.charCodeAt(0)) | 0;
  return hash;
}
