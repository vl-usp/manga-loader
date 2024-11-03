package types

import "encoding/json"

type MangaWrapper struct {
	Data Manga `json:"data"`
}

type Manga struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	RusName  string `json:"rus_name"`
	Slug     string `json:"slug"`
	Chapters []Chapter
}

func UnwrapMangaJSON(input []byte) (Manga, error) {
	var wrapper MangaWrapper

	// Декодируем JSON в структуру
	err := json.Unmarshal(input, &wrapper)
	if err != nil {
		return Manga{}, err
	}

	// Кодируем структуру обратно в JSON
	return wrapper.Data, nil
}
