package types

import "encoding/json"

type ChapterWrapper struct {
	Data Chapter `json:"data"`
}

type Page struct {
	ID    int    `json:"id"`
	Image string `json:"image"`
	Slug  int    `json:"slug"`
	URL   string `json:"url"`
}

func UnwrapPagesJSON(input []byte) ([]Page, error) {
	var wrapper ChapterWrapper

	// Декодируем JSON в структуру
	err := json.Unmarshal(input, &wrapper)
	if err != nil {
		return nil, err
	}

	// Кодируем структуру обратно в JSON
	return wrapper.Data.Pages, nil
}
