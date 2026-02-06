package main

import (
	"fmt"
	"image/color"
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type App struct{ State *GameState }

func (a *App) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
		a.State.mu.Lock()
		a.State.IsLocked = !a.State.IsLocked
		ebiten.SetWindowMousePassthrough(a.State.IsLocked)
		a.State.mu.Unlock()
	}
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	// Темный фон
	clr := color.RGBA{10, 10, 20, 220}
	if !a.State.IsLocked {
		clr.R = 50
	}
	ebitenutil.DrawRect(screen, 0, 0, 600, 400, clr)

	a.State.mu.Lock()
	defer a.State.mu.Unlock()

	// Общий статус
	statusMsg := fmt.Sprintf("Status: %s | F12: Lock\n\nENEMIES:", a.State.Status)
	for _, id := range a.State.EnemyHeroesIDs {
		statusMsg += fmt.Sprintf(" [%s]", a.State.HeroIDToName[id])
	}

	// НОВОЕ: Вывод последних строк лога прямо на экран
	statusMsg += "\n\n--- LIVE DOTA LOGS ---"
	for _, line := range a.State.OverlayLogs {
		// Обрезаем длинные строки, чтобы влезали
		if len(line) > 70 {
			line = line[:67] + "..."
		}
		statusMsg += "\n> " + line
	}

	ebitenutil.DebugPrint(screen, statusMsg)
}

func (a *App) Layout(w, h int) (int, int) { return 400, 300 }

func main() {
	state := &GameState{
		Status: "Starting...",
		InternalToID: map[string]int{
			"pudge": 14, "axe": 2, "phantom_lancer": 12, "crystal_maiden": 5,
		},
		HeroIDToName: map[int]string{
			14: "Pudge", 2: "Axe", 12: "PL", 5: "CM",
		},
	}

	logPath := `C:\Program Files (x86)\Steam\steamapps\common\dota 2 beta\game\dota\console.log`

	// Если путь найден автоматически, используем его, но меняем имя файла на console.log
	if autoPath, err := findDotaLogPath(); err == nil {
		logPath = filepath.Join(filepath.Dir(autoPath), "console.log")
	}

	go startParser(state, logPath)

	ebiten.SetWindowSize(400, 300)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetScreenTransparent(true)

	if err := ebiten.RunGame(&App{State: state}); err != nil {
		log.Fatal(err)
	}
}
