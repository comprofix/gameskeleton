# CLAUDE.md

Guidance for Claude Code when working in this repo. This file is tracked in git
and comes with every game made from the skeleton.

## 1. Project

**Game Skeleton:** a starter repository for Go and Ebitengine retro and arcade
games. New games are created from it (it is a GitHub template repository), so
it must stay **game-agnostic**. The only game content is the small demo
("Dot Dash").

- Remote: `git@github.com:comprofix/gameskeleton.git`
- `DESIGN.md` and `PLAN.md` are local working documents, **not tracked in git**.

> During new game setup (§2), Claude replaces this section with the game's own
> description and fixed decisions, and removes §11.

**Current phase:** the skeleton is being built. DESIGN.md and PLAN.md describe
it, and PLAN.md lists milestones K0–K8.

## 2. New game setup

**At the start of a session, check whether `DESIGN.md` and `PLAN.md` exist.** If
either is missing:

1. **Ask:** "Is this a new game made from the skeleton, or work on the
   skeleton itself?"
   - **The skeleton itself:** ask whether to write fresh DESIGN.md and PLAN.md
     for that work, or carry on without them. Skip the rest of this section.
2. **For a new game, ask about it.** Group the questions; don't ask one at a
   time.
   - The name, a one-line summary, and the module path
     (`github.com/comprofix/<game>`).
   - A brand-new game, or a remake of an existing one (§9 applies to remakes).
   - The preset:
     - **Modern:** 640×360, Resurrect 64, `Chip4`.
     - **Classic PC:** 320×200, Sweetie 16, `Speaker`, optional CRT 4:3.
     - Or custom settings.
   - Theme, hero, controls, main rules, scoring, difficulty and any special
     mechanics.
   - What is not included in v1.
3. **Summarise the answers and wait for the go-ahead.** Then:
   1. Run `go run ./tools/newgame -module … -name … -title "…"`. Add
      `-keep-demo` only if asked.
   2. Set the title and the preset in `internal/config/project.go`: `Modern`,
      `Classic`, or a custom `Preset` added to `internal/config/preset.go`.
   3. Write `DESIGN.md` and `PLAN.md` using the outlines below. Mark anything
      not yet decided as **(open)** in the decisions table.
   4. Change §1 of this file to describe the game, including its fixed
      decisions. Change the names in §5's commands. Remove §11.
   5. Run `go vet ./...` and `go test ./...`.
4. **Hand back:** show the documents, list the open decisions, and stop. Do not
   write game code or commit until the user says so.

### DESIGN.md outline

```
# DESIGN — <Game>
<what the game is; for a remake, what it is faithful to>
> Status line
1. Pillars              3–5 guiding principles
2. Theme and setting    world, hero, name
3. Look and sound       resolution, palette, art style, screen layout mock-up, audio
4. Controls             key/gamepad table; reserved keys
5. Rules                movement, hazards, goals, timer, lives, scoring table, difficulty and skills
6. Special mechanics    each with a short description
7. Screens and flow     scene diagram; title, menus, game over, high scores
8. Options and accessibility
9. Not in v1
10. Decisions           table: # | Decision | Status (Decided: … / (open) + recommendation)
```

### PLAN.md outline

```
# PLAN — <Game>
> Status line
1. Technology           table (from the skeleton, plus anything the game adds)
2. Code layout          the skeleton packages plus the game's own (e.g. world/, fx/)
3. Technical approach   game logic, collision, events, levels as data, randomness
4. Testing              unit tests for game logic, golden replays, manual checklists
5. Milestones           M0…Mn: "- [ ]" tasks, then "Done when:"; each one playable
6. Risks                table: Risk | What we do about it
7. Discuss later
```

## 3. Phases and decisions

- **Design phase:** only DESIGN.md, PLAN.md and CLAUDE.md change. Do not write
  code, install anything or create build files until the user explicitly says
  to start a milestone.
- **Build phase:** one PLAN.md milestone at a time. Each milestone ends with a
  playable build and passing tests. Tick its checkboxes when it's done.
- **Decisions** live in the DESIGN.md decisions table. Never pick an answer to
  an open decision yourself; bring the question to the user. Once it's decided,
  update DESIGN.md and PLAN.md so they stay the single source of truth.
- Rules or facts that aren't confirmed yet are tagged **[verify]** in
  DESIGN.md. Flag them rather than guessing.

## 4. Working with the user

- **Confirm before** creating, changing or deleting files, installing modules,
  or making commits or pushes. Reading and inspecting are fine. Planning and
  discussion are not a go-ahead.
- **Never commit or push** unless the user says so for that specific commit or
  push.
- **One commit per milestone.** When a milestone's checks pass, report back
  and wait. Once the user approves, commit it on its own as
  `<milestone>: <summary>` (e.g. `K1: …` in the skeleton, `M3: …` in a game).
- Commits have no `Co-Authored-By: Claude` or `Claude-Session:` trailers. The
  user is the only author.
- **No CI/CD.** Do not add GitHub workflows unless the user explicitly asks.
- **Do not track** DESIGN.md, PLAN.md or `.claude/` in git, and do not remove
  them from `.gitignore`.
- Report results truthfully: if tests fail, show the output; if a step was
  skipped, say so.

## 5. Stack and commands

- Go 1.27 and Ebitengine v2.10.5.
- Settings and scores are saved as `settings.json` and `scores.json` next to
  the game.

```sh
go run ./cmd/gameskeleton                 # run the game
go run -tags debug ./cmd/gameskeleton     # with the debug overlay
go run ./cmd/gameskeleton -seed 42        # with a fixed random seed
go generate ./assets                      # rebuild sprite sheets from assets/sprites/*.txt
go run ./tools/spritegen -in assets/sprites -out assets/gen -preview <dir>
go build -o bin/ ./cmd/gameskeleton
go vet ./...
go test ./...
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/gameskeleton.exe ./cmd/gameskeleton
```

