*This is a submission for the [Hacktoberfest Open-Source AI Challenge Week 1: Touch Grass](https://dev.to/challenges/hacktoberfest-week1-2026-10-05)*

## What I Built

**The main quest can wait.** One small escape at a time.

You finish work, you know there has to be more to life than another evening on a screen, and you still don't know what to actually do with the next two hours. That gap is what Sidequestar is for.

You tell it your situation ("two hours this evening, want to be outside"), a model running on your own machine writes one specific quest, and you go do it. Finish it and you get XP and levels. Add a photo and you get a bit more. The screen part takes about twenty seconds. The rest happens outside.

It also gets to know you. Every time you finish a quest, the model rewrites a short summary of what you seem to enjoy and when you're free, and that summary shapes the next quest. I didn't want that to be a black box, so the profile page shows, word for word, what the model is told about you, and you can edit the summary if it gets something wrong. Everyone on the machine has their own account, so housemates get their own quests, XP and summaries.

<!-- TODO (Antony): bonus points for taking it outside. A few sentences: a quest it gave you, whether you did it, how it went. -->

## Demo

<!-- TODO: video or screenshots of the loop: summon a quest, accept it, complete it with a photo, then the profile page. -->

## Code

{% github tony-nyagah/sidequestar %}

## How I Built It

The backend is Go (standard library `net/http`) with SQLite through `sqlc`. The frontend is Astro with Tailwind CSS and daisyUI. Quest generation and the profile summaries both run through Ollama with `llama3.2:3b`, an open-weight model, entirely on the user's machine. `make dev` starts Ollama, the API and the frontend together.

The interesting part was getting a 3B model to behave. A few things I ran into:

- **It made things up, and then kept them.** Early summaries invented details like "free until 10am". The bigger problem was that each rewrite started from the previous summary, so once the model invented something, it treated it as fact from then on. The fix was to stop feeding it its own output. A rewrite now only sees what you wrote, the quests you actually finished, and your last manual edit.
- **It treated starter quests as history.** New accounts come with a few seed quests on the list, and the model read them as things the person had done ("she tends to finish reading and stretching quests" from someone who'd done neither). Labelling the list as "not done yet, not evidence" and dropping the temperature for summaries fixed most of it. It still slips occasionally, which is why the summary is editable.
- **Showing the prompt beats describing it.** The profile page gets its "what the model knows" lines from the same function that builds the prompt, so there's no way for the two to drift apart.
- **Errors should tell you what to do.** If Ollama isn't running, or the model isn't pulled yet, the app says so and gives you the command, instead of "something went wrong". With local models that's the error you'll hit most.

The design is a topographic map sheet: an orange map-cover header, contour lines behind the quest box, your active quests as a footpath between numbered waypoints, and finished ones as photo insets. Light and dark mode follow your system.

## Why Does Open Innovation Matter?

Because of what this app knows about you. Your situation, your history and a running summary of who you are never leave your laptop. With a closed API, "I'm burnt out and have two free hours" goes to a vendor, every quest costs money, and the whole thing stops working wherever there's no signal, which is exactly where you want to be if the quest worked.

Running an open-weight model through Ollama costs nothing, works offline and keeps everything local. It isn't tied to a provider either. Swap in another model, or replace Ollama with llama.cpp, and the code doesn't change.
