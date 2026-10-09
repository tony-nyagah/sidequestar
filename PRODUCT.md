# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

One person running Sidequestar on their own machine. They've finished work or school, know there's more to life than the screen, and don't know where to start. They open the app to get one concrete thing to go do, then leave. Use is split across devices: quests are summoned on a laptop as often as on a phone, and completion (with a photo) frequently happens on a phone, out in the world.

## Product Purpose

Sidequestar hands out one small, specific real-life quest at a time and rewards the person with XP for actually doing it. Success is the person away from the screen: about twenty seconds in the app to get a quest, then the real activity, then a quick return to mark it done.

## Positioning

The quest-giver is a local open-weight model (Ollama, `llama3.2:3b` by default) running on the person's own machine. The situation they describe and their quest history never leave it, it works offline, and costs nothing per quest. Generation is the product: without the local model there are no quests.

## Operating Context

- Built for the Hacktoberfest Open-Source AI Challenge, Week 1: "Touch Grass". The UI will appear as static screenshots in the dev.to submission post.
- Runs as a Go API plus an Astro SSR frontend on localhost; no accounts, no hosted backend (ADR 0001).
- Generation takes several seconds and around 30 seconds on a cold model load, so waiting on the quest-giver is a real part of the flow.

## Capabilities and Constraints

- Quest lifecycle: generated (proposed) → active (on the list) → completed. Quests can also be added by hand; a few seed quests ship with the app.
- Completion awards XP by duration bucket plus a completion bonus, plus a photo bonus. XP rolls up into levels with cosmetic titles ("Grass Toucher" through "Legend").
- Profile is optional location and a free-text note; interests are derived from quest tags.
- Terminology is fixed by GLOSSARY.md (quest, generation, completion, XP, level, title; avoid "task", "points", "rank").
- Stack: Tailwind CSS 4 + daisyUI 5 must stay; work is component-level, not a custom component system.
- Light and dark themes are both required.
- Deadline is tight: a finished, polished result beats an ambitious partial one.

## Evidence on Hand

- Real quests from seed data and live generation; no testimonials, user counts, or metrics exist, and none should be invented.
- Photos are the user's own completion photos, stored locally.

## Product Principles

1. The screen is the shortest part of the experience. Every view should get the person to a quest and out the door.
2. One quest at a time. Specific and doable today beats a menu of options.
3. Reward showing up, not optimising. XP celebrates doing the thing; it is not a productivity score.
4. Everything stays on the person's machine, and the product can say so plainly.
