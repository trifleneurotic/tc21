# Tombstone City: 21st Century — Recreation Design Document

**Platform:** Go / Oak game engine (oakmound)
**Target completion:** October 9
**Original:** Texas Instruments, TI-99/4A, 1981 (designer John Plaster)

_This design document was compiled through a collaborative process with Claude (Anthropic), combining web research on the original game with the author's firsthand knowledge of its mechanics. Code implementation is entirely human-written._
---

## 1. Overview & Scope

A recreation of the TI-99/4A multidirectional shooter *Tombstone City: 21st Century*. The player controls a wagon ("schooner") defending a 4×4 safe zone, shooting alien "Morgs" and tumbleweeds to raise the town's "Population" (score). Single screen, no scrolling, sprite-based 2D.

**Out of scope for v1:**
- Skill-select menu (Novice / Insane difficulty tiers) — Master is hardcoded as the only playable difficulty
- Pause, restart-in-place, and quit meta-controls
- Persistent save between application launches (session-best score only)

---

## 2. Screen Layout & Coordinate System

- **Resolution:** 640×480
- **Coordinate origin:** top-left, (0,0)
- **Base grid unit:** 32×32 px tile
- **HUD:** bottom strip, 64px (2 tiles) — shows Day, Population, Schooners remaining, session-best Population
- **Playfield:** 640×416 (20×13 tiles), extends edge-to-edge, hard-wall boundary, no border margin
- **Placement model:** all entities (Building, Cactus, Morg, Tumbleweed) snap to the 32px grid; "touching/adjacent" is defined as grid adjacency, not pixel distance
- **Visual convention:** sprites are drawn with a small inset within their occupied cell for visual breathing room; this is cosmetic only — grid-position math always treats an entity as occupying the full cell

### Safe Zone Geometry
- 16 Building tiles arranged in a solid 4×4 block, each 32×32
- One tile-width gap (path) between every adjacent pair of buildings — **no outer buffer ring**; buildings sit flush with the exterior field
- The schooner roams freely within the full connected lattice of internal paths
- **12 Exit points**: the gap cells where an internal path lane (3 per side) reaches the outer boundary of the building block
- Schooner's default spawn point: dead center of the grid

---

## 3. Entity List

### Schooner
- **Visual:** small ship, two wings, single nose turret. One sprite asset, rendered at 4 rotations based on tracked facing direction (no separate directional art)
- **Size:** 32×32 (1 tile)
- **Position/spawn rule:** spawns at grid center at level start; **normal respawn** — returns to grid center when hit by a Morg, if exits aren't all currently blocked; **blockade ejection** — if all 12 exits become blocked while the schooner is inside, it is immediately teleported to a random position outside the grid (relocation only, no life lost, unless subsequently hit by a Morg while stranded)
- **Movement:** up / down / left / right, arrow keys; can fire while moving
- **State:** inside/outside the 4×4 grid; current facing direction
- **Collision behavior:** stopped by Cactus and Building; pushes Tumbleweed; destroyed on Morg contact → respawns per rule above, one life lost
- **Lifecycle:** created at level start; never permanently destroyed, only respawns with a life deducted, until lives reach zero

### Morg
- **Visual:** oblong green oval, two legs, 2-frame walk animation that alternates once per tile moved (tied to movement, not a timer)
- **Size:** 32×32 (1 tile)
- **Position/spawn rule:** from a randomly touching Cactus pair; from a pair destroyed via a kill adjacent to it; automatically every ~10 seconds regardless of Cactus state; immediately if the player fires while fewer than 10 Tumbleweeds and zero Morgs remain
- **Movement:** continuously seeks the Schooner; can never enter the 4×4 safe-zone interior (exit gap cells count as "not interior" for this purpose, so a Morg can sit adjacent to/at a gap without violating this rule)
- **State:** touching-exit (for the exit-seal check)
- **Collision behavior:** stopped by Cactus and Building; pushes Tumbleweed; on missile kill while adjacent to an exit gap, converts to Cactus there and permanently blocks that exit for the rest of the day; on contact with Schooner, Schooner respawns per its rule, one life lost
- **Lifecycle:** created per spawn rules above; destroyed (converts to Cactus) on missile hit

### Cactus
- **Visual:** vertical line on a base, two arms at differing heights
- **Size:** 32×32 (1 tile)
- **Position/spawn rule:** random at level start; also appears wherever a Morg is shot
- **Movement:** stationary
- **State:** unpaired / paired / adjacent-to-exit-gap / **about-to-spawn** (white background highlight — active when part of a generating pair about to produce a Morg; synced with the Building pending-spawn flash below)
- **Collision behavior:** blocks Schooner and Morg movement; blocks Missiles (absorbed, no effect)
- **Lifecycle** — three cases on missile kill of a Morg:
  1. Morg shot away from any cactus → becomes a lone Cactus
  2. Morg shot adjacent to an active pair → pair destroyed; spawns 1 new Morg (Master/Novice) or 2 new Morgs (Insane, deferred)
  3. Morg shot within one space of a lone Cactus → creates a new adjacent pair
  - A Cactus occupying an exit gap behaves identically to any other Cactus, including remaining eligible for pair-adjacency — it is not a special case

