// @ts-check
import { defineConfig } from "astro/config";

import tailwindcss from "@tailwindcss/vite";
import icon from "astro-icon";
import node from "@astrojs/node";

// https://astro.build/config
export default defineConfig({
  site: "https://sidequestar.example",
  output: "server",
  adapter: node({ mode: "standalone" }),
  integrations: [icon()],
  vite: {
    plugins: [tailwindcss()],
  },
});
