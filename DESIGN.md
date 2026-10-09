---
name: Sidequestar
description: One small real-life quest at a time, drawn as a topographic map sheet.
colors:
  map-paper: "oklch(98.4% 0.006 230)"
  map-paper-shade: "oklch(95.6% 0.01 228)"
  ink: "oklch(22% 0.012 120)"
  woodland: "oklch(45% 0.1 152)"
  cover-orange: "oklch(73% 0.17 55)"
  cover-ink: "oklch(20% 0.03 50)"
  contour: "oklch(60% 0.13 48)"
  water: "oklch(50% 0.13 245)"
  warning: "oklch(80% 0.15 85)"
  error: "oklch(54% 0.19 27)"
  night-paper: "oklch(21% 0.016 165)"
  night-paper-shade: "oklch(25% 0.018 165)"
  night-ink: "oklch(93% 0.012 120)"
  night-woodland: "oklch(74% 0.13 150)"
  night-woodland-ink: "oklch(19% 0.03 155)"
  night-cover-orange: "oklch(70% 0.16 55)"
  night-contour: "oklch(68% 0.12 55)"
  night-water: "oklch(74% 0.1 240)"
typography:
  display:
    fontFamily: "Barlow Semi Condensed, Barlow, ui-sans-serif, sans-serif"
    fontSize: "4.5rem"
    fontWeight: 700
    lineHeight: 1
  headline:
    fontFamily: "Barlow Semi Condensed, Barlow, ui-sans-serif, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 700
    lineHeight: 1.33
  title:
    fontFamily: "Barlow Semi Condensed, Barlow, ui-sans-serif, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 700
    lineHeight: 1.375
  body:
    fontFamily: "Barlow, ui-sans-serif, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.5
    fontFeature: "tnum"
  label:
    fontFamily: "Barlow, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 500
    lineHeight: 1.43
    fontFeature: "tnum"
rounded:
  field: "0.125rem"
  box: "0.25rem"
  symbol: "9999px"
spacing:
  xs: "0.5rem"
  sm: "0.75rem"
  md: "1.25rem"
  lg: "1.75rem"
  xl: "2.5rem"
  grid: "96px"
components:
  button-primary:
    backgroundColor: "{colors.woodland}"
    textColor: "{colors.map-paper}"
    rounded: "{rounded.field}"
  button-outline:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    rounded: "{rounded.field}"
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    rounded: "{rounded.field}"
  sheet-box:
    backgroundColor: "{colors.map-paper}"
    textColor: "{colors.ink}"
    rounded: "{rounded.box}"
    padding: "1.25rem"
  cover-band:
    backgroundColor: "{colors.cover-orange}"
    textColor: "{colors.cover-ink}"
    padding: "2rem 1.25rem"
  input:
    backgroundColor: "{colors.map-paper}"
    textColor: "{colors.ink}"
    rounded: "{rounded.field}"
  waypoint-marker:
    backgroundColor: "{colors.woodland}"
    textColor: "{colors.map-paper}"
    rounded: "{rounded.symbol}"
    size: "2.25rem"
  tag:
    textColor: "{colors.water}"
    typography: "{typography.label}"
---

# Design System: Sidequestar

## Overview

**Creative North Star: "The Evening Sheet"**

Every screen is a printed topographic map sheet: cool map-white paper ruled with a faint blue kilometre grid, black ink linework, woodland green for routes and go-actions, contour orange for relief, water blue for annotation, and an Explorer-orange cover band across the top. The person's evening is the territory; quests are pins, the active list is a dashed footpath joining numbered waypoints, and completed quests are inset photo frames.

The system is flat and drawn, not stacked. Containers are ink-ruled boxes with crisp near-square corners sitting directly on the paper; there is no card elevation. Density is moderate: one primary panel (the summon box on its contour field) and one route column, with secondary forms below. Two themes, day and night, follow `prefers-color-scheme`; night is the same map under a headlamp (deep green-black paper, pale ink), not an inverted palette.

Built on Tailwind CSS 4 + daisyUI 5 with two custom themes (`sheet-day`, `sheet-night`); all colour flows through daisyUI roles, so components use stock daisyUI classes re-skinned by the theme.

**Key Characteristics:**
- Cool map-white paper with a 96px faint blue grid on every page.
- 1.5px ink rules on every container, input, and section heading.
- Near-square corners (2px fields, 4px boxes); circles only for map symbols.
- Authored contour linework (public/contours.svg) as the one decorative field.
- Barlow Semi Condensed display over Barlow body, tabular numerals throughout.
- Values shown as map scale bars, not progress bars or percentages.

