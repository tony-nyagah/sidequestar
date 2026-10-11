export type QuestStatus = "generated" | "active" | "completed";
export type DurationBucket =
  | "under-30m"
  | "30-60m"
  | "1-2h"
  | "half-day"
  | "full-day";

export interface Quest {
  id: string;
  title: string;
  description: string;
  duration_bucket: DurationBucket;
  source: "generated" | "hand" | "seed";
  status: QuestStatus;
  tags: string[];
  photo_path?: string;
  xp_awarded?: number;
  created_at: string;
  completed_at?: string;
}

export interface User {
  id: string;
  username: string;
}

export interface Profile {
  username: string;
  location: string;
  note: string;
  summary: string;
  summary_updated_at?: string;
  interests: string[];
  xp_total: number;
  completed_count: number;
  level: number;
  xp_into_level: number;
  xp_needed: number;
  title: string;
  known_facts: string[];
  model: string;
}

export const DURATION_BUCKETS: { value: DurationBucket; label: string }[] = [
  { value: "under-30m", label: "Under 30 minutes" },
  { value: "30-60m", label: "30 to 60 minutes" },
  { value: "1-2h", label: "1 to 2 hours" },
  { value: "half-day", label: "Half a day" },
  { value: "full-day", label: "Full day" },
];

export function durationLabel(bucket: string): string {
  return DURATION_BUCKETS.find((b) => b.value === bucket)?.label ?? bucket;
}

// Read at runtime: import.meta.env is baked in at build time, which would make
// GO_API_URL impossible to change for a built frontend.
export const GO_API_URL: string =
  (process.env as Record<string, string | undefined>).GO_API_URL ??
  import.meta.env.GO_API_URL ??
  "http://127.0.0.1:8080";

// goFetch calls the Go API as the signed-in user (middleware puts their token in locals).
export function goFetch(locals: App.Locals, path: string, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers);
  if (locals.token) headers.set("Authorization", `Bearer ${locals.token}`);
  if (locals.clientIP) headers.set("X-Client-IP", locals.clientIP);
  return fetch(`${GO_API_URL}${path}`, { ...init, headers });
}

export async function listQuests(locals: App.Locals, status?: QuestStatus): Promise<Quest[]> {
  const res = await goFetch(locals, status ? `/api/quests?status=${status}` : "/api/quests");
  if (!res.ok) throw new Error(`list quests: ${res.status}`);
  return res.json();
}

export async function getProfile(locals: App.Locals): Promise<Profile> {
  const res = await goFetch(locals, "/api/profile");
  if (!res.ok) throw new Error(`get profile: ${res.status}`);
  return res.json();
}

export function durationIndex(bucket: string): number {
  return Math.max(0, DURATION_BUCKETS.findIndex((b) => b.value === bucket));
}
