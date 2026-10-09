import type { APIRoute } from "astro";
import { GO_API_URL } from "../../../lib/api";

export const prerender = false;

export const GET: APIRoute = async ({ params }) => {
  const res = await fetch(`${GO_API_URL}/api/photos/${encodeURIComponent(params.name!)}`);
  if (!res.ok) return new Response("Not found", { status: 404 });

  const headers = new Headers();
  for (const h of ["Content-Type", "Content-Length", "Cache-Control", "Last-Modified"]) {
    const v = res.headers.get(h);
    if (v) headers.set(h, v);
  }
  headers.set("X-Content-Type-Options", "nosniff");
  return new Response(res.body, { status: 200, headers });
};