## Colors

A cartographic palette: cool paper, warm-black ink, and four map inks each holding one job.

### Primary
- **Woodland Green** (woodland / night-woodland): the go colour. Primary buttons (Generate, Accept), waypoint markers, the dashed footpath, the XP awarded line, caret colour, and the route heading icon.

### Secondary
- **Explorer Cover Orange** (cover-orange / night-cover-orange): the cover band behind the wordmark, the quest pin, text selection tint, and `theme-color`. Text on it is always Cover Ink (7.2:1).

### Tertiary
- **Contour Orange** (contour / night-contour): relief. Inks the contour field (mixed to 55% day, 40% night), the busy-state survey pulse, the photo-control outline, and the add-quest icon.
- **Water Blue** (water / night-water): annotation. Tags in italic, the 13% kilometre grid lines, the focus ring, and the profile icon.

### Neutral
- **Map Paper** (map-paper / night-paper): page and container background.
- **Paper Shade** (map-paper-shade / night-paper-shade): badge fill and scrollbar track.
- **Ink** (ink / night-ink): all text, all 1.5px rules, filled scale-bar blocks. Secondary text is ink at 70-80% opacity; placeholders at 60%.
- **Warning / Error**: full-bleed alert boxes with ink rule; success shares Woodland's hue and is used only for completion check icons.

### Named Rules
**The One Ink Per Job Rule.** Green means go or route, cover orange means the sheet's identity, contour orange means relief, blue means annotation. Do not swap jobs to add variety.

**The Paper Holds Rule.** Contour orange is a line and symbol colour. On day paper it reaches only 3.96:1, so it never carries small text; small accent-coloured labels use ink or night theme only.

## Typography

**Display Font:** Barlow Semi Condensed (600, 700; fallback Barlow, ui-sans-serif)
**Body Font:** Barlow (400, 400 italic, 500, 600; fallback ui-sans-serif, system-ui)

**Character:** The condensed grotesque of map titling over a plain, legible companion; together they read as printed sheet lettering. Numerals are tabular everywhere (`font-variant-numeric: tabular-nums` on `html`).

### Hierarchy
- **Display** (700, 4.5rem desktop / 3.75rem mobile, line-height 1): the level numeral in the sheet box; the wordmark uses the same face at 3.75rem / 2.25rem with tight tracking.
- **Headline** (700, 1.5rem; summon heading 1.875rem on sm+): section headings, always with a leading icon and a 1.5px ink rule beneath.
- **Title** (700, 1.25rem / 1.5rem in quest cards, line-height snug): quest titles; 600 at 1.5rem for the level title and 1.125rem for completed captions.
- **Body** (400, 1rem, max 65ch): quest descriptions at 80% ink.
- **Label** (500, 0.875rem): form labels (1rem 500), duration labels, XP meta. Tags are label size in italic Water Blue.

### Named Rules
**The Map Lettering Rule.** Headings, numerals that matter, and names of places-in-the-flow (quest titles, level title) are Semi Condensed; everything you read in a sentence is Barlow. No uppercase tracking labels.

## Layout

- Page container max 72rem (`max-w-6xl`), 1.25rem side padding.
- Cover band spans full width; inside it, wordmark and tagline left, the sheet box (20rem) bottom-aligned right from md up; stacked on phones.
- Main body is a 12-column split from lg (1024px): left 7 columns hold the summon panel, fresh quests, and completed insets; right 5 columns hold "Your route". Stacked on smaller screens, summon first.
- Secondary forms (add quest, profile) sit in a 2-column row below, separated by 3.5rem.
- Rhythm: 2.5rem between sections, 1rem between quest cards, 1.25rem container padding (1.75rem for the summon panel on sm+), 0.75rem between form fields.
- The 96px kilometre grid is a page background, not a layout grid; content does not snap to it.

## Elevation & Depth

Flat. daisyUI `--depth: 0` and `--noise: 0` in both themes; containers have no shadow and separate from the paper only by their 1.5px ink rule. Depth is cartographic: contour linework behind the summon panel and in empty inset frames, and the grid under everything.

### Named Rules
**The Printed Sheet Rule.** Nothing floats. Separation comes from ink rules and paper, never from box-shadow.

## Shapes

