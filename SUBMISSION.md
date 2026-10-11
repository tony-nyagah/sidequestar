*This is a submission for the [Hacktoberfest Open-Source AI Challenge Week 1: Touch Grass](https://dev.to/challenges/hacktoberfest-week1-2026-10-05)*

## What I Built

**The main quest can wait.** One small escape at a time.

Sidequestar gives you one small thing to do away from work and school, and XP for actually doing it. You tell it your situation ("two hours this evening, want to be outside"), a model running on your own machine writes one specific quest, and you go do it. Finish it and you earn XP and levels. Add a photo and you earn a bit more.

It's for anyone who gets to the end of a workday with a nagging feeling that there has to be more to life than the screen, but no idea where to start. The screen is the short part: about twenty seconds to get a quest, and then you leave.

The quest-giver also gets to know you. Every time you finish a quest, the model rewrites a short summary of what you seem to enjoy and when you're free, and that summary shapes the next quest. You can check it: your profile page shows, word for word, what the model is told about you, and you can correct the summary yourself. Everyone on the machine has their own account, so housemates get their own quests, XP and summaries.

<!-- TODO (Antony): the challenge gives bonus points for taking it outside and saying how it went. A few sentences here: a quest it gave you, whether you did it, what happened. -->

## Demo

<!-- TODO: video or screenshots of the loop: summon a quest, accept it, complete it with a photo, then the profile page. -->

## Code

{% github tony-nyagah/sidequestar %}

## How I Built It

The backend is Go (the standard library's `net/http`) with SQLite through `sqlc`. The frontend is Astro with Tailwind CSS and daisyUI. Quest generation and the profile summaries both run through Ollama with the open-weight `llama3.2:3b` model, entirely on the user's machine, and `make dev` starts Ollama, the API and the frontend together.

Getting a 3B model to behave taught me a few things:

- Show the prompt instead of describing it. The profile page gets its lines from the same function that builds the prompt, so what you read there is exactly what the model gets.
- Don't feed a model its own guesses. Early summaries invented details ("free until 10am"), and because each rewrite started from the previous one, the inventions stuck. Now a rewrite only sees what you wrote, the quests you actually finished, and your last edit to the summary. It never sees its own earlier output.
- Starter quests aren't evidence. The model treated the seed quests on a new account's list as things the person had done. Labelling them as not done yet and lowering the temperature for summaries fixed most of that, and the summary stays editable for the rest.
- An error should say how to fix it. If Ollama isn't running or the model isn't pulled, the app tells you and gives you the command to run, instead of "something went wrong".

The design is a topographic map sheet: an orange map-cover header, contour lines behind the quest box, your active quests drawn as a footpath between numbered waypoints, and finished ones as photo insets. Light and dark mode follow your system setting.

## Why Does Open Innovation Matter?

Because the quests, the personal context behind them, and the running summary of who you are never leave the laptop. With a closed API, "I'm burnt out and have two free hours" would go to a vendor, every quest would cost money, and the app would stop working wherever there's no signal. An open-weight model running through Ollama costs nothing, works offline, and keeps everything local. It also isn't tied to any provider: swap in another model, or replace Ollama with llama.cpp, and the code doesn't change.
