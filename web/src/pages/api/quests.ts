import type { APIRoute } from "astro";
import { formAction, postJSON } from "../../lib/api";

export const POST: APIRoute = async (ctx) => {
  const form = await ctx.request.formData();

  const tags = String(form.get("tags") ?? "")
    .split(",")
    .map((t) => t.trim())
    .filter(Boolean);

  return formAction(ctx, "Could not add quest", () =>
    postJSON("/api/quests", {
      title: String(form.get("title") ?? ""),
      description: String(form.get("description") ?? ""),
      duration_bucket: String(form.get("duration_bucket") ?? ""),
      tags,
    }),
  );
};
