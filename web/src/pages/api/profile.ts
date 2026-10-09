import type { APIRoute } from "astro";
import { GO_API_URL } from "../../lib/api";

export const POST: APIRoute = async ({ request, redirect }) => {
  const form = await request.formData();

  const res = await fetch(`${GO_API_URL}/api/profile`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      location: String(form.get("location") ?? ""),
      note: String(form.get("note") ?? ""),
    }),
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return redirect(`/?error=${encodeURIComponent(body.error ?? "Could not update profile")}`);
  }
  return redirect("/");
};
