package types

import "encoding/json"

type ChaptersWrapper struct {
	Data []Chapter `json:"data"`
}

type ChapterWrapper struct {
	Data Chapter `json:"data"`
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

func UnwrapChapterJSON(input []byte) (Chapter, error) {
	var wrapper ChapterWrapper

	// Декодируем JSON в структуру
	err := json.Unmarshal(input, &wrapper)
	if err != nil {
		return Chapter{}, err
	}

	// Кодируем структуру обратно в JSON
	return wrapper.Data, nil
}
