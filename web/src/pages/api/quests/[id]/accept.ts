import type { APIRoute } from "astro";
import { goFetch } from "../../../../lib/api";

export const prerender = false;

export const POST: APIRoute = async ({ params, redirect, locals }) => {
  const res = await goFetch(locals, `/api/quests/${params.id}/accept`, {
    method: "POST",
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return redirect(`/?error=${encodeURIComponent(body.error ?? "Could not accept quest")}`);
  }
  return redirect("/");
};
