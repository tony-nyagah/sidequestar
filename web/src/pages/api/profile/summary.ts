import type { APIRoute } from "astro";
import { goFetch } from "../../../lib/api";

export const POST: APIRoute = async ({ redirect, locals }) => {
  const res = await goFetch(locals, "/api/profile/summary", { method: "POST" });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return redirect(`/profile?error=${encodeURIComponent(body.error ?? "Could not refresh the summary")}`);
  }
  return redirect("/profile");
};
