package main

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"overlay/internal/app"
	"overlay/internal/gsi"
	"overlay/internal/opendota"
	"overlay/internal/parser"
	"overlay/internal/paths"
	"overlay/internal/state"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	st := state.NewGameState(
		map[string]int{
			"pudge":          14,
			"axe":            2,
			"phantom_lancer": 12,
			"crystal_maiden": 5,
		},
		map[int]string{
			14: "Pudge",
			2:  "Axe",
			12: "PL",
			5:  "CM",
		},
	)

	logPath := `C:\Program Files (x86)\Steam\steamapps\common\dota 2 beta\game\dota\console.log`
	if autoPath, err := paths.FindDotaLogPath(); err == nil {
		logPath = filepath.Join(filepath.Dir(autoPath), "console.log")
	}

	client := opendota.NewClient(os.Getenv("OPENDOTA_API_KEY"))

	go func() {
		st.SetLoading(true, "Fetching OpenDota heroes...")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		heroes, err := client.GetHeroes(ctx)
		if err != nil {
			st.SetLoading(false, "OpenDota heroes error: "+err.Error())
			return
		}

		internalToID := make(map[string]int, len(heroes))
		idToName := make(map[int]string, len(heroes))
		for _, h := range heroes {
			idToName[h.ID] = h.LocalizedName
			internal := strings.TrimPrefix(h.Name, "npc_dota_hero_")
			if internal != "" {
				internalToID[internal] = h.ID
			}
		}

		st.SetMappings(internalToID, idToName)
		st.SetLoading(false, "Ready")
	}()

	onNewHero := func(heroID int) {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			matchups, err := client.GetHeroMatchups(ctx, heroID)
			if err != nil {
				st.SetStatus("OpenDota error: " + err.Error())
				return
			}

			counters := opendota.CalculateCounters(matchups, 20, 5)
			picks := make([]state.CounterPick, 0, len(counters))
			for _, c := range counters {
				picks = append(picks, state.CounterPick{
					HeroID:  c.HeroID,
					Games:   c.Games,
					WinRate: c.WinRate,
				})
			}
			st.SetHeroCounters(heroID, picks)

			enemies := st.EnemyHeroes()
			results, err := opendota.AnalyzeCounters(
				enemies,
				func(enemyID int) ([]opendota.HeroMatchup, error) {
					reqCtx, reqCancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer reqCancel()
					return client.GetHeroMatchups(reqCtx, enemyID)
				},
				10,
				100*time.Millisecond,
			)
			if err != nil {
				st.SetStatus("OpenDota analyze error: " + err.Error())
				return
			}

			enemySet := make(map[int]struct{}, len(enemies))
			for _, id := range enemies {
				enemySet[id] = struct{}{}
			}

			best := make([]state.ScoredHero, 0, len(results))
			for _, r := range results {
				if _, isEnemy := enemySet[r.HeroID]; isEnemy {
					continue
				}
				best = append(best, state.ScoredHero{HeroID: r.HeroID, Score: r.Score})
				if len(best) >= 10 {
					break
				}
			}
			st.SetBestCounters(best)
		}()
	}

	go func() {
		err := gsi.ListenAndServe("127.0.0.1:3001", func(heroID int) {
			if added := st.AddEnemyHeroByID(heroID); added {
				onNewHero(heroID)
			}
		}, func() {
			st.SetGSISeen(time.Now())
		}, st)
		if err != nil {
			st.SetStatus("GSI error: " + err.Error())
		}
	}()

	go func() {
		st.SetGSIStatus("GSI self-test...")
		client := &http.Client{Timeout: 2 * time.Second}
		body := []byte(`{"player":{"team_name":"spectator"},"draft":{"picks_bans":[]}}`)
		for i := 0; i < 5; i++ {
			resp, err := client.Post("http://127.0.0.1:3001/", "application/json", bytes.NewReader(body))
			if err == nil {
				resp.Body.Close()
				st.SetGSIStatus("GSI self-test OK")
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
		st.SetGSIStatus("GSI self-test failed")
	}()

	go parser.Start(st, logPath, onNewHero)

	ebiten.SetWindowSize(app.ViewWidth, app.ViewHeight)
	ebiten.SetWindowFloating(true)
	ebiten.SetWindowDecorated(false)
	ebiten.SetScreenTransparent(true)

	if err := ebiten.RunGame(app.New(st)); err != nil {
		log.Fatal(err)
	}
}
