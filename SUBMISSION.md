*This is a submission for the [Hacktoberfest Open-Source AI Challenge Week 1: Touch Grass](https://dev.to/challenges/hacktoberfest-week1-2026-10-05)*

## What I Built

Sidequestar is a local-first app that hands you one small real-life quest at a time — things to do away from work and school — and gives you XP for actually doing them. You tell it your situation ("two hours this evening, want to be outside"), a local open-weight model writes one specific quest, you go do it, and you earn XP and levels for showing up. Add a photo for bonus XP.

It's for anyone who finishes a workday and realises there's more to life than the screen but doesn't know where to start. The screen is the shortest part of the experience: about twenty seconds to get a mission, then you leave.

## Demo

<!-- TODO: deployed link or a short video demo -->

## Code

<!-- TODO: embed the GitHub repo link -->

## How I Built It

Go backend (stdlib `net/http`), SQLite via `sqlc`, and an Astro + Tailwind CSS + daisyUI frontend. The open-source AI is the core: quest generation runs through Ollama with the open-weight `llama3.2:3b` model, entirely on the user's machine.

## Why Does Open Innovation Matter?

Because the quests — and the personal context used to write them — never leave the laptop. A closed API would mean sending "I'm burnt out and have two free hours" to a vendor, paying per call, and needing internet in the middle of nowhere. An open-weight model running through Ollama costs nothing to run, works offline, and keeps everything local. And because the model is a swappable open-weight file, the app isn't locked to any provider — swap the model, or drop Ollama for raw llama.cpp, and the code doesn't care.

## My Agent Session

<!-- Optional, but judges love it. Save your session with DevRelay and embed it with the agent_session tag, or link to it. -->

## Prize Categories

<!-- Which partner categories are you entering? List every one that applies, or remove this section. -->
