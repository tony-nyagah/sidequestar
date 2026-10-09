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

export interface Profile {
  location: string;
  note: string;
  interests: string[];
  xp_total: number;
  completed_count: number;
  level: number;
  xp_into_level: number;
  xp_needed: number;
  title: string;
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

export const GO_API_URL: string =
  import.meta.env.GO_API_URL ?? "http://127.0.0.1:8080";

export async function listQuests(status?: QuestStatus): Promise<Quest[]> {
  const url = status
    ? `${GO_API_URL}/api/quests?status=${status}`
    : `${GO_API_URL}/api/quests`;
  const res = await fetch(url);
  if (!res.ok) throw new Error(`list quests: ${res.status}`);
  return res.json();
}

export async function getProfile(): Promise<Profile> {
  const res = await fetch(`${GO_API_URL}/api/profile`);
  if (!res.ok) throw new Error(`get profile: ${res.status}`);
  return res.json();
}