### Tumbleweed
- **Visual:** pointy circle, possibly purple
- **Size:** 16×16 (¼ tile)
- **Position/spawn rule:** batch of 20 spawns randomly at level start; a fresh batch of 20 spawns once the current batch is fully destroyed
- **Movement:** random cardinal hop every few seconds, max distance = own length/width
- **Collision behavior:** disintegrates on Missile hit; pushed by Schooner or Morg contact
- **Lifecycle:** created in batches of 20, at level start and on full depletion

### Missile
- **Visual:** a line
- **Size:** 10×2 px (horizontal, when Schooner faces left/right) or 2×10 px (vertical, when facing up/down) — orientation fixed at spawn based on facing at fire time, does not change mid-flight
- **Position/spawn rule:** spawns at the Schooner's nose/turret on Space press
- **Movement:** travels in the direction the Schooner was facing when fired
- **Collision behavior:** Tumbleweed → destroys it; Morg → converts to Cactus (per Cactus/Morg rules); lone Cactus → blocked, no effect; Building → blocked, no effect
- **Lifecycle:** one missile on screen at a time; despawns on any collision or on reaching the screen edge

### Explosion
- **Visual:** red spiky circle, no animation frames
- **Size:** 16×16 (tumbleweed size)
- **Position/spawn rule:** appears at the target's location when a Tumbleweed is destroyed or a Morg is killed (including kills that trigger a pair-destruction)
- **Movement:** stationary
- **Collision behavior:** none
- **Lifecycle:** created on Tumbleweed-kill or Morg-kill only (not on "blocked, no effect" hits); disappears after ~1–2 seconds

