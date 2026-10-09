import type { APIRoute } from "astro";
import { GO_API_URL } from "../../../lib/api";

export const prerender = false;

export const GET: APIRoute = async ({ params }) => {
  const res = await fetch(`${GO_API_URL}/api/photos/${encodeURIComponent(params.name ?? "")}`);
  return new Response(res.body, {
    status: res.status,
    headers: {
      "Content-Type": res.headers.get("Content-Type") ?? "application/octet-stream",
      "Cache-Control": "private, max-age=31536000, immutable",
    },
  });
};
