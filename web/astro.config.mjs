// @ts-check
import { defineConfig } from "astro/config";

import tailwindcss from "@tailwindcss/vite";
import icon from "astro-icon";
import node from "@astrojs/node";

// https://astro.build/config
export default defineConfig({
  site: "https://sidequestar.nyagah.me",
  output: "server",
  // Behind the HTTPS proxy, trust its forwarded host and protocol for this domain only,
  // so Astro sees https:// (and the session cookie gets the Secure flag).
  security: {
    allowedDomains: [{ hostname: "sidequestar.nyagah.me", protocol: "https" }],
  },
  adapter: node({ mode: "standalone" }),
  integrations: [icon()],
  vite: {
    plugins: [tailwindcss()],
  },
});