### Building
- **Visual:** blue square
- **Size:** 32×32 (placeholder, tunable)
- **Position/spawn rule:** static; 16 tiles in a 4×4 arrangement, one-tile-width internal gaps only (see Safe Zone Geometry)
- **Movement:** stationary
- **State:** normal / pending-spawn — **all 16 tiles** change color together whenever any Cactus pair is about to spawn a Morg (synced with that pair's about-to-spawn highlight)
- **Collision behavior:** blocks Missile, Tumbleweed, Morg, Schooner
- **Lifecycle:** created at round start; never destroyed or transformed

### Exit
- **Visual:** none — an open path by default; visually apparent only when a Cactus occupies and blocks it
- **Size:** same footprint as one Building tile
- **Position/spawn rule:** fixed — the 12 gap cells where an internal path lane meets the safe zone's outer boundary
- **Movement:** stationary
- **State:** open / blocked
- **Collision behavior:** passable while open; while blocked, governed entirely by normal Cactus rules
- **Lifecycle:** all start open at round start; a given Exit becomes blocked when a Morg dies adjacent to it and converts to a Cactus there; resets to open at the start of a new day
- *Kept as a distinct entity (rather than folded into Cactus) for extensibility in future iterations, even though its behavior is currently identical to a plain Cactus.*

---

## 4. Core Gameplay Loop

The schooner defends its safe zone, venturing out through the 12 exits to shoot Morgs and tumbleweeds for points. Morgs continuously spawn from adjacent cactus pairs (visually telegraphed by a synced white Cactus highlight + all-Building color flash) and hunt the schooner, which can never fully hide since Morgs auto-spawn every ~10 seconds regardless of player action. Careless shooting near an active pair extends the threat by spawning a fresh Morg instead of clearing it; careful shooting collapses pairs down to lone cacti or clears them outright. Killing Morgs at exit points seals them permanently for the day — advantageous for defense, but risky, since sealing all 12 exits at once ejects the schooner into the open field. A day ends when every generating pair is cleared, awarding a score bonus and an extra life before a harder day begins.

---

## 5. Controls & Input Mapping

| Action | Key |
|---|---|
| Move up / down / left / right | Arrow keys |
| Fire | Space bar |
| Retreat to safe zone | R |
| Pause / skill-select / restart / quit | Deferred — not in v1 |

- Movement and firing are independent, polled every frame — schooner can fire while moving
- Retreat costs 1,000 points (see Scoring)

---

## 6. Difficulty & Progression

- **v1 ships with a single hardcoded difficulty: Master** (Schooner speed = Morg speed; a Morg killed adjacent to an active Cactus pair spawns exactly 1 new Morg)
- Novice (2× Schooner speed) and Insane (2 Morgs per pair-kill) are designed but deferred to a post-menu iteration
- **Day progression:** Day 1 starts with 2 generating Cactus pairs; each subsequent day adds 1 pair, capped at 30

---

## 7. Scoring & Win/Lose Conditions

- **Population (score):** +150 per Morg killed · +100 per Tumbleweed destroyed · +1,000 bonus on clearing a day · −1,000 for using Retreat
- **Lives:** start with 10 Schooners; clearing a day awards +1 bonus Schooner; game ends at 0 Schooners
- **Win condition:** none — this is a score-attack/survival loop, not a game with a final victory state
- **Session-best Population:** tracked and displayed on both Title and Game Over screens; updates when beaten, with a visually distinct "New Best!" indicator on Game Over; persists across playthroughs within a session only — not saved between app launches

---

## 8. Game States & Flow

- **Title Screen** — displays game name and session-best Population; Space starts a new game
- **Gameplay** — the core loop above
- **Day Complete** — shows day-clear bonus (+1,000, +1 Schooner); Space advances to the next day (pairs +1, capped at 30)
- **Game Over** — triggered on 10th Schooner lost; shows final Population; updates and highlights session-best if beaten; any key returns to Title

**Reset note:** returning to Title and starting a new game resets everything to initial state — Schooners to 10, Population to 0, Day to 1 with 2 Cactus pairs, all Exits reopened — except the session-best score, which is the one value that survives a reset.

---

## 9. Audio

**v1 sound effects** (3 unique clips, matching the original's minimal audio design):
- Missile fire
- Morg killed → converts to Cactus
- Tumbleweed destroyed (shares the same clip as Morg killed)
- Schooner destroyed (hit by Morg)

**Music:**
- Title screen tune — tentative, contingent on locating a usable copy of the original TI-99/4A theme; ships silent if unavailable

**Deferred — potential future sound effects** (not present in the original; would be new additions, not restorations, if implemented later):
- Cactus-pair destruction (chain event)
- Exit sealed
- Retreat action
- Day cleared jingle
- Game Over jingle
- New session-best fanfare
- Looping background music during gameplay
- Audio cue on Building pending-spawn flash

---

## 10. Asset List

**Sprites**

| Asset | Size | Variants |
|---|---|---|
| Schooner | 32×32 (inset) | 1 sprite, rendered at 4 rotations via tracked facing |
| Morg | 32×32 (inset) | 2 frames, alternate per tile moved |
| Cactus | 32×32 (inset) | 2 states: normal, about-to-spawn (white highlight) |
| Tumbleweed | 16×16 | 1, static |
| Missile | 10×2 / 2×10 | 2 orientations |
| Explosion | 16×16 | 1, static |
| Building | 32×32 | 2 color states: normal, pending-spawn |
| Exit | — | None (no sprite of its own) |

**HUD / UI**
- Day counter · Population · Schooners remaining · Session-best Population (Title + Game Over) · "New Best!" indicator
- Title screen: game title art/text, "Press Space to Start" prompt
- Day Complete screen: bonus text
- Game Over screen: final score display

**Audio files**
- `missile_fire` (sfx)
- `kill` (sfx — shared by Morg killed and Tumbleweed destroyed)
- `schooner_destroyed` (sfx)
- Title theme (music, tentative)

---

## 11. Fidelity Notes

**Deliberate design additions (not confirmed in the original):**
- Schooner pushes Tumbleweeds on contact
- Session-best score tracking, Title/Game Over display, and "New Best!" indicator
- All deferred sound effects (Cactus-pair destruction, Exit sealed, Retreat, Day Cleared, Game Over, New Best fanfare, looping gameplay music, Building flash cue) — genuine new additions if implemented later, not restorations of original content
- Retreat remapped to **R** (original used Space, which is reassigned here to Fire)

**Judgment calls made where source documentation had gaps:**
- The 12-exit, solid-4×4-grid-with-no-outer-buffer geometry — reconstructed from memory of gameplay feel, cross-checked against the documented 16-building count, not confirmed against original source code or a manual diagram
- Day progression's starting pair count (2) and per-day increment (+1, cap 30) — source confirmed only the cap
- Missile-vs-lone-Cactus and Missile-vs-Building behavior (both "blocked, no effect") — plausible, not documented
- Explosion effect extended to Morg kills, not just Tumbleweed kills as literally described in source material

**Scope limitations for v1 (not deviations — planned future work):**
- Only Master difficulty is playable; Novice and Insane are designed but not wired to a menu
- No skill-select, pause, or quit meta-controls
- No persistent save between app launches — session-best only

**Matches the original as documented:**
- Core scoring values (150 / 100 / 1,000 / −1,000), 10 starting lives, Morg 10-second auto-spawn timer, cactus-pair spawn mechanics including the Insane double-spawn rule, all three original sound effect triggers, and the Cactus/Morg/Cactus lifecycle
