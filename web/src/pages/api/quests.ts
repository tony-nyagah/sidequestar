import type { APIRoute } from "astro";
import { goFetch } from "../../lib/api";

export const POST: APIRoute = async ({ request, redirect, locals }) => {
  const form = await request.formData();

  const tags = String(form.get("tags") ?? "")
    .split(",")
    .map((t) => t.trim())
    .filter(Boolean);

  const res = await goFetch(locals, `/api/quests`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      title: String(form.get("title") ?? ""),
      description: String(form.get("description") ?? ""),
      duration_bucket: String(form.get("duration_bucket") ?? ""),
      tags,
    }),
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return redirect(`/?error=${encodeURIComponent(body.error ?? "Could not add quest")}`);
  }
  return redirect("/");
};
