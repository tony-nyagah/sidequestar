import type { AstroCookies } from "astro";

export const SESSION_COOKIE = "sq_session";

export function setSessionCookie(cookies: AstroCookies, url: URL, token: string, expiresAt: string) {
  cookies.set(SESSION_COOKIE, token, {
    path: "/",
    httpOnly: true,
    sameSite: "lax",
    secure: url.protocol === "https:",
    expires: new Date(expiresAt),
  });
}
