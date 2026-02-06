package gsi

type Payload struct {
	Provider struct {
		Name      string `json:"name"`
		AppID     int    `json:"appid"`
		Version   int    `json:"version"`
		Timestamp int64  `json:"timestamp"`
	} `json:"provider"`

	Map struct {
		Phase     string `json:"phase"`
		MatchID   string `json:"matchid"`
		Name      string `json:"name"`
		GameTime  int    `json:"game_time"`
		ClockTime int    `json:"clock_time"`
	} `json:"map"`

	Player struct {
		SteamID        string `json:"steamid"`
		AccountID      string `json:"accountid"`
		Name           string `json:"name"`
		Team           string `json:"team_name"`
		Activity       string `json:"activity"`
		PlayerSlot     int    `json:"player_slot"`
		TeamSlot       int    `json:"team_slot"`
		Kills          int    `json:"kills"`
		Deaths         int    `json:"deaths"`
		Assists        int    `json:"assists"`
		LastHits       int    `json:"last_hits"`
		Denies         int    `json:"denies"`
		Gold           int    `json:"gold"`
		GoldReliable   int    `json:"gold_reliable"`
		GoldUnreliable int    `json:"gold_unreliable"`
		GPM            int    `json:"gpm"`
		XPM            int    `json:"xpm"`
	} `json:"player"`

	Hero struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Level     int    `json:"level"`
		XP        int    `json:"xp"`
		Health    int    `json:"health"`
		MaxHealth int    `json:"max_health"`
		Mana      int    `json:"mana"`
		MaxMana   int    `json:"max_mana"`
		HealthPct int    `json:"health_percent"`
		ManaPct   int    `json:"mana_percent"`
		Alive     bool   `json:"alive"`
		XPos      int    `json:"xpos"`
		YPos      int    `json:"ypos"`
		Facet     int    `json:"facet"`
	} `json:"hero"`

	Draft struct {
		PicksBans []struct {
			IsPick bool `json:"is_pick"`
			HeroID int  `json:"hero_id"`
			Team   int  `json:"team"`
		} `json:"picks_bans"`
	} `json:"draft"`

	Auth struct {
		Token string `json:"token"`
	} `json:"auth"`
}
