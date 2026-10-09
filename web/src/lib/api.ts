import type { APIContext, AstroCookies } from "astro";
import { GO_API_URL } from "astro:env/server";

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

export { GO_API_URL };

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

const FLASH_COOKIE = "flash";

/** Reads and clears the one-shot error message set by a form action. */
export function takeFlash(cookies: AstroCookies): string | undefined {
  const msg = cookies.get(FLASH_COOKIE)?.value;
  if (msg) cookies.delete(FLASH_COOKIE, { path: "/" });
  return msg;
}

/**
 * Runs a form action against the Go API and redirects back home. Errors are
 * passed through a short-lived cookie rather than the URL, so they don't
 * linger on refresh and can't be injected via a crafted link.
 */
export async function formAction(
  { cookies, redirect }: APIContext,
  fallbackError: string,
  send: () => Promise<Response>,
): Promise<Response> {
  let error: string | undefined;
  try {
    const res = await send();
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      error = body.error ?? fallbackError;
    }
  } catch {
    error = "Can't reach the API. Is it running?";
  }
  if (error) {
    cookies.set(FLASH_COOKIE, error, { path: "/", httpOnly: true, sameSite: "lax", maxAge: 60 });
  }
  return redirect("/", 303);
}

export function postJSON(path: string, body: unknown, init?: RequestInit): Promise<Response> {
  return fetch(`${GO_API_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    ...init,
  });
}
