import type { APIRoute } from "astro";
import { goFetch } from "../../lib/api";

export const POST: APIRoute = async ({ request, redirect, locals }) => {
  const form = await request.formData();

  const res = await goFetch(locals, `/api/profile`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    // Only send what this form had, so the about form and the summary form don't overwrite each other.
    body: JSON.stringify(
      Object.fromEntries(
        ["location", "note", "summary"].filter((k) => form.has(k)).map((k) => [k, String(form.get(k))]),
      ),
    ),
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return redirect(`/profile?error=${encodeURIComponent(body.error ?? "Could not update profile")}`);
  }
  return redirect("/profile");
};
