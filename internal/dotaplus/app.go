package dotaplus

import (
	"encoding/json"
	"fmt"
	"image/color"
	"net/http"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ViewWidth  = 360
	ViewHeight = 220
)

const fetchInterval = 500 * time.Millisecond

type Snapshot struct {
	Status       string    `json:"status"`
	GSIStatus    string    `json:"gsi_status"`
	GSILastAt    time.Time `json:"gsi_last_at"`
	GSIMatchID   string    `json:"gsi_match_id"`
	GSIMapPhase  string    `json:"gsi_map_phase"`
	GSIMapName   string    `json:"gsi_map_name"`
	GSIHeroID    int       `json:"gsi_hero_id"`
	GSIHeroName  string    `json:"gsi_hero_name"`
	GSIHeroLevel int       `json:"gsi_hero_level"`
	GSIHeroHP    int       `json:"gsi_hero_hp"`
	GSIHeroHPMax int       `json:"gsi_hero_hp_max"`
	GSIHeroMP    int       `json:"gsi_hero_mp"`
	GSIHeroMPMax int       `json:"gsi_hero_mp_max"`
	GSIKills     int       `json:"gsi_kills"`
	GSIDeaths    int       `json:"gsi_deaths"`
	GSIAssists   int       `json:"gsi_assists"`
	GSILastHits  int       `json:"gsi_last_hits"`
	GSIDenies    int       `json:"gsi_denies"`
	GSIGold      int       `json:"gsi_gold"`
	GSIGoldR     int       `json:"gsi_gold_r"`
	GSIGoldU     int       `json:"gsi_gold_u"`
	GSIGPM       int       `json:"gsi_gpm"`
	GSIXPM       int       `json:"gsi_xpm"`
}

type App struct {
	mu          sync.RWMutex
	snap        Snapshot
	lastFetch   time.Time
	lastUpdated time.Time
	fetching    bool
	fetchErr    string

	dragging     bool
	dragStartX   int
	dragStartY   int
	windowStartX int
	windowStartY int

	resizing     bool
	resizeStartX int
	resizeStartY int
	windowStartW int
	windowStartH int
}

func New() *App {
	return &App{}
}

func (a *App) Update() error {
	a.handleDragResize()

	if time.Since(a.lastFetch) < fetchInterval {
		return nil
	}
	if a.fetching {
		return nil
	}
	a.fetching = true
	a.lastFetch = time.Now()

	go a.fetchSnapshot()
	return nil
}

func (a *App) handleDragResize() {
	const headerHeight = 34
	const resizeZone = 16
	const minW = 260
	const minH = 160

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		winW, winH := ebiten.WindowSize()

		inResize := x >= winW-resizeZone && y >= winH-resizeZone
		inHeader := y <= headerHeight

		if inResize {
			a.resizing = true
			a.resizeStartX, a.resizeStartY = x, y
			a.windowStartW, a.windowStartH = winW, winH
		} else if inHeader {
			a.dragging = true
			a.dragStartX, a.dragStartY = x, y
			a.windowStartX, a.windowStartY = ebiten.WindowPosition()
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if a.resizing {
			x, y := ebiten.CursorPosition()
			dx := x - a.resizeStartX
			dy := y - a.resizeStartY
			newW := a.windowStartW + dx
			newH := a.windowStartH + dy
			if newW < minW {
				newW = minW
			}
			if newH < minH {
				newH = minH
			}
			ebiten.SetWindowSize(newW, newH)
		} else if a.dragging {
			x, y := ebiten.CursorPosition()
			dx := x - a.dragStartX
			dy := y - a.dragStartY
			ebiten.SetWindowPosition(a.windowStartX+dx, a.windowStartY+dy)
		}
	} else {
		a.dragging = false
		a.resizing = false
	}
}

func (a *App) fetchSnapshot() {
	defer func() { a.fetching = false }()

	resp, err := http.Get("http://127.0.0.1:3001/snapshot")
	if err != nil {
		a.setFetchError(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.setFetchError(fmt.Errorf("status %d", resp.StatusCode))
		return
	}

	var snap Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		a.setFetchError(err)
		return
	}

	a.mu.Lock()
	a.snap = snap
	a.fetchErr = ""
	a.lastUpdated = time.Now()
	a.mu.Unlock()
}

