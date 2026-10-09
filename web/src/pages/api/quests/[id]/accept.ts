import type { APIRoute } from "astro";
import { GO_API_URL } from "../../../../lib/api";

export const prerender = false;

export const POST: APIRoute = async ({ params, redirect }) => {
  const res = await fetch(`${GO_API_URL}/api/quests/${params.id}/accept`, {
    method: "POST",
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return redirect(`/?error=${encodeURIComponent(body.error ?? "Could not accept quest")}`);
  }
  return redirect("/");
};
