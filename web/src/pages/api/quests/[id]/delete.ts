import type { APIRoute } from "astro";
import { goFetch } from "../../../../lib/api";

export const prerender = false;

export const POST: APIRoute = async ({ params, redirect, locals }) => {
  await goFetch(locals, `/api/quests/${params.id}`, { method: "DELETE" });
  return redirect("/");
};