func (a *App) setFetchError(err error) {
	a.mu.Lock()
	a.fetchErr = err.Error()
	a.mu.Unlock()
}

func (a *App) Draw(screen *ebiten.Image) {
	w, h := screen.Size()

	// Transparent background with small panels.
	ebitenutil.DrawRect(screen, 0, 0, float64(w), float64(h), color.RGBA{0, 0, 0, 0})
	ebitenutil.DrawRect(screen, 0, 0, float64(w), 34, color.RGBA{20, 22, 34, 200})
	ebitenutil.DebugPrintAt(screen, "DOTA PLUS", 12, 8)

	snap, errText, updatedAt := a.snapshot()

	statusLine := fmt.Sprintf("GSI: %s", fallback(snap.GSIStatus, "No data"))
	if !snap.GSILastAt.IsZero() {
		age := time.Since(snap.GSILastAt).Round(time.Second)
		statusLine += " | last " + age.String() + " ago"
	}
	if errText != "" {
		statusLine = "GSI error: " + errText
	}

	if !updatedAt.IsZero() {
		statusLine += " | updated " + time.Since(updatedAt).Round(time.Second).String() + " ago"
	}

	ebitenutil.DrawRect(screen, 10, 42, float64(w-20), 38, color.RGBA{18, 20, 32, 200})
	ebitenutil.DebugPrintAt(screen, statusLine, 16, 52)

	panelY := 90
	panelH := h - panelY - 10
	if panelH < 90 {
		panelH = 90
	}
	ebitenutil.DrawRect(screen, 10, float64(panelY), float64(w-20), float64(panelH), color.RGBA{16, 18, 28, 200})

	lines := buildPanelLines(snap)
	y := panelY + 10
	for _, line := range lines {
		ebitenutil.DebugPrintAt(screen, line, 16, y)
		y += 16
	}

	// Resize grip
	ebitenutil.DrawRect(screen, float64(w-16), float64(h-16), 16, 16, color.RGBA{35, 38, 52, 200})
}

func (a *App) snapshot() (Snapshot, string, time.Time) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.snap, a.fetchErr, a.lastUpdated
}

func (a *App) Layout(w, h int) (int, int) {
	return w, h
}

func buildPanelLines(snap Snapshot) []string {
	lines := make([]string, 0, 12)
	lines = append(lines, "MODE")
	if snap.GSIHeroID == 0 || snap.GSIMapPhase == "picks" {
		hero := fallback(snap.GSIHeroName, "Not picked")
		lines = append(lines, "Pick Stage")
		lines = append(lines, "Hero: "+hero)
		lines = append(lines, fmt.Sprintf("HeroID: %d", snap.GSIHeroID))
		lines = append(lines, "Match: "+fallback(snap.GSIMatchID, "-"))
		lines = append(lines, "Map: "+fallback(snap.GSIMapName, "-"))
		lines = append(lines, "Phase: "+fallback(snap.GSIMapPhase, "-"))
		return lines
	}

	lines = append(lines, "In-Game")
	lines = append(lines, fmt.Sprintf("Hero: %s (Lv %d)", fallback(snap.GSIHeroName, "Unknown"), snap.GSIHeroLevel))
	lines = append(lines, fmt.Sprintf("HeroID: %d", snap.GSIHeroID))
	lines = append(lines, fmt.Sprintf("K/D/A: %d/%d/%d", snap.GSIKills, snap.GSIDeaths, snap.GSIAssists))
	lines = append(lines, fmt.Sprintf("LH/D: %d/%d", snap.GSILastHits, snap.GSIDenies))
	lines = append(lines, fmt.Sprintf("GPM/XPM: %d/%d", snap.GSIGPM, snap.GSIXPM))
	lines = append(lines, fmt.Sprintf("Gold: %d (%d+%d)", snap.GSIGold, snap.GSIGoldR, snap.GSIGoldU))
	lines = append(lines, fmt.Sprintf("HP/MP: %d/%d  %d/%d", snap.GSIHeroHP, snap.GSIHeroHPMax, snap.GSIHeroMP, snap.GSIHeroMPMax))
	lines = append(lines, "Match: "+fallback(snap.GSIMatchID, "-"))
	lines = append(lines, "Map: "+fallback(snap.GSIMapName, "-"))
	lines = append(lines, "Phase: "+fallback(snap.GSIMapPhase, "-"))
	return lines
}

func fallback(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
