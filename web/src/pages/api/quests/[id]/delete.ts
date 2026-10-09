import type { APIRoute } from "astro";
import { GO_API_URL } from "../../../../lib/api";

export const prerender = false;

export const POST: APIRoute = async ({ params, redirect }) => {
  await fetch(`${GO_API_URL}/api/quests/${params.id}`, { method: "DELETE" });
  return redirect("/");
};
