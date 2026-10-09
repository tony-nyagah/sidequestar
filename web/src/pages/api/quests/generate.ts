import type { APIRoute } from "astro";
import { GO_API_URL } from "../../../lib/api";

export const POST: APIRoute = async ({ request, redirect }) => {
  const form = await request.formData();

  const res = await fetch(`${GO_API_URL}/api/quests/generate`, {
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
