# Sidequestar

**The main quest can wait.** One small escape at a time.

A local-first app that hands you small real-life quests — things to do away from work and school — and rewards you with XP for doing them. You describe your day, a local open-weight model gives you one concrete quest, you go do it, and you earn XP for actually showing up. Over time it learns what you enjoy and writes it down where you can read and correct it.

Built for the [Hacktoberfest Open-Source AI Challenge, Week 1: Touch Grass](https://dev.to/challenges/hacktoberfest-week1-2026-10-05).

## How it works

1. **Summon a quest** — tell the quest-giver your situation ("two hours this evening, I want to be outside").
2. A **local model** (Ollama running `llama3.2:3b`) writes one specific quest, or you add your own.
3. **Do it**, then mark it done — add a photo for bonus XP.
4. **Earn XP and levels** to keep yourself going.

Each person has their own **account**. Your profile page shows your XP, where it came from, and exactly what the quest-giver is told about you: your location, a note in your own words, and a short **summary** the model writes from the quests you finish. The summary updates after every completion; you can rewrite it on demand or correct it by hand, and the model keeps your corrections.

Everything — accounts, quests, XP, photos, summaries — lives in one SQLite file on the machine running the app. Nothing is sent anywhere else.

## Stack

- **Backend**: Go (stdlib `net/http`), SQLite (pure-Go driver) via `sqlc`, Ollama SDK, bcrypt for passwords.
- **Frontend**: Astro (SSR) + Tailwind CSS 4 + daisyUI 5, proxying to the Go API.

## Quick start

Requires Go 1.26+, [Ollama](https://ollama.com), and Bun.

```bash
make setup   # once: installs frontend dependencies and pulls llama3.2:3b
make dev     # starts Ollama (if needed), the API on :8080, and the frontend on :4321
```

Open http://localhost:4321 and create an account. The frontend is also served on your local network, so you can open the **Network** address it prints (for example `http://192.168.1.20:4321`) on your phone to complete quests and add photos. Ctrl-C stops everything. `make test` runs the Go tests and the Astro type check.

**Upgrading from before accounts?** The first account you register inherits your existing quests, XP, and profile. Register yourself first.

### Configuration

| Variable | Read by | Default |
|---|---|---|
| `MODEL` (make) / `OLLAMA_MODEL` | API | `llama3.2:3b` |
| `DATA_DIR` | API | `api/data` (database and photos) |
| `PORT` | API | `8080` |
| `GO_API_URL` | frontend | `http://127.0.0.1:8080` |

For example, `make dev MODEL=gemma2:9b` uses a different local model.

### Security notes

Passwords are stored as bcrypt hashes, and sessions are random tokens kept only as hashes, in 30-day HttpOnly cookies. Sidequestar is built to run on your own machine or local network: there is no password reset, account deletion, or sign-in rate limiting yet, so don't expose it to the internet as it is.

## Why open-source AI is the core

Without the local model there are no quests — generation *is* the product. Running it locally through Ollama means it works offline, costs nothing, and your situation, history, and the summary it writes about you never leave your machine. The model is a swappable open-weight model, so the whole thing stays in your control.

## Repository

```
api/      # Go backend (the Go module lives here)
web/      # Astro frontend
Makefile  # setup, dev, test
```

## The prompt

> ## Touch Grass
>
> **Build something with open-weight models or open-source AI that gets people off the screen and into the world.**
>
> That can mean running an open-weight model, building on an open-source agent harness or framework, running inference locally, or all three. Whatever you pick, the open pieces should be what makes your project work.
>
> Hiking, gardening, birding, run clubs, fall foliage: if it gets someone outside, it counts. The best builds here should make the screen the shortest part of the experience.