Linux builds need cgo and the X11/GL/ALSA development packages:
`libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev
libxxf86vm-dev libasound2-dev pkg-config`. Build them in WSL.

## 6. Code conventions

- Standard Go style: `gofmt`, small packages under `internal/`, no global
  variables that change.
- **Tick-based (60 ticks per second) and deterministic.**
  - No `time.Now()` in gameplay.
  - All randomness comes from the RNG in `Context`, created from the seed.
  - Use whole-number or fixed-point positions where exact results matter.
- `Update` changes state; `Draw` only reads it. Never change state in `Draw`.
- **Game logic runs headless:** no GPU calls in `Update`. Game-logic packages
  must not import graphics or audio packages. Scenes and audio react to game
  state and events.
- **Project settings** are chosen only in `internal/config/project.go`:
  - the title and tick rate;
  - `Project`, the preset, which sets the resolution, scale modes and
    palette.

  Presets are defined in `internal/config/preset.go`. Code reads them through
  `config.Project`. Never hard-code a screen size or palette anywhere else.
- Colours are named with the constants of the chosen palette's package (e.g.
  `resurrect64.Gold`) and turned into RGBA with `gfx.RGBA`.
- Tuning values (speeds, timings, scores, levels, skill tables) live in
  `internal/config`. No magic numbers in game code.
- Colours come from the palette, by name. No hex codes outside the palette
  files, including in sprite files.
- Assets are built into the game with `go:embed` from `assets/`.
- Code comments cite DESIGN sections (e.g. "DESIGN §5") to explain why the code
  does something.
- Alt+Enter, F9, F10 and F11 are reserved for the window and never trigger
  game actions.

### Package map

```
cmd/<game>/        main.go: flags, saved files, window, run
internal/
  game/            ebiten.Game: scenes, window keys, scale modes, presentation
  scene/           scene manager, fades, shake, shared scenes
  ui/              list menu, panel, text helpers
  config/          project.go, settings, tuning values, skills, levels
  session/         one game in progress
  score/           high score tables
  gfx/             palette, schemes, sheets, animation, fonts, scaler
  gfx/palettes/    ready-made palettes
  audio/           synth, tunes, sound effects, scheduling, output
  input/           actions, keyboard, gamepad, scripted input
  spritegrid/      text-grid sprite parser and packer
  debug/           debug.Enabled (build tag: debug)
  <game logic>/    the game's own world, with no graphics or audio imports
tools/spritegen/   sprites → PNG + JSON (+ preview)
assets/            assets.go, sprites/, gen/, font/
```

## 7. Art

- Claude makes all the art as **text-grid sprites** in `assets/sprites/*.txt`
  (format in the `spritegrid` package docs):
  - `.` is a see-through pixel.
  - With palettes of up to 16 colours, each pixel can be a hex digit.
  - Larger palettes use a `colours` key in each file that names the palette
    colours.
  - `recolour` makes colour-swapped copies.
- Facing left is made by mirroring when drawn. Do not draw separate
  left-facing art.
- After changing a sprite, run `go generate ./assets`. The regenerated
  `assets/gen/*` is part of the change.
- Check new art by eye on the preview sheet before using it. Previews go in the
  scratchpad or `/preview/`, never in `assets/gen`.
- Effects (lighting, particles, weather) are drawn by code at runtime, not
  stored as sprites.
- The game only loads PNG and JSON, so a sprite drawn by hand can replace a
  generated one without code changes.

## 8. Audio

- All sound is made in code by the synth in `internal/audio`. No audio files
  unless the user agrees.
- Tunes and sound effects are note tables in Go. A nil `*audio.Synth` ignores
  every call, so tests run without sound.

## 9. Originality and reference material

- All art, music, sound and text is original. Never add or reference another
  game's assets: no ripped sprites, no transcribed music, no copied text.
- **For a brand-new game "in the style of" a classic:** do not use the
  original's name, characters, art, music or level layouts.
- **For a remake, follow the facts-only rule.** Disassemblies, source code or
  documents about the original may be read **for facts only**: rules,
  formulas, counts, thresholds and difficulty tables.
  - Record each fact in DESIGN.md, tagged **[ref]**, before using it.
  - Never copy or translate its code, sprites, fonts, sound tables or text.
  - Write Go code from DESIGN.md, not from the reference.
  - Clone reference repos into the scratchpad, never into this repo.
- Reference videos and screenshots go in `ref/`, which is git-ignored.
  - Extract frames into the scratchpad.
  - Use them for layout, behaviour and timing only. Never trace sprites from
    them.

## 10. Testing

- Every change to game logic comes with unit tests that run without a window
  or sound.
- Prefer golden replay tests (seed plus inputs → expected final state) for
  gameplay sequences.
- Run `go vet ./...` and `go test ./...` before saying work is done.
- Visual and audio work: check by hand against a short checklist at each
  milestone, and offer screenshots for the user to review.

## 11. Rules for the skeleton itself

Removed during new game setup.

- **Stay game-agnostic.** Only add what a typical retro arcade game needs.
  Anything else goes in PLAN.md "Discuss later".
- **Both presets must work:** Modern and Classic PC. Test both whenever the
  scaler, palettes or audio change.
- **Keep the demo separate.** Demo code lives only in `internal/demo` and
  `assets/sprites/demo.txt`. `tools/newgame` must be able to remove it
  cleanly.
- **Keep new game setup working.** Any change to module paths, folders or §2 of
  this file must keep the `tools/newgame` test passing.
