import type { APIRoute } from "astro";
import { goFetch } from "../../../lib/api";

export const POST: APIRoute = async ({ request, redirect, locals }) => {
  const form = await request.formData();
  const next = String(form.get("new_password") ?? "");
  if (next !== String(form.get("confirm_password") ?? "")) {
    return redirect(`/profile?error=${encodeURIComponent("The new passwords don't match.")}#account`);
  }
  const res = await goFetch(locals, "/api/me/password", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ current_password: String(form.get("current_password") ?? ""), new_password: next }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return redirect(`/profile?error=${encodeURIComponent(body.error ?? "Could not change your password")}#account`);
  }
  return redirect(`/profile?notice=${encodeURIComponent("Password changed. Your other devices have been signed out.")}#account`);
};
