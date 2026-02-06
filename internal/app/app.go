package app

import (
	"fmt"
	"image/color"
	"time"

	"overlay/internal/state"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ViewWidth  = 520
	ViewHeight = 360
)

const maxOverlayLogs = 10

type App struct {
	state        *state.GameState
	dragging     bool
	dragStartX   int
	dragStartY   int
	windowStartX int
	windowStartY int
	scale        float64
	clicking     bool
	clickStartX  int
	clickStartY  int
	selectedIdx  int
}

func New(state *state.GameState) *App {
	return &App{state: state, scale: 1.0}
}

func (a *App) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		locked := a.state.ToggleLocked()
		ebiten.SetWindowMousePassthrough(locked)
		if locked {
			a.dragging = false
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		a.setScale(a.scale + 0.1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		a.setScale(a.scale - 0.1)
	}

	if !a.state.IsLocked() {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			a.dragging = true
			a.dragStartX, a.dragStartY = ebiten.CursorPosition()
			a.windowStartX, a.windowStartY = ebiten.WindowPosition()
			a.clicking = true
			a.clickStartX, a.clickStartY = a.dragStartX, a.dragStartY
		}

		if a.dragging {
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
				curX, curY := ebiten.CursorPosition()
				dx := curX - a.dragStartX
				dy := curY - a.dragStartY
				if abs(dx)+abs(dy) > 6 {
					a.clicking = false
				}
				ebiten.SetWindowPosition(a.windowStartX+dx, a.windowStartY+dy)
			} else {
				a.dragging = false
				if a.clicking {
					a.selectNextHero()
				}
				a.clicking = false
			}
		}
	}

	return nil
}

func (a *App) setScale(scale float64) {
	if scale < 0.7 {
		scale = 0.7
	}
	if scale > 1.6 {
		scale = 1.6
	}
	if scale == a.scale {
		return
	}
	a.scale = scale
	w := int(float64(ViewWidth) * scale)
	h := int(float64(ViewHeight) * scale)
	ebiten.SetWindowSize(w, h)
}

func (a *App) Draw(screen *ebiten.Image) {
	clr := color.RGBA{10, 10, 20, 220}
	if !a.state.IsLocked() {
		clr.R = 50
	}
	ebitenutil.DrawRect(screen, 0, 0, ViewWidth, ViewHeight, clr)

	snap := a.state.Snapshot(maxOverlayLogs)

	statusMsg := fmt.Sprintf("Status: %s | F12: Lock | +/- Scale (%.1fx)\n%s\n\nENEMIES:", snap.Status, a.scale, gsiLine(snap))
	for _, id := range snap.EnemyHeroesIDs {
		statusMsg += fmt.Sprintf(" [%s]", snap.HeroIDToName[id])
	}

	statusMsg += "\n\n" + buildGSIPanel(snap)
	statusMsg += "\n\n" + buildCounterTable(snap, a.selectedHeroID(snap))
	statusMsg += "\n\n" + buildBestPicksTable(snap)

	ebitenutil.DebugPrint(screen, statusMsg)
}

func (a *App) Layout(w, h int) (int, int) {
	return ViewWidth, ViewHeight
}

func (a *App) selectedHeroID(snap state.Snapshot) int {
	if len(snap.EnemyHeroesIDs) == 0 {
		return snap.LastCounterHero
	}
	if a.selectedIdx < 0 || a.selectedIdx >= len(snap.EnemyHeroesIDs) {
		a.selectedIdx = len(snap.EnemyHeroesIDs) - 1
	}
	if a.selectedIdx < 0 {
		return snap.LastCounterHero
	}
	return snap.EnemyHeroesIDs[a.selectedIdx]
}

func (a *App) selectNextHero() {
	snap := a.state.Snapshot(0)
	if len(snap.EnemyHeroesIDs) == 0 {
		return
	}
	a.selectedIdx = (a.selectedIdx + 1) % len(snap.EnemyHeroesIDs)
}

func buildCounterTable(snap state.Snapshot, heroID int) string {
	name := snap.HeroIDToName[heroID]
	if name == "" {
		if heroID == 0 {
			name = "Waiting for pick..."
		} else {
			name = "Unknown"
		}
	}

	rows := make([]string, 0, 6)
	rows = append(rows, fmt.Sprintf("Picked: %-20s |        ", name))

	picks := snap.CounterPicksBy[heroID]
	for i := 0; i < 5; i++ {
		if i < len(picks) {
			pick := picks[i]
			pickName := snap.HeroIDToName[pick.HeroID]
			if pickName == "" {
				pickName = fmt.Sprintf("ID %d", pick.HeroID)
			}
			rows = append(rows, fmt.Sprintf("%-22s | %5.1f%%", pickName, pick.WinRate*100))
		} else {
			rows = append(rows, fmt.Sprintf("%-22s | %5s", "-", "-"))
		}
	}

	return formatTable("COUNTERS", rows)
}

func buildBestPicksTable(snap state.Snapshot) string {
	rows := make([]string, 0, 6)
	rows = append(rows, fmt.Sprintf("%-22s | %5s", "Best Picks", "Score"))

	for i := 0; i < 5; i++ {
		if i < len(snap.BestCounters) {
			pick := snap.BestCounters[i]
			pickName := snap.HeroIDToName[pick.HeroID]
			if pickName == "" {
				pickName = fmt.Sprintf("ID %d", pick.HeroID)
			}
			rows = append(rows, fmt.Sprintf("%-22s | %5.1f%%", pickName, pick.Score*100))
		} else {
			rows = append(rows, fmt.Sprintf("%-22s | %5s", "-", "-"))
		}
	}

	return formatTable("BEST PICKS", rows)
}

func formatTable(title string, rows []string) string {
	out := title + "\n"
	for _, r := range rows {
		out += r + "\n"
	}
	return out[:len(out)-1]
}

func gsiLine(snap state.Snapshot) string {
	status := snap.GSIStatus
	if status == "" {
		status = "No data"
	}
	if snap.GSILastAt.IsZero() {
		return "GSI: " + status
	}
	age := time.Since(snap.GSILastAt).Round(time.Second)
	return fmt.Sprintf("GSI: %s (last %s ago)", status, age)
}

func buildGSIPanel(snap state.Snapshot) string {
	if snap.GSIHeroID == 0 || snap.GSIMapPhase == "picks" {
		name := snap.GSIHeroName
		if name == "" {
			name = "Not picked"
		}
		return fmt.Sprintf("PICK STAGE\nHero: %s\nMatch: %s", name, fallback(snap.GSIMatchID, "-"))
	}

	return fmt.Sprintf(
		"IN-GAME\nHero: %s (Lv %d)\nK/D/A: %d/%d/%d  LH/D: %d/%d\nGPM/XPM: %d/%d  Gold: %d (%d+%d)\nHP/MP: %d/%d  %d/%d",
		fallback(snap.GSIHeroName, "Unknown"),
		snap.GSIHeroLevel,
		snap.GSIKills,
		snap.GSIDeaths,
		snap.GSIAssists,
		snap.GSILastHits,
		snap.GSIDenies,
		snap.GSIGPM,
		snap.GSIXPM,
		snap.GSIGold,
		snap.GSIGoldR,
		snap.GSIGoldU,
		snap.GSIHeroHP,
		snap.GSIHeroHPMax,
		snap.GSIHeroMP,
		snap.GSIHeroMPMax,
	)
}

func fallback(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
