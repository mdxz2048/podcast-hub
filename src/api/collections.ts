import { apiRequest, readCookie } from "./client";
import type { Collection } from "../types/domain";
import type { UserProgramApi } from "./programs";
import { toProgram } from "./programs";

export interface UserCollectionApi {
  id: string;
  title: string;
  description: string;
  programs: UserProgramApi[];
  created_at: string;
  updated_at: string;
}

export interface UserCollectionView extends Collection {
  programs: ReturnType<typeof toProgram>[];
}

function csrfHeaders() {
  const csrf = readCookie("podcast_hub_csrf");
  return csrf ? { "X-CSRF-Token": csrf } : undefined;
}

export async function listCollections() {
  const result = await apiRequest<{ collections: UserCollectionApi[] }>("/me/collections");
  return result.collections.map(toCollection);
}

export async function createCollection(body: { title: string; description?: string }) {
  const result = await apiRequest<{ collection: UserCollectionApi }>("/me/collections", {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify(body)
  });
  return toCollection(result.collection);
}

export async function updateCollection(collectionId: string, body: { title?: string; description?: string }) {
  const result = await apiRequest<{ collection: UserCollectionApi }>(`/me/collections/${encodeURIComponent(collectionId)}`, {
    method: "PATCH",
    headers: csrfHeaders(),
    body: JSON.stringify(body)
  });
  return toCollection(result.collection);
}

export async function deleteCollection(collectionId: string) {
  await apiRequest<void>(`/me/collections/${encodeURIComponent(collectionId)}`, { method: "DELETE", headers: csrfHeaders() });
}

export async function addProgram(collectionId: string, programId: string) {
  const result = await apiRequest<{ collection: UserCollectionApi }>(`/me/collections/${encodeURIComponent(collectionId)}/programs`, {
    method: "POST",
    headers: csrfHeaders(),
    body: JSON.stringify({ program_id: programId })
  });
  return toCollection(result.collection);
}

export async function removeProgram(collectionId: string, programId: string) {
  const result = await apiRequest<{ collection: UserCollectionApi }>(`/me/collections/${encodeURIComponent(collectionId)}/programs/${encodeURIComponent(programId)}`, {
    method: "DELETE",
    headers: csrfHeaders()
  });
  return toCollection(result.collection);
}

function toCollection(item: UserCollectionApi): UserCollectionView {
  const programs = item.programs.map(toProgram);
  return {
    id: item.id,
    title: item.title,
    description: item.description,
    programIds: programs.map((program) => program.id),
    programs,
    accessScope: "private",
    rssTokenState: "active",
    lastUpdatedAt: new Date(item.updated_at).toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" }),
    rules: { sortOrder: "newest", perProgramLimit: 3, totalLimit: 12 }
  };
}
