import type { APIRoute } from "astro";
import { goFetch } from "../../../lib/api";

export const POST: APIRoute = async ({ request, redirect, locals }) => {
  const form = await request.formData();

  const res = await goFetch(locals, `/api/quests/generate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ context: String(form.get("context") ?? "") }),
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return redirect(`/?error=${encodeURIComponent(body.error ?? "Could not generate a quest")}`);
  }
  return redirect("/");
};
