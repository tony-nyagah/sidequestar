import { defineMiddleware } from "astro:middleware";
import { goFetch } from "./lib/api";
import type { User } from "./lib/api";
import { SESSION_COOKIE } from "./lib/session";

const PUBLIC = ["/login", "/register", "/api/auth/login", "/api/auth/register"];

// Every page and proxy route needs a signed-in user; the session token lives in an
// HttpOnly cookie and is checked against the Go API on each request.
export const onRequest = defineMiddleware(async (context, next) => {
  const { pathname } = context.url;
  if (PUBLIC.includes(pathname) || pathname.startsWith("/_astro/") || pathname === "/contours.svg") {
    return next();
  }

  const token = context.cookies.get(SESSION_COOKIE)?.value;
  if (token) {
    context.locals.token = token;
    const res = await goFetch(context.locals, "/api/me").catch(() => null);
    if (res?.ok) {
      context.locals.user = (await res.json()) as User;
      return next();
    }
    if (res === null) {
      // API is down; let the page render its own "can't reach the API" message.
      return next();
    }
    context.cookies.delete(SESSION_COOKIE, { path: "/" });
    context.locals.token = undefined;
  }
  return context.redirect("/login");
});
