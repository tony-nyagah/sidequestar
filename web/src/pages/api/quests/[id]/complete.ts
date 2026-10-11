import type { APIRoute } from "astro";
import { goFetch } from "../../../../lib/api";

export const prerender = false;

export const POST: APIRoute = async ({ request, params, redirect, locals }) => {
  const form = await request.formData();

  const goForm = new FormData();
  const photo = form.get("photo");
  if (photo instanceof File && photo.size > 0) {
    goForm.set("photo", photo, photo.name);
  }

  const res = await goFetch(locals, `/api/quests/${params.id}/complete`, {
    method: "POST",
    body: goForm,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    return redirect(`/?error=${encodeURIComponent(body.error ?? "Could not complete quest")}`);
  }
  return redirect("/");
};
