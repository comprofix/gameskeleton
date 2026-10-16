// Command gameskeleton runs the skeleton's demo game (DESIGN §3).
package main

import (
	"flag"
	"log"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/gameskeleton/internal/config"
	"github.com/comprofix/gameskeleton/internal/game"
)

func main() {
	seed := flag.Uint64("seed", 0, "random seed, for repeatable runs (0 picks one from the clock)")
	flag.Parse()
	if *seed == 0 {
		*seed = uint64(time.Now().UnixNano())
	}

	// The settings live beside the game (DESIGN §10).
	p := config.Project
	dir := config.SaveDir()
	saved := game.Saved{
		Settings: config.LoadSettings(filepath.Join(dir, config.SettingsFile), p),
		Dir:      dir,
	}

	ebiten.SetWindowTitle(config.Title)
	ebiten.SetWindowSize(saved.Settings.WindowW, saved.Settings.WindowH)
	minW, minH := p.MinWindow()
	ebiten.SetWindowSizeLimits(minW, minH, config.MaxWindowW, config.MaxWindowH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(saved.Settings.Fullscreen)
	ebiten.SetWindowClosingHandled(true) // to save the window size on close
	ebiten.SetTPS(config.TPS)

	g, err := game.New(*seed, saved)
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
