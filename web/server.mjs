import http from "node:http";
import compression from "compression";

process.env.ASTRO_NODE_AUTOSTART = "disabled";
const { handler } = await import("./dist/server/entry.mjs");

const compress = compression();
const port = Number(process.env.PORT ?? 4321);
// Localhost by default; set HOST=0.0.0.0 to open it to your phone on the same network.
const host = process.env.HOST ?? "127.0.0.1";

http
  .createServer((req, res) => compress(req, res, () => handler(req, res)))
  .listen(port, host, () => console.log(`Sidequestar listening on http://${host}:${port}`));
