import type { APIRoute } from "astro";
import { GO_API_URL, formAction } from "../../../../lib/api";

export const prerender = false;

export const POST: APIRoute = async (ctx) =>
  formAction(ctx, "Could not remove quest", () =>
    fetch(`${GO_API_URL}/api/quests/${encodeURIComponent(ctx.params.id!)}`, {
      method: "DELETE",
    }),
  );
