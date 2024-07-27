package wp

import (
	"downloader/manga"
	"downloader/utils"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
)

type LoadJob struct {
	ID       int
	Filepath string
	LoadUrl  *url.URL
}

func GetLoadJobs(data *manga.Manga, dirpath string) []LoadJob {
	imageCnt := 0
	for _, item := range data.Chapters {
		imageCnt += len(item.ImageUrls)
	}
	jobs := make([]LoadJob, 0, imageCnt)
	id := 0
	for _, item := range data.Chapters {
		for _, imageUrl := range item.ImageUrls {
			u, err := url.Parse(imageUrl)
			if err != nil {
				panic(err.Error())
			}
			jobs = append(jobs, LoadJob{
				ID:       id,
				Filepath: fmt.Sprintf("%s/%d/%s %s/%s", dirpath, item.VolumeNum, item.ChapterNum, item.ChapterName, utils.GetImageName(u.Path)),
				LoadUrl:  u,
			})
			id++
		}
	}
	return jobs
}

type ResultJob struct {
	ID       int
	Filepath string
	Message  string
	LoadUrl  *url.URL
}

type WorkerPool struct {
	NumWorkers int
	JobQueue   chan LoadJob
	Results    chan ResultJob
	wg         sync.WaitGroup
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func New(numWorkers, jobQueueSize int) *WorkerPool {
	wp := &WorkerPool{
		NumWorkers: numWorkers,
		JobQueue:   make(chan LoadJob, jobQueueSize),
		Results:    make(chan ResultJob, jobQueueSize),
	}

	slog.Info("worker pool initialized", "num_workers", numWorkers)
	return wp
}

func (wp *WorkerPool) AddJobs(jobs []LoadJob) {
	go func() {
		for _, job := range jobs {
			wp.JobQueue <- job
			slog.Info("job added", "id", job.ID, "filepath", job.Filepath, "url", job.LoadUrl.String())
		}
		defer close(wp.JobQueue)
	}()
}

// RunJobs
func (wp *WorkerPool) Start() {
	for i := 1; i <= wp.NumWorkers; i++ {
		wp.wg.Add(1)
		slog.Info("run worker", "id", i)
		go wp.loadData()
	}
}

// Wait waits for all workers to finish and closes the results channel
func (wp *WorkerPool) Wait() {
	slog.Info("worker pool waiting for workers to finish", "num_workers", wp.NumWorkers)
	wp.wg.Wait()
	close(wp.Results)
}

// CollectResults collects and prints results from the results channel
func (wp *WorkerPool) CollectResults() {
	go func() {
		for result := range wp.Results {
			slog.Info("result received", "id", result.ID, "filepath", result.Filepath, "message", result.Message, "url", result.LoadUrl.String())
		}
	}()
}

// worker function to process jobs from the queue
func (wp *WorkerPool) loadData() {
	defer wp.wg.Done()
	for j := range wp.JobQueue {
		result := ResultJob{
			ID:       j.ID,
			Filepath: j.Filepath,
			LoadUrl:  j.LoadUrl,
		}

		slog.Info("start job", "id", j.ID, "filepath", j.Filepath, "url", j.LoadUrl.String())
		err := utils.DownloadImage(j.LoadUrl.String(), j.Filepath)
		if err != nil {
			result.Message = err.Error()
			slog.Error("load images", "id", j.ID, "filepath", j.Filepath, "url", j.LoadUrl.String(), "msg", err.Error())
		}
		wp.Results <- result
	}
}
