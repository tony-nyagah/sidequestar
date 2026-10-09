# Sidequestar

Sidequestar is a local-first, single-user app that hands you small real-life quests — things to do away from work and school — and rewards you with XP for doing them. A local open-weight model is the quest-giver; everything else lives on your own machine.

## Language

**Quest**:
A single, concrete real-life activity a person can start today, with a title, description, an expected duration, and tags.
_Avoid_: task, todo, item, activity, goal

**Duration bucket**:
One of five fixed time ranges a quest is expected to take (under 30 minutes, 30–60 minutes, 1–2 hours, half a day, full day). Used to award base XP.
_Avoid_: estimate, effort

**Source**:
Where a quest came from: `generated` (written by the local model), `hand` (added by the user), or `seed` (shipped with the app).
_Avoid_: origin, author, creator

**Status**:
The stage of a quest's life: `generated` (proposed, not yet accepted), `active` (on the user's list), `completed` (done).
_Avoid_: state, phase

**Generation**:
The act of asking the local model to invent a single quest from the user's situation. The core of the app.
_Avoid_: suggestion, recommendation

**Completion**:
Marking an active quest as done, optionally with a photo. Awards XP.
_Avoid_: finish, close

**XP**:
Experience points earned for completing a quest: a duration-based amount plus a flat completion bonus plus an optional photo bonus.
_Avoid_: points, score, coins

**Level**:
A rank derived from total XP. Each level needs `100 × level` XP more than the last.
_Avoid_: rank, tier, grade

**Title**:
The name given to a level (for example "Trailblazer"). Purely cosmetic.
_Avoid_: rank name, badge

**Profile**:
A single row of optional facts about the user — location and a free-text note — fed to the model to shape generation.
_Avoid_: account, user settings

**Interests**:
The most frequent tags across the user's completed quests, computed on read. Used to shape generation. Generated and seed quests don't count, so the model's own tags can't feed back into its prompt.
_Avoid_: preferences, hobbies

**Seed quest**:
A starter quest shipped with the app so it is usable before any generation.
_Avoid_: default quest, template
