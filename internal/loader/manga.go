package loader

import (
	"fmt"
	"log/slog"
	"mangalib-loader/types"
	"net/http"
	"os"
	"time"

	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
)

type MangaLoader struct {
	mangaURL string
	workers  int
	volume   int

	client *cloudscraper.CloudScrapper
}

func New(mangaSlug string, workers int, volume int) (*MangaLoader, error) {
	c, err := cloudscraper.Init(false, false)
	if err != nil {
		return nil, err
	}

	return &MangaLoader{
		mangaURL: fmt.Sprintf("%s/manga/%s", os.Getenv("API_URL"), mangaSlug),
		workers:  workers,
		volume:   volume,

		client: c,
	}, nil
}

func (l *MangaLoader) Load() (*types.Manga, error) {
	manga, err := l.getManga()
	if err != nil {
		return nil, err
	}

	chapters, err := l.getChapters()
	if err != nil {
		return nil, err
	}

	manga.Chapters = chapters

	// конкурентность
	for i, chapter := range manga.Chapters {
		time.Sleep(500 * time.Millisecond)

		pages, err := l.getChapterPages(chapter.Volume, chapter.Number)
		if err != nil {
			return nil, err
		}

		manga.Chapters[i].Pages = pages
	}

	return manga, nil
}

func (l *MangaLoader) SaveManga() error {

	return nil
}

func (l *MangaLoader) getManga() (*types.Manga, error) {
	res, err := l.client.Get(l.mangaURL, make(map[string]string), "")
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

	slog.Info("manga loaded", "manga", manga)

	return &manga, nil
}

func (l *MangaLoader) getChapters() ([]types.Chapter, error) {
	res, err := l.client.Get(l.mangaURL+"/chapters", make(map[string]string), "")
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

	slog.Info("chapters loaded", "count", len(chapters))

	// Filter chapters by volume
	chapters = types.FilterChapters(chapters, l.volume)

	slog.Info("chapters filtered", "count", len(chapters))

	return chapters, nil
}

func (l *MangaLoader) getChapterPages(volume, chapter string) ([]types.Page, error) {
	chapterUrl := fmt.Sprintf("%s/chapter?number=%s&volume=%s", l.mangaURL, chapter, volume)
	res, err := l.client.Get(chapterUrl, make(map[string]string), "")
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter: %w", err)
	}

	if res.Status != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.Status)
	}

	pages, err := types.UnwrapPagesJSON([]byte(res.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal chapter: %w", err)
	}

	slog.Info("pages loaded", "volume", volume, "chapter", chapter, "count", len(pages))

	return pages, nil
}
