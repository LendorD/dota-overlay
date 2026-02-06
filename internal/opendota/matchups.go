package opendota

import (
	"sort"
	"time"
)

type HeroMatchup struct {
	HeroID      int `json:"hero_id"`
	GamesPlayed int `json:"games_played"`
	Wins        int `json:"wins"`
}

type CounterPick struct {
	HeroID  int
	Games   int
	WinRate float64
}

type ScoredHero struct {
	HeroID int
	Score  float64
}

func CalculateCounters(matchups []HeroMatchup, minGames int, limit int) []CounterPick {
	if minGames < 0 {
		minGames = 0
	}

	counters := make([]CounterPick, 0, len(matchups))
	for _, m := range matchups {
		if m.GamesPlayed < minGames || m.GamesPlayed == 0 {
			continue
		}
		winRateAgainst := 1 - float64(m.Wins)/float64(m.GamesPlayed)
		counters = append(counters, CounterPick{
			HeroID:  m.HeroID,
			Games:   m.GamesPlayed,
			WinRate: winRateAgainst,
		})
	}

	sort.Slice(counters, func(i, j int) bool {
		if counters[i].WinRate == counters[j].WinRate {
			return counters[i].Games > counters[j].Games
		}
		return counters[i].WinRate > counters[j].WinRate
	})

	if limit > 0 && len(counters) > limit {
		return counters[:limit]
	}
	return counters
}

func AnalyzeCounters(
	enemyIDs []int,
	matchupsByEnemy func(enemyID int) ([]HeroMatchup, error),
	minGames int,
	delay time.Duration,
) ([]ScoredHero, error) {
	if len(enemyIDs) == 0 {
		return nil, nil
	}
	if minGames < 0 {
		minGames = 0
	}

	totalScores := make(map[int]float64)

	for _, enemyID := range enemyIDs {
		matchups, err := matchupsByEnemy(enemyID)
		if err != nil {
			return nil, err
		}

		for _, m := range matchups {
			if m.GamesPlayed < minGames || m.GamesPlayed == 0 {
				continue
			}
			enemyWinRate := float64(m.Wins) / float64(m.GamesPlayed)
			ourAdvantage := 1.0 - enemyWinRate
			totalScores[m.HeroID] += ourAdvantage
		}

		if delay > 0 {
			time.Sleep(delay)
		}
	}

	results := make([]ScoredHero, 0, len(totalScores))
	for id, score := range totalScores {
		avgScore := score / float64(len(enemyIDs))
		results = append(results, ScoredHero{HeroID: id, Score: avgScore})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}
