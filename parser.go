package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

var overlayPatterns = []*regexp.Regexp{
	regexp.MustCompile(`sHeroSelection:.*npc_dota_hero_([a-z_]+)`),
}

func startParser(s *GameState, path string) {
	// Создаем/очищаем файл отладки при каждом запуске
	debugFile, _ := os.Create("debug_capture.txt")
	defer debugFile.Close()

	for {
		file, err := os.Open(path)
		if err != nil {
			s.mu.Lock()
			s.Status = "Файл логов не найден..."
			s.mu.Unlock()
			time.Sleep(2 * time.Second)
			continue
		}

		// Сбрасываем каретку в конец, чтобы не читать старье
		file.Seek(0, io.SeekEnd)
		reader := bufio.NewReader(file)

		fmt.Println("🚀 Подключено к:", path)

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

			// 1. ПИШЕМ В ТЕРМИНАЛ
			fmt.Println("LOG:", cleanLine)

			// 2. ПИШЕМ В ФАЙЛ
			debugFile.WriteString(cleanLine + "\n")

			// 3. ОБНОВЛЯЕМ ЭКРАН (только отфильтрованные строки)
			s.mu.Lock()
			matched := false
			for _, re := range overlayPatterns {
				m := re.FindStringSubmatch(cleanLine)
				if m != nil {
					matched = true
					if len(m) >= 2 {
						heroName := m[1]
						s.Status = "Detected: " + heroName
						// Тут можно добавить логику добавления ID
					}
					break
				}
			}

			if matched {
				// Добавляем строчку в оверлей для визуализации
				s.OverlayLogs = append(s.OverlayLogs, cleanLine)
				if len(s.OverlayLogs) > 10 {
					s.OverlayLogs = s.OverlayLogs[1:]
				}
			}
			s.mu.Unlock()
		}
		file.Close()
	}
}

