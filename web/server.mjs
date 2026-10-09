import http from "node:http";
import compression from "compression";

process.env.ASTRO_NODE_AUTOSTART = "disabled";
const { handler } = await import("./dist/server/entry.mjs");

const compress = compression();
const port = Number(process.env.PORT ?? 4321);
const host = process.env.HOST ?? "0.0.0.0";

http
  .createServer((req, res) => compress(req, res, () => handler(req, res)))
  .listen(port, host, () => console.log(`Sidequestar listening on http://${host}:${port}`));
