*This is a submission for the [Hacktoberfest Open-Source AI Challenge Week 1: Touch Grass](https://dev.to/challenges/hacktoberfest-week1-2026-10-05)*

## What I Built

**The main quest can wait.** One small escape at a time.

Sidequestar is a local-first app that hands you one small real-life quest at a time — things to do away from work and school — and gives you XP for actually doing them. You tell it your situation ("two hours this evening, want to be outside"), a local open-weight model writes one specific quest, you go do it, and you earn XP and levels for showing up. Add a photo for bonus XP.

It's for anyone who finishes a workday and realises there's more to life than the screen but doesn't know where to start. The screen is the shortest part of the experience: about twenty seconds to get a mission, then you leave.

The quest-giver also gets to know you. Every time you finish a quest, the model rewrites a short summary of what you seem to enjoy and when you're free, and that summary shapes the next quest. It's yours to check: your profile page shows exactly what the model is told about you, word for word, and you can correct the summary by hand. Each person on the machine has their own account, so housemates get their own quests, XP, and summaries.

## Demo

<!-- TODO: deployed link or a short video demo -->

## Code

<!-- TODO: embed the public GitHub repo link (pending the public repo split) -->

## How I Built It

Go backend (stdlib `net/http`), SQLite via `sqlc`, and an Astro + Tailwind CSS + daisyUI frontend. The open-source AI is the core: quest generation and the profile summaries both run through Ollama with the open-weight `llama3.2:3b` model, entirely on the user's machine. `make dev` starts Ollama, the API, and the frontend together.

A few things I learned getting a 3B model to behave:

- **Show the prompt, don't describe it.** The lines the profile page shows come from the same function that builds the prompt, so what you read is exactly what the model gets.
- **Don't feed a model its own guesses.** Early summaries invented details ("free until 10am"), and because each rewrite started from the last one, the inventions stuck around. Now a rewrite only sees what you wrote and the quests you actually finished, plus your last hand edit. Never its own previous output.
- **Starter quests aren't evidence.** The model read the seed quests on your list as things you'd done. Labelling them as not-yet-done and sampling summaries at a low temperature fixed most of it; the summary stays editable for the rest.
- **Errors should say how to fix them.** If Ollama isn't running or the model isn't pulled, the app says so and gives you the command, instead of "something went wrong".

The look is a topographic map sheet: an orange map-cover header, contour lines, your active quests drawn as a footpath route, and completed ones as photo insets. Light and dark follow your system.

## Why Does Open Innovation Matter?

Because the quests — and the personal context used to write them, including a running summary of who you are — never leave the laptop. A closed API would mean sending "I'm burnt out and have two free hours" to a vendor, paying per call, and needing internet in the middle of nowhere. An open-weight model running through Ollama costs nothing to run, works offline, and keeps everything local. And because the model is a swappable open-weight file, the app isn't locked to any provider — swap the model, or drop Ollama for raw llama.cpp, and the code doesn't care.

## My Agent Session

<!-- Optional, but judges love it. Save your session with DevRelay and embed it with the agent_session tag, or link to it. -->

## Prize Categories

<!-- Which partner categories are you entering? List every one that applies, or remove this section. -->
