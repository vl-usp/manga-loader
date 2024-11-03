package loader

import (
	"fmt"
	"log/slog"
	"mangalib-loader/types"
	"mangalib-loader/utils"
	"net/http"
	"os"
	"time"

	"github.com/RomainMichau/cloudscraper_go/cloudscraper"
)

type MangaLoader struct {
	mangaURL  string
	imageURL  string
	workers   int
	volume    int
	extension string

	client *cloudscraper.CloudScrapper
}

func New(mangaSlug string, workers int, volume int, extension string) (*MangaLoader, error) {
	c, err := cloudscraper.Init(false, false)
	if err != nil {
		return nil, err
	}

	return &MangaLoader{
		mangaURL:  fmt.Sprintf("%s/manga/%s", os.Getenv("API_URL"), mangaSlug),
		imageURL:  os.Getenv("IMAGE_URL"),
		workers:   workers,
		volume:    volume,
		extension: extension,

		client: c,
	}, nil
}

func (l *MangaLoader) Load() error {
	manga, err := l.fetchManga()
	if err != nil {
		return err
	}

	chapters, err := l.fetchChapters()
	if err != nil {
		return err
	}

	manga.Chapters = chapters

	// TODO: make it concurrent
	for i, chapter := range manga.Chapters {
		time.Sleep(500 * time.Millisecond)

		pages, err := l.fetchChapterPages(chapter.Volume, chapter.Number)
		if err != nil {
			return err
		}

		manga.Chapters[i].Pages = pages
	}

	err = l.saveManga(manga)
	if err != nil {
		return err
	}

	return nil
}

func (l *MangaLoader) saveManga(manga *types.Manga) error {
	rootDir := "output"
	dir := fmt.Sprintf("%s/%s", rootDir, manga.Name)

	// TODO: make it concurrent
	for _, c := range manga.Chapters {
		for _, p := range c.Pages {
			err := utils.DownloadImage(l.imageURL+p.URL, fmt.Sprintf("%s/%d/%s/%s", dir, l.volume, c.Number, utils.GetImageName(p.URL)))
			if err != nil {
				return err
			}
		}
	}

	err := utils.CompressDirectory(fmt.Sprintf("%s/%s_%d_vol.%s", rootDir, manga.Name, l.volume, l.extension), dir)
	if err != nil {
		return err
	}

	err = utils.DeleteDirectory(dir)
	if err != nil {
		return err
	}

	return nil
}

func (l *MangaLoader) fetchManga() (*types.Manga, error) {
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

func (l *MangaLoader) fetchChapters() ([]types.Chapter, error) {
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

func (l *MangaLoader) fetchChapterPages(volume, chapter string) ([]types.Page, error) {
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
