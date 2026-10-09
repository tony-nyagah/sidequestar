# Sidequestar

A local-first app that hands you small real-life quests — things to do away from work and school — and rewards you with XP for doing them. You describe your day, a local open-weight model gives you one concrete quest, you go do it, and you earn XP for actually showing up.

Built for the [Hacktoberfest Open-Source AI Challenge, Week 1: Touch Grass](https://dev.to/challenges/hacktoberfest-week1-2026-10-05).

## How it works

1. **Summon a quest** — tell the quest-giver your situation ("two hours this evening, I want to be outside").
2. A **local model** (Ollama running `llama3.2:3b`) writes one specific quest, or you add your own.
3. **Do it**, then mark it done — add a photo for bonus XP.
4. **Earn XP and levels** to keep yourself going.

Quests, XP, photos, and your profile live in a SQLite file on your machine. Nothing leaves it.

## Stack

- **Backend**: Go (stdlib `net/http`), SQLite (pure-Go driver) via `sqlc`, Ollama SDK.
- **Frontend**: Astro (SSR) + Tailwind CSS 4 + daisyUI 5, proxying to the Go API.

## Quick start

Requires Go 1.26+, [Ollama](https://ollama.com) with `llama3.2:3b` pulled, and Bun (or Node).

```bash
# 1. Pull the model (once)
ollama pull llama3.2:3b

# 2. Start the Go API (port 8080)
cd api
make run

# 3. Start the frontend (port 4321)
cd ../web
bun install
bun run dev
```

Open http://localhost:4321. The Go API reads `OLLAMA_MODEL` and `PORT`; the frontend reads `GO_API_URL` (defaults to `http://127.0.0.1:8080`).

## Why open-source AI is the core

Without the local model there are no quests — generation *is* the product. Running it locally through Ollama means it works offline, costs nothing, and your situation and history never leave your machine. The model is a swappable open-weight model, so the whole thing stays in your control. See [ADR 0002](docs/adr/0002-local-open-weight-generation.md).

## Repository

```
api/   # Go backend (the Go module lives here)
web/   # Astro frontend
docs/  # ADRs
```

## The prompt

> ## Touch Grass
>
> **Build something with open-weight models or open-source AI that gets people off the screen and into the world.**
>
> That can mean running an open-weight model, building on an open-source agent harness or framework, running inference locally, or all three. Whatever you pick, the open pieces should be what makes your project work.
>
> Hiking, gardening, birding, run clubs, fall foliage: if it gets someone outside, it counts. The best builds here should make the screen the shortest part of the experience.
