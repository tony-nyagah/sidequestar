import type { APIRoute } from "astro";
import { goFetch } from "../../../lib/api";
import { SESSION_COOKIE } from "../../../lib/session";

export const POST: APIRoute = async ({ redirect, locals, cookies }) => {
  await goFetch(locals, "/api/auth/logout", { method: "POST" }).catch(() => null);
  cookies.delete(SESSION_COOKIE, { path: "/" });
  return redirect("/login");
};
