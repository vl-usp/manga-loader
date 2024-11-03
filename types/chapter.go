package types

import (
	"encoding/json"
	"strconv"
)

type ChaptersWrapper struct {
	Data []Chapter `json:"data"`
}

type Chapter struct {
	ID     int    `json:"id"`
	Volume string `json:"volume"`
	Number string `json:"number"`
	Name   string `json:"name"`
	Pages  []Page `json:"pages"`
}

func UnwrapChaptersJSON(input []byte) ([]Chapter, error) {
	var wrapper ChaptersWrapper

	// Декодируем JSON в структуру
	err := json.Unmarshal(input, &wrapper)
	if err != nil {
		return nil, err
	}

	// Кодируем структуру обратно в JSON
	return wrapper.Data, nil
}

func FilterChapters(chapters []Chapter, volume int) []Chapter {
	var filtered []Chapter
	for _, chapter := range chapters {
		if chapter.Volume == strconv.Itoa(volume) {
			filtered = append(filtered, chapter)
		}
	}
	return filtered
}
