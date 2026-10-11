import type { APIRoute } from "astro";
import { goFetch } from "../../../lib/api";

export const prerender = false;

export const GET: APIRoute = async ({ params, locals }) => {
  const res = await goFetch(locals, `/api/photos/${encodeURIComponent(params.name ?? "")}`);
  return new Response(res.body, {
    status: res.status,
    headers: {
      "Content-Type": res.headers.get("Content-Type") ?? "application/octet-stream",
      "Cache-Control": "private, max-age=31536000, immutable",
    },
  });
};
