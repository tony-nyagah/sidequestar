/// <reference types="astro/client" />

declare namespace App {
  interface Locals {
    token?: string;
    clientIP?: string;
    user?: import("./lib/api").User;
  }
}
