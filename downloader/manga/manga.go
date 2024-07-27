package manga

import (
	"encoding/json"
	"fmt"
	"os"
)

type Manga struct {
	Id       int       `json:"id"`
	Name     string    `json:"name"`
	RusName  string    `json:"rusName"`
	Slug     string    `json:"slug"`
	Chapters []Chapter `json:"chapters"`
}

type Chapter struct {
	Id          int      `json:"chapter_id"`
	VolumeNum   int      `json:"chapter_volume"`
	ChapterName string   `json:"chapter_name"`
	ChapterNum  string   `json:"chapter_number"`
	Username    string   `json:"username"`
	ImageUrls   []string `json:"urls"`
}

func New(filepath string, from int, to int) (*Manga, error) {
	m := &Manga{
		Chapters: make([]Chapter, 0),
	}
	err := m.load(filepath)
	if err != nil {
		return nil, err
	}
	m.filterByVolume(from, to)
	fmt.Println("loaded", len(m.Chapters), "chapters for", m.Name)
	return m, nil
}

func (m *Manga) load(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manga) filterByVolume(from int, to int) {
	if from < 0 || to < 0 || from > to {
		panic("Invalid input")
	}
	var items []Chapter
	for _, item := range m.Chapters {
		if item.VolumeNum >= from && item.VolumeNum <= to {
			items = append(items, item)
		}
	}
	m.Chapters = items
}
