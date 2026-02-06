package gsi

import "overlay/internal/state"

func (s *Server) handle(p *Payload, st *state.GameState, onEnemyHero func(int)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st.SetGSISnapshot(
		p.Map.Phase,
		p.Map.MatchID,
		p.Map.Name,
		p.Map.GameTime,
		p.Map.ClockTime,
		p.Hero.ID,
		p.Hero.Name,
		p.Hero.Level,
		p.Hero.Health,
		p.Hero.MaxHealth,
		p.Hero.Mana,
		p.Hero.MaxMana,
		p.Player.Kills,
		p.Player.Deaths,
		p.Player.Assists,
		p.Player.LastHits,
		p.Player.Denies,
		p.Player.Gold,
		p.Player.GoldReliable,
		p.Player.GoldUnreliable,
		p.Player.GPM,
		p.Player.XPM,
	)

	if s.prev == nil {
		s.prev = p
		return
	}

	prevHero := s.prev.Hero.ID
	currHero := p.Hero.ID
	if currHero > 0 && currHero != prevHero {
		onEnemyHero(currHero)
	}

	s.prev = p
}
