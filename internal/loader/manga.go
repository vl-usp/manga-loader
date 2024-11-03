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

type chapterJob struct {
	chapter types.Chapter
	err     error
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

	err = l.saveManga(manga)
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
		slog.Error("response body", "body", res.Body)
		return nil, fmt.Errorf("failed to unmarshal chapters: %w", err)
	}

	slog.Info("chapters loaded", "count", len(chapters))

	// Filter chapters by volume
	chapters = types.FilterChapters(chapters, l.volume)

	slog.Info("chapters filtered", "count", len(chapters))

	// Fetch chapter pages concurrently
	chapters, err = l.fetchChapterPages(chapters)
	if err != nil {
		return nil, err
	}

	return chapters, nil
}

func (l *MangaLoader) fetchChapterPagesWorker(volume, chapter string) ([]types.Page, error) {
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

func (l *MangaLoader) fetchChapterPages(chapters []types.Chapter) ([]types.Chapter, error) {
	jobs := make(chan chapterJob, len(chapters))
	results := make(chan chapterJob, len(chapters))
	defer close(results)

	// start workers
	for i := 1; i < l.workers; i++ {
		go func(id int, jobs <-chan chapterJob, results chan<- chapterJob) {
			for job := range jobs {
				slog.Info("fetch chapter pages worker", "id", id, "chapter_num", job.chapter.Number, "volume", job.chapter.Volume)
				pages, err := l.fetchChapterPagesWorker(job.chapter.Volume, job.chapter.Number)
				if err != nil {
					slog.Error("fetch chapter pages error: " + err.Error())
					job.err = err
				}

				job.chapter.Pages = pages
				results <- job

				time.Sleep(500 * time.Millisecond)
			}
		}(i, jobs, results)
	}

	// send jobs
	for _, chapter := range chapters {
		jobs <- chapterJob{
			chapter: chapter,
		}
	}
	close(jobs)

	out := make([]types.Chapter, 0, len(chapters))
	// get results
	for i := 0; i < len(chapters); i++ {
		result := <-results
		if result.err != nil {
			return nil, result.err
		}
		out = append(out, result.chapter)
	}

	return out, nil
}

func (l *MangaLoader) saveChapterWorker(chapter types.Chapter, dirpath string) error {
	for _, page := range chapter.Pages {
		filepath := fmt.Sprintf("%s/%d/%s/%s", dirpath, l.volume, chapter.Number, utils.GetImageName(page.URL))
		err := utils.DownloadImage(l.imageURL+page.URL, filepath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *MangaLoader) saveManga(manga *types.Manga) error {
	rootDir := "output"
	dir := fmt.Sprintf("%s/%s", rootDir, manga.Name)

	jobs := make(chan chapterJob, len(manga.Chapters))
	results := make(chan chapterJob, len(manga.Chapters))
	defer close(results)

	// start workers
	for i := 1; i < l.workers; i++ {
		go func(id int, jobs <-chan chapterJob, results chan<- chapterJob) {
			for job := range jobs {
				slog.Info("save chapter worker", "id", id, "chapter_num", job.chapter.Number, "volume", job.chapter.Volume)
				err := l.saveChapterWorker(job.chapter, dir)
				if err != nil {
					slog.Error("save chapter error: " + err.Error())
					job.err = err
				}
				results <- job

				time.Sleep(500 * time.Millisecond)
			}
		}(i, jobs, results)
	}

	// send jobs
	for _, chapter := range manga.Chapters {
		jobs <- chapterJob{
			chapter: chapter,
		}
	}
	close(jobs)

	// get results from workers
	for i := 0; i < len(manga.Chapters); i++ {
		result := <-results
		if result.err != nil {
			return result.err
		}
	}

	// compress
	err := utils.CompressDirectory(fmt.Sprintf("%s/%s_%d_vol.%s", rootDir, manga.Name, l.volume, l.extension), dir)
	if err != nil {
		return err
	}

	// cleanup
	err = os.RemoveAll(dir)
	if err != nil {
		return err
	}

	return nil
}
