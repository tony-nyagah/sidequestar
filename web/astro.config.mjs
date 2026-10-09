// @ts-check
import { defineConfig, envField } from "astro/config";

import tailwindcss from "@tailwindcss/vite";
import icon from "astro-icon";
import node from "@astrojs/node";

// https://astro.build/config
export default defineConfig({
  output: "server",
  adapter: node({ mode: "standalone" }),
  integrations: [icon()],
  env: {
    schema: {
      // "secret" server vars are read from process.env at runtime instead of
      // being inlined at build time, so the URL can change per deployment.
      GO_API_URL: envField.string({
        context: "server",
        access: "secret",
        default: "http://127.0.0.1:8080",
      }),
    },
  },
  vite: {
    plugins: [tailwindcss()],
  },
});
