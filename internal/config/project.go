// Package config holds the project settings, the player's saved settings
// and the game's tuning values (DESIGN §4, §10).
package config

// Title is the game's name, shown in the window title and on the title
// screen.
const Title = "Game Skeleton"

// TPS is the game loop rate. Game logic counts in ticks of 1/TPS seconds.
const TPS = 60

// Project is the preset this game uses: its resolution, scale modes and
// palette (DESIGN §4). This is the only place they are chosen.
var Project = Modern
