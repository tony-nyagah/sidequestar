import type { APIRoute } from "astro";
import { goFetch } from "../../../lib/api";
import { setSessionCookie } from "../../../lib/session";

export const POST: APIRoute = async ({ request, redirect, locals, cookies, url }) => {
  const form = await request.formData();
  const username = String(form.get("username") ?? "");

  const res = await goFetch(locals, "/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password: String(form.get("password") ?? "") }),
  }).catch(() => null);

  if (!res?.ok) {
    const body = res ? await res.json().catch(() => ({})) : { error: "Can't reach the API. Start it with make dev." };
    const params = new URLSearchParams({ error: body.error ?? "Something went wrong. Try again.", username });
    return redirect(`/login?${params}`);
  }
  const session = await res.json();
  setSessionCookie(cookies, url, session.token, session.expires_at);
  return redirect("/");
};
