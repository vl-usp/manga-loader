package loader

import (
	"fmt"
	"log/slog"
	"mangalib-loader/types"
	"net/http"
	"os"

	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
)

type Loader struct {
	apiURL    string
	mangaName string

	client *cloudscraper.CloudScrapper
}

func New(mangaName string) (*Loader, error) {
	url := os.Getenv("API_URL")

	c, err := cloudscraper.Init(false, false)
	if err != nil {
		return nil, err
	}

	return &Loader{
		apiURL:    url,
		mangaName: mangaName,

		client: c,
	}, nil
}

func (l *Loader) Load() (*types.Manga, error) {
	manga, err := l.getManga()
	if err != nil {
		return nil, err
	}

	manga, err = l.getChapters(manga)
	if err != nil {
		return nil, err
	}

	return manga, nil
}

func (l *Loader) getManga() (*types.Manga, error) {
	url := l.apiURL + "/manga/" + l.mangaName

	res, err := l.client.Get(url, make(map[string]string), "")
	if err != nil {
		return nil, fmt.Errorf("failed to get manga: %w", err)
	}

	if res.Status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.Status)
	}

	manga, err := types.UnwrapMangaJSON([]byte(res.Body))
	if err != nil {
		// Выводим тело при ошибке декодирования
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &manga, nil
}

func (l *Loader) getChapters(manga *types.Manga) (*types.Manga, error) {
	url := l.apiURL + "/manga/" + l.mangaName

	res, err := l.client.Get(url+"/chapters", make(map[string]string), "")
	if err != nil {
		return nil, fmt.Errorf("failed to get chapters: %w", err)
	}

	if res.Status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.Status)
	}

	chapters, err := types.UnwrapChaptersJSON([]byte(res.Body))
	if err != nil {
		// Выводим тело при ошибке декодирования
		slog.Error("response body", "body", res.Body)
		return nil, fmt.Errorf("failed to unmarshal chapters: %w", err)
	}

	manga.Chapters = make([]types.Chapter, 0, len(chapters))

	// TODO: конкурентность
	for _, chapter := range chapters {
		chapterUrl := fmt.Sprintf("%s/chapter?number=%s&volume=%s", url, chapter.Number, chapter.Volume)
		res, err = l.client.Get(chapterUrl, make(map[string]string), "")
		if err != nil {
			return nil, fmt.Errorf("failed to get chapter: %w", err)
		}

		if res.Status != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", res.Status)
		}

		ch, err := types.UnwrapChapterJSON([]byte(res.Body))
		if err != nil {
			// Выводим тело при ошибке декодирования
			return nil, fmt.Errorf("failed to unmarshal chapter: %w", err)
		}

		manga.Chapters = append(manga.Chapters, ch)
	}

	return manga, nil
}
