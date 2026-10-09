import type { APIRoute } from "astro";
import { GO_API_URL } from "../../lib/api";

export const POST: APIRoute = async ({ request, redirect }) => {
  const form = await request.formData();

  const tags = String(form.get("tags") ?? "")
    .split(",")
    .map((t) => t.trim())
    .filter(Boolean);

  const res = await fetch(`${GO_API_URL}/api/quests`, {
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
