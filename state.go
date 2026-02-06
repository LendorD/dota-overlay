package main

import "sync"

type GameState struct {
	mu             sync.Mutex
	EnemyHeroesIDs []int
	HeroIDToName   map[int]string
	InternalToID   map[string]int
	IsLocked       bool
	Status         string
	OverlayLogs    []string
}
