package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"overlay/internal/state"
)

var overlayPatterns = []*regexp.Regexp{
	regexp.MustCompile(`sHeroSelection:.*npc_dota_hero_([a-z_]+)`),
	regexp.MustCompile(`PR:SetSelectedHero\s+\d+:\[U:1:\d+\]\s+npc_dota_hero_([a-z_]+)\(\d+\)`),
}

func Start(s *state.GameState, path string, onNewHero func(heroID int)) {
	logFile, _ := os.OpenFile("log_dota.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if logFile != nil {
		defer logFile.Close()
	}

	for {
		file, err := os.Open(path)
		if err != nil {
			s.SetStatus("Log file not found...")
			time.Sleep(2 * time.Second)
			continue
		}

		file.Seek(0, io.SeekEnd)
		reader := bufio.NewReader(file)

		fmt.Println("Connected to:", path)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				break
			}

			cleanLine := strings.TrimSpace(line)
			if cleanLine == "" {
				continue
			}

			if logFile != nil {
				logFile.WriteString(cleanLine + "\n")
			}

			matched := false
			var heroInternal string
			for _, re := range overlayPatterns {
				m := re.FindStringSubmatch(cleanLine)
				if m != nil {
					matched = true
					if len(m) >= 2 {
						heroInternal = m[1]
						s.SetStatus("Detected: " + heroInternal)
					}
					break
				}
			}

			if matched {
				s.AppendOverlayLog(cleanLine, 10)
				if heroInternal != "" {
					added, heroID := s.AddEnemyHeroByInternalName(heroInternal)
					if added && onNewHero != nil {
						onNewHero(heroID)
					}
				}
			}
		}
		file.Close()
	}
}
