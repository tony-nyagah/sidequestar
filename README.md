# Sidequestar

**The main quest can wait.** One small escape at a time.

Sidequestar gives you small things to do away from work and school, and XP for doing them. You describe your day, a model running on your own machine suggests one concrete quest, and you go do it. Over time it learns what you enjoy and writes that down where you can read it and correct it.

Built for the [Hacktoberfest Open-Source AI Challenge, Week 1: Touch Grass](https://dev.to/challenges/hacktoberfest-week1-2026-10-05).

## How it works

1. Tell the quest-giver your situation, for example "two hours this evening, I want to be outside".
2. A local model (Ollama running `llama3.2:3b`) writes one specific quest. You can also add your own.
3. Go do it, then mark it done. Adding a photo earns bonus XP.
4. XP adds up to levels.

Everyone gets their own account. Your profile page shows your XP, where it came from, and exactly what the quest-giver is told about you: your location, a note in your own words, and a short summary the model writes from the quests you finish. The summary updates after every quest you complete. You can ask for a fresh one at any time or edit it yourself, and the model keeps your edits.

Accounts, quests, XP, photos and summaries all live in one SQLite file on the machine running the app, and none of it is sent anywhere else.

## Stack

- Backend: Go (stdlib `net/http`), SQLite (pure-Go driver) via `sqlc`, the Ollama SDK, and bcrypt for passwords.
- Frontend: Astro (SSR) with Tailwind CSS 4 and daisyUI 5, proxying to the Go API.

## Quick start

You need Go 1.26 or later, [Ollama](https://ollama.com), and Bun.

```bash
make setup   # once: installs frontend dependencies and pulls llama3.2:3b
make dev     # starts Ollama (if needed), the API on :8080 and the frontend on :4321
```

Open http://localhost:4321 and create an account. Ctrl-C stops everything, and `make test` runs the Go tests and the Astro type check.

If you used Sidequestar before it had accounts, register yourself first: the first account inherits your existing quests, XP and profile.

### Configuration

| Variable | Read by | Default |
|---|---|---|
| `MODEL` (make) / `OLLAMA_MODEL` | API | `llama3.2:3b` |
| `DATA_DIR` | API | `api/data` (database and photos) |
| `PORT` | API | `8080` |
| `API_HOST` | API | `127.0.0.1` (only this machine; the frontend talks to it) |
| `GO_API_URL` | frontend | `http://127.0.0.1:8080` |
| `TRUST_PROXY` | frontend | unset (set it behind a reverse proxy that sends the visitor's address in `X-Client-IP`, so sign-in limits see real addresses) |

For example, `make dev MODEL=gemma2:9b` uses a different local model.

### Security notes

Passwords are stored as bcrypt hashes. Sessions are random tokens, stored only as hashes and kept in 30-day HttpOnly cookies (Secure when served over HTTPS). Sign-in is limited to 5 failed attempts per username and 20 per address every 15 minutes, and an address can create 5 accounts an hour. The model handles one request at a time, and each account can ask for 30 quests an hour, so one person can't keep the CPU busy for everyone else.

You can change your password or delete your account (quests, photos, summary and all) from the profile page. There's no email, so if someone forgets their password, whoever runs the server resets it:

```bash
cd api && go run ./cmd/server reset-password <username>
```

That prints a new password and signs the account out everywhere.

## Why the model runs locally

The quests come from the model, so without it there's no app. Running it through Ollama means it works offline and costs nothing, and your situation, your history and the summary it writes about you stay on your machine. It's an open-weight model you can swap for another, so you aren't tied to any provider.

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
