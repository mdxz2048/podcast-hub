import { createContext, useContext, useMemo, useState } from "react";
import type { ReactNode } from "react";
import { collections as initialCollections } from "./data";
import type { Collection, CollectionRules, UserRssFeed } from "../types/domain";
import type { ToastMessage } from "../components/Toast";
import { ToastViewport } from "../components/Toast";

const initialRssFeeds: UserRssFeed[] = [
  {
    id: "feed-1",
    name: "每日通勤订阅",
    status: "active",
    tokenPrefix: "rss_7fa9",
    createdAt: "2026-06-21 09:15",
    lastUsedAt: "2026-06-27 08:10"
  }
];

type RevealedRssToken = {
  feedId: string;
  token: string;
  url: string;
  action: "created" | "rotated";
};

interface MockStateValue {
  collections: Collection[];
  rssFeeds: UserRssFeed[];
  revealedRssToken: RevealedRssToken | null;
  createCollection: (title: string, programId?: string) => Collection;
  addProgramToCollection: (collectionId: string, programId: string) => void;
  removeProgramFromCollection: (collectionId: string, programId: string) => void;
  moveProgram: (collectionId: string, programId: string, direction: "up" | "down") => void;
  updateCollection: (collectionId: string, patch: Partial<Pick<Collection, "title" | "description" | "programIds">> & { rules?: CollectionRules }) => void;
  resetRssToken: (collectionId: string) => void;
  createRssFeed: (name: string) => RevealedRssToken;
  rotateRssFeed: (feedId: string) => RevealedRssToken | null;
  revokeRssFeed: (feedId: string) => void;
  deleteRssFeed: (feedId: string) => void;
  clearRevealedRssToken: () => void;
  showToast: (toast: Omit<ToastMessage, "id">) => void;
}

const MockStateContext = createContext<MockStateValue | null>(null);

export function MockStateProvider({ children }: { children: ReactNode }) {
  const [collections, setCollections] = useState<Collection[]>(initialCollections);
  const [rssFeeds, setRssFeeds] = useState<UserRssFeed[]>(initialRssFeeds);
  const [revealedRssToken, setRevealedRssToken] = useState<RevealedRssToken | null>(null);
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  function showToast(toast: Omit<ToastMessage, "id">) {
    const message = { ...toast, id: `toast_${Date.now()}_${Math.random().toString(16).slice(2)}` };
    setToasts((current) => [...current.slice(-2), message]);
    window.setTimeout(() => {
      setToasts((current) => current.filter((item) => item.id !== message.id));
    }, 3200);
  }

  const value = useMemo<MockStateValue>(() => ({
    collections,
    rssFeeds,
    revealedRssToken,
    createCollection: (title, programId) => {
      const collection: Collection = {
        id: `collection_mock_${Date.now()}`,
        title: title.trim() || "新的合集",
        description: "由静态原型创建的本地合集。",
        programIds: programId ? [programId] : [],
        accessScope: "private",
        rssTokenState: "active",
        lastUpdatedAt: "刚刚",
        rules: {
          sortOrder: "newest",
          perProgramLimit: 3,
          totalLimit: 8
        }
      };
      setCollections((current) => [collection, ...current]);
      return collection;
    },
    addProgramToCollection: (collectionId, programId) => {
      setCollections((current) => current.map((collection) => {
        if (collection.id !== collectionId || collection.programIds.includes(programId)) return collection;
        return { ...collection, programIds: [...collection.programIds, programId], lastUpdatedAt: "刚刚" };
      }));
    },
    removeProgramFromCollection: (collectionId, programId) => {
      setCollections((current) => current.map((collection) => (
        collection.id === collectionId
          ? { ...collection, programIds: collection.programIds.filter((id) => id !== programId), lastUpdatedAt: "刚刚" }
          : collection
      )));
    },
    moveProgram: (collectionId, programId, direction) => {
      setCollections((current) => current.map((collection) => {
        if (collection.id !== collectionId) return collection;
        const index = collection.programIds.indexOf(programId);
        const target = direction === "up" ? index - 1 : index + 1;
        if (index < 0 || target < 0 || target >= collection.programIds.length) return collection;
        const next = [...collection.programIds];
        [next[index], next[target]] = [next[target], next[index]];
        return { ...collection, programIds: next, lastUpdatedAt: "刚刚" };
      }));
    },
    updateCollection: (collectionId, patch) => {
      setCollections((current) => current.map((collection) => (
        collection.id === collectionId
          ? { ...collection, ...patch, rules: patch.rules ?? collection.rules, lastUpdatedAt: "刚刚" }
          : collection
      )));
    },
    resetRssToken: (collectionId) => {
      setCollections((current) => current.map((collection) => (
        collection.id === collectionId ? { ...collection, rssTokenState: "active", lastUpdatedAt: "刚刚" } : collection
      )));
    },
    createRssFeed: (name) => {
      const now = new Date();
      const feedId = `feed_${Date.now()}`;
      const suffix = Math.random().toString(16).slice(2, 8);
      const token = `rss_${suffix}_${Date.now().toString(36)}`;
      const feed: UserRssFeed = {
        id: feedId,
        name: name.trim() || "新的私有 RSS",
        status: "active",
        tokenPrefix: token.slice(0, 8),
        createdAt: now.toLocaleString("zh-CN", { hour12: false }),
        lastUsedAt: "尚未使用"
      };
      setRssFeeds((current) => [feed, ...current]);
      const revealed = { feedId, token, url: `https://rss.example.invalid/rss/private/${token}.xml`, action: "created" as const };
      setRevealedRssToken(revealed);
      return revealed;
    },
    rotateRssFeed: (feedId) => {
      const now = new Date();
      let revealed: RevealedRssToken | null = null;
      setRssFeeds((current) => current.map((feed) => {
        if (feed.id !== feedId) return feed;
        const suffix = Math.random().toString(16).slice(2, 8);
        const token = `rss_${suffix}_${Date.now().toString(36)}`;
        revealed = { feedId, token, url: `https://rss.example.invalid/rss/private/${token}.xml`, action: "rotated" };
        setRevealedRssToken(revealed);
        return {
          ...feed,
          status: "active",
          tokenPrefix: token.slice(0, 8),
          rotatedAt: now.toLocaleString("zh-CN", { hour12: false })
        };
      }));
      return revealed;
    },
    revokeRssFeed: (feedId) => {
      setRssFeeds((current) => current.map((feed) => (
        feed.id === feedId ? { ...feed, status: "revoked" } : feed
      )));
      setRevealedRssToken((current) => (current?.feedId === feedId ? null : current));
    },
    deleteRssFeed: (feedId) => {
      setRssFeeds((current) => current.filter((feed) => feed.id !== feedId));
      setRevealedRssToken((current) => (current?.feedId === feedId ? null : current));
    },
    clearRevealedRssToken: () => {
      setRevealedRssToken(null);
    },
    showToast
  }), [collections, revealedRssToken, rssFeeds]);

  return (
    <MockStateContext.Provider value={value}>
      {children}
      <ToastViewport messages={toasts} />
    </MockStateContext.Provider>
  );
}

export function useMockState() {
  const value = useContext(MockStateContext);
  if (!value) throw new Error("useMockState must be used within MockStateProvider");
  return value;
}