- Fields, buttons, badges: 2px corners (`--radius-field`, `--radius-selector`).
- Boxes (panels, cards, alerts, inset frames): 4px (`--radius-box`).
- Circles are reserved for map symbols: numbered waypoint markers.
- Borders are 1.5px ink (`--border: 1.5px`) on every container, field, and heading underline.
- Inset photo frames: square, 1.5px ink rule, 6px paper mat, image corners 1px.
- The footpath is a 2px dashed vertical line (6px dash, 4px gap) in Woodland.

## Components

### Buttons
- **Shape:** 2px corners, daisyUI sizing.
- **Primary:** Woodland fill, paper text; `btn-lg` for Generate, full width on phones. One primary per region.
- **Outline:** ink-ruled, transparent (Save, Add quest, Mark done in `btn-sm`).
- **Ghost:** Dismiss beside Accept.
- **Photo control:** small button with a Contour Orange outline and camera icon, ink label on paper (orange never carries the text); the native file input is visually hidden and its label carries the focus ring.
- **Busy:** button disables, label swaps to a spinner and "The quest-giver is thinking…".
- **Focus:** 2px Water Blue outline, 2px offset, globally.

### Tags
- **Style:** unboxed italic Water Blue list items, 0.875rem, spaced 0.75rem. In the profile they become ink-ruled badges on Paper Shade, still italic.

### Cards / Containers (sheet boxes)
- **Corner Style:** 4px.
- **Background:** Map Paper.
- **Shadow Strategy:** none (see Elevation).
- **Border:** 1.5px ink.
- **Internal Padding:** 1.25rem; 1.5rem for forms on sm+.
- **Quest card:** left padding 3.5rem for a Cover Orange map pin at top-left; title, description, duration scale bar, tags, actions.

### Inputs / Fields
- **Style:** daisyUI input/textarea/select, ink border, paper fill, 2px corners, placeholder at 60% ink, textareas non-resizable.
- **Labels:** above the field, 500 weight; "(optional)" in 400 at 70% ink.

### Navigation
- None beyond the wordmark link home; single-page app.

### Cover Band
Full-width Explorer Orange with a 1.5px ink bottom rule. Compass icon plus wordmark, tagline in Semi Condensed 600, one line of Barlow at 80%. Hosts the sheet box: level numeral spanning two rows, level title, quest/XP count, and a large scale bar with "x / y XP" and "Level n+1" beneath.

### Scale Bar
Map scale bar as the only value meter. Segments with 1.5px ink top/bottom and end borders; filled segments alternate ink and paper, unfilled are 25% ink outline. Small (6px x 12px blocks, 3 segments) for quest duration beside the bucket label; large (10px tall, 10 segments, full width) for level XP. Always carries an `aria-label` with the real values.

### Contour Field
Authored contour linework (`public/contours.svg`, generated vector) used as a CSS mask, inked in Contour Orange and faded from top-right toward bottom-left so text stays on clean paper. Used behind the summon panel and as the fill of completed insets without a photo. While generating, an overlay pulses outward from the summit (1.6s, expo-out, infinite); disabled under reduced motion.

### Route and Waypoints
Ordered list. Each waypoint is a 2.25rem Woodland circle with ink rule and a Semi Condensed numeral, joined to the next by the dashed footpath; content column to the right with title, description, duration, tags, and Mark done / photo actions.

### Completed Insets
2-column (3 on sm+) grid of square inset frames with caption: title in Semi Condensed 600 and "+n XP" in Woodland.

## Do's and Don'ts

### Do:
- **Do** rule every container, field, and section heading in 1.5px ink and keep containers flat.
- **Do** keep corners at 2px for controls and 4px for boxes; use circles only for map symbols.
- **Do** express quantities (duration, XP) as scale bars with an accessible label.
- **Do** draw relief with the authored contour artwork, faded away from text.
- **Do** define every colour through the two daisyUI themes so day and night stay in step.
- **Do** keep motion to the survey pulse and the 600ms pin drop, both behind `prefers-reduced-motion: no-preference`.

### Don't:
- **Don't** add box-shadows or lifted cards; the sheet is printed.
- **Don't** fake contours with CSS radial-gradient rings; use the vector artwork.
- **Don't** set small text in day Contour Orange (3.96:1 on paper).
- **Don't** add uppercase eyebrow labels or kickers above headings.
- **Don't** use progress bars, percentages, or rings for XP.
- **Don't** introduce rounded pill cards or large radii.
