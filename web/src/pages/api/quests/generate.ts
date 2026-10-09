import type { APIRoute } from "astro";
import { formAction, postJSON } from "../../../lib/api";

export const POST: APIRoute = async (ctx) => {
  const form = await ctx.request.formData();

  return formAction(ctx, "Could not generate a quest", () =>
    postJSON(
      "/api/quests/generate",
      { context: String(form.get("context") ?? "") },
      // Slightly longer than the API's own generation timeout.
      { signal: AbortSignal.timeout(100_000) },
    ),
  );
};
