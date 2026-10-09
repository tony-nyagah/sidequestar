import type { APIRoute } from "astro";
import { GO_API_URL, formAction } from "../../../../lib/api";

export const prerender = false;

export const POST: APIRoute = async (ctx) => {
  const form = await ctx.request.formData();

  const goForm = new FormData();
  const photo = form.get("photo");
  if (photo instanceof File && photo.size > 0) {
    goForm.set("photo", photo, photo.name);
  }

  return formAction(ctx, "Could not complete quest", () =>
    fetch(`${GO_API_URL}/api/quests/${encodeURIComponent(ctx.params.id!)}/complete`, {
      method: "POST",
      body: goForm,
    }),
  );
};
