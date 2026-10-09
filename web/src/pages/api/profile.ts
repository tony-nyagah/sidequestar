import type { APIRoute } from "astro";
import { formAction, postJSON } from "../../lib/api";

export const POST: APIRoute = async (ctx) => {
  const form = await ctx.request.formData();

  return formAction(ctx, "Could not update profile", () =>
    postJSON(
      "/api/profile",
      {
        location: String(form.get("location") ?? ""),
        note: String(form.get("note") ?? ""),
      },
      { method: "PATCH" },
    ),
  );
};
