import type { APIRoute } from "astro";
import { goFetch } from "../../../lib/api";
import { SESSION_COOKIE } from "../../../lib/session";

export const POST: APIRoute = async ({ request, redirect, locals, cookies }) => {
  const form = await request.formData();
  const res = await goFetch(locals, "/api/me", {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ password: String(form.get("password") ?? "") }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return redirect(`/profile?error=${encodeURIComponent(body.error ?? "Could not delete your account")}#account`);
  }
  cookies.delete(SESSION_COOKIE, { path: "/" });
  return redirect(`/register?notice=${encodeURIComponent("Your account and everything in it has been deleted.")}`);
};
