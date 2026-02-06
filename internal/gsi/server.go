package gsi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"overlay/internal/state"
)

type Server struct {
	mu       sync.Mutex
	prev     *Payload
	lastSeen time.Time
}

type snapshotResponse struct {
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

func ListenAndServe(
	addr string,
	onEnemyHero func(heroID int),
	onSeen func(),
	st *state.GameState,
) error {

	s := &Server{}

	return http.ListenAndServe(addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			switch r.URL.Path {
			case "/":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
				return
			case "/snapshot":
				snap := st.Snapshot(0)
				resp := snapshotResponse{
					Status:       snap.Status,
					GSIStatus:    snap.GSIStatus,
					GSILastAt:    snap.GSILastAt,
					GSIMatchID:   snap.GSIMatchID,
					GSIMapPhase:  snap.GSIMapPhase,
					GSIMapName:   snap.GSIMapName,
					GSIHeroID:    snap.GSIHeroID,
					GSIHeroName:  snap.GSIHeroName,
					GSIHeroLevel: snap.GSIHeroLevel,
					GSIHeroHP:    snap.GSIHeroHP,
					GSIHeroHPMax: snap.GSIHeroHPMax,
					GSIHeroMP:    snap.GSIHeroMP,
					GSIHeroMPMax: snap.GSIHeroMPMax,
					GSIKills:     snap.GSIKills,
					GSIDeaths:    snap.GSIDeaths,
					GSIAssists:   snap.GSIAssists,
					GSILastHits:  snap.GSILastHits,
					GSIDenies:    snap.GSIDenies,
					GSIGold:      snap.GSIGold,
					GSIGoldR:     snap.GSIGoldR,
					GSIGoldU:     snap.GSIGoldU,
					GSIGPM:       snap.GSIGPM,
					GSIXPM:       snap.GSIXPM,
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		raw, _ := io.ReadAll(r.Body)
		if len(raw) > 0 {
			ts := time.Now().Format(time.RFC3339)
			if logFile, err := os.OpenFile("gsi_log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
				logFile.WriteString(ts + " " + string(raw) + "\n")
				logFile.Close()
			}
			fmt.Println("GSI:", string(raw))
		}

		var p Payload
		if err := json.Unmarshal(raw, &p); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if p.Auth.Token != "" {
			fmt.Println("GSI auth token:", p.Auth.Token)
		}

		onSeen()
		s.handle(&p, st, onEnemyHero)

		w.WriteHeader(http.StatusOK)
	}))
}
