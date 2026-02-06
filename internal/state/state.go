package state

import "sync"

type Snapshot struct {
	EnemyHeroesIDs  []int
	HeroIDToName    map[int]string
	InternalToID    map[string]int
	IsLocked        bool
	Status          string
	OverlayLogs     []string
	LastCounterHero int
	CounterPicksBy  map[int][]CounterPick
	BestCounters    []ScoredHero
	IsLoading       bool
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

type GameState struct {
	mu              sync.RWMutex
	enemyHeroesIDs  []int
	heroIDToName    map[int]string
	internalToID    map[string]int
	isLocked        bool
	status          string
	overlayLogs     []string
	lastCounterHero int
	counterPicksBy  map[int][]CounterPick
	bestCounters    []ScoredHero
	isLoading       bool
}

func NewGameState(internalToID map[string]int, heroIDToName map[int]string) *GameState {
	return &GameState{
		internalToID: internalToID,
		heroIDToName: heroIDToName,
		status:       "Starting...",
	}
}

func (s *GameState) ToggleLocked() bool {
	s.mu.Lock()
	s.isLocked = !s.isLocked
	locked := s.isLocked
	s.mu.Unlock()
	return locked
}

func (s *GameState) IsLocked() bool {
	s.mu.RLock()
	locked := s.isLocked
	s.mu.RUnlock()
	return locked
}

func (s *GameState) SetStatus(status string) {
	s.mu.Lock()
	s.status = status
	s.mu.Unlock()
}

func (s *GameState) SetLoading(loading bool, status string) {
	s.mu.Lock()
	s.isLoading = loading
	if status != "" {
		s.status = status
	}
	s.mu.Unlock()
}

func (s *GameState) SetMappings(internalToID map[string]int, heroIDToName map[int]string) {
	s.mu.Lock()
	s.internalToID = internalToID
	s.heroIDToName = heroIDToName
	if s.counterPicksBy == nil {
		s.counterPicksBy = make(map[int][]CounterPick)
	}
	s.mu.Unlock()
}

func (s *GameState) AddEnemyHeroByInternalName(internal string) (bool, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.internalToID[internal]
	if !ok {
		return false, 0
	}

	for _, existing := range s.enemyHeroesIDs {
		if existing == id {
			return false, id
		}
	}

	s.enemyHeroesIDs = append(s.enemyHeroesIDs, id)
	return true, id
}

func (s *GameState) EnemyHeroes() []int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]int(nil), s.enemyHeroesIDs...)
}

func (s *GameState) AppendOverlayLog(line string, max int) {
	s.mu.Lock()
	s.overlayLogs = append(s.overlayLogs, line)
	if max > 0 && len(s.overlayLogs) > max {
		s.overlayLogs = s.overlayLogs[len(s.overlayLogs)-max:]
	}
	s.mu.Unlock()
}

func (s *GameState) SetHeroCounters(heroID int, picks []CounterPick) {
	s.mu.Lock()
	if s.counterPicksBy == nil {
		s.counterPicksBy = make(map[int][]CounterPick)
	}
	s.lastCounterHero = heroID
	s.counterPicksBy[heroID] = append([]CounterPick(nil), picks...)
	s.mu.Unlock()
}

func (s *GameState) SetBestCounters(counters []ScoredHero) {
	s.mu.Lock()
	s.bestCounters = append([]ScoredHero(nil), counters...)
	s.mu.Unlock()
}

func (s *GameState) Snapshot(maxLogs int) Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snap := Snapshot{
		EnemyHeroesIDs:  append([]int(nil), s.enemyHeroesIDs...),
		HeroIDToName:    cloneMapIntString(s.heroIDToName),
		InternalToID:    cloneMapStringInt(s.internalToID),
		IsLocked:        s.isLocked,
		Status:          s.status,
		OverlayLogs:     append([]string(nil), s.overlayLogs...),
		LastCounterHero: s.lastCounterHero,
		CounterPicksBy:  cloneCountersMap(s.counterPicksBy),
		BestCounters:    append([]ScoredHero(nil), s.bestCounters...),
		IsLoading:       s.isLoading,
	}

	if maxLogs > 0 && len(snap.OverlayLogs) > maxLogs {
		snap.OverlayLogs = snap.OverlayLogs[len(snap.OverlayLogs)-maxLogs:]
	}

	return snap
}

func cloneCountersMap(src map[int][]CounterPick) map[int][]CounterPick {
	if src == nil {
		return nil
	}
	dst := make(map[int][]CounterPick, len(src))
	for k, v := range src {
		dst[k] = append([]CounterPick(nil), v...)
	}
	return dst
}

func cloneMapIntString(src map[int]string) map[int]string {
	if src == nil {
		return nil
	}
	dst := make(map[int]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneMapStringInt(src map[string]int) map[string]int {
	if src == nil {
		return nil
	}
	dst := make(map[string]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
