# Game Skeleton

A starter template for retro and arcade games written in Go with
[Ebitengine](https://ebitengine.org/).

> **Work in progress.** The skeleton is being built and isn't ready to use yet.

## What it will provide

- A window drawn at a fixed internal resolution and scaled up by whole numbers,
  with fullscreen and an optional CRT 4:3 mode.
- Ready-made palettes (Resurrect 64, Sweetie 16) and text-grid sprites packed
  into sprite sheets by `go generate`.
- Scenes with fades, menus, options, pause and high score tables.
- Keyboard and gamepad input.
- Sound made in code: a 4-channel chip synth or a single PC-speaker voice.
- Saved settings and scores, a debug overlay, and tests that replay games
  exactly.
- A tiny demo game, plus a `CLAUDE.md` that walks Claude Code through setting
  up a new game.

## Requirements

- Go 1.27 or later
- **Linux only:** cgo plus the X11/GL/ALSA development packages:
  `libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config`

## Build

```sh
go build ./...
go test ./...
```

## Personal documents

`DESIGN.md` and `PLAN.md` are personal working documents. `.gitignore` keeps
them out of git, so back them up somewhere else. `CLAUDE.md` is tracked.

## License

MIT. See [LICENSE](LICENSE).
