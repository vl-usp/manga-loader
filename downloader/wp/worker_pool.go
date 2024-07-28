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

// AddJobs adds a list of jobs to the worker pool's job queue and
// closes the job queue once all jobs have been added.
//
// jobs: the list of jobs to be added to the job queue.
func (wp *WorkerPool) AddJobs(jobs []LoadJob) {
	// Start a new goroutine to add jobs to the job queue.
	go func() {
		// Iterate over each job in the jobs list.
		for _, job := range jobs {
			// Add the job to the job queue.
			wp.JobQueue <- job
			// Log the addition of the job.
			slog.Info("job added",
				"id", job.ID,
				"filepath", job.Filepath,
				"url", job.LoadUrl.String())
		}
		// Close the job queue once all jobs have been added.
		defer close(wp.JobQueue)
	}()
}

func (wp *WorkerPool) Start() {
	/*
		Start starts the worker pool by creating the specified number of workers.
		Each worker is added to the WaitGroup and then started in a separate goroutine.
	*/
	for i := 1; i <= wp.NumWorkers; i++ {
		wp.wg.Add(1) // Increment the WaitGroup counter

		// Start a new goroutine for each worker
		go func(id int) {
			wp.loadData() // Call the worker function
			wp.wg.Done()  // Decrement the WaitGroup counter
		}(i)

		slog.Info("run worker", "id", i) // Log the start of each worker
	}
}

// Wait waits for all workers to finish and closes the results channel.
//
// This function blocks until all workers have finished. It logs the start of
// waiting and the number of workers. It then waits for the WaitGroup to be
// done, which indicates that all workers have finished. Finally, it closes
// the results channel.
func (wp *WorkerPool) Wait() {
	// Log the start of waiting and the number of workers.
	slog.Info("worker pool waiting for workers to finish", "num_workers", wp.NumWorkers)

	// Wait for the WaitGroup to be done. This blocks until all workers have
	// finished.
	wp.wg.Wait()

	// Close the results channel to signal that no more results will be
	// added to the channel.
	close(wp.Results)
}

// CollectResults collects and prints results from the results channel

// CollectResults starts a goroutine to collect and print results from the results channel.
//
// This function spawns a new goroutine that continuously receives results from the results channel
// and logs them. The function does not return until the results channel is closed.
func (wp *WorkerPool) CollectResults() {
	// Start a new goroutine to collect and print results.
	go func() {
		// Continuously receive results from the results channel.
		for result := range wp.Results {
			// Log each received result.
			slog.Info("result received",
				"id", result.ID, // The ID of the result.
				"filepath", result.Filepath, // The filepath of the result.
				"message", result.Message, // The message of the result.
				"url", result.LoadUrl.String(), // The URL of the result.
			)
		}
	}()
}

// loadData is a function that processes jobs from the queue.
// It runs in a separate goroutine for each worker.
// It receives jobs from the JobQueue channel and processes them.
// For each job, it creates a ResultJob struct, logs the start of the job,
// downloads the image specified by the LoadUrl field, and logs any errors.
// It then sends the ResultJob struct to the Results channel.
func (wp *WorkerPool) loadData() {

	// Continuously receive jobs from the JobQueue channel.
	for j := range wp.JobQueue {
		// Create a new ResultJob struct with the job's ID, Filepath, and LoadUrl fields.
		result := ResultJob{
			ID:       j.ID,
			Filepath: j.Filepath,
			LoadUrl:  j.LoadUrl,
		}

		// Log the start of the job.
		slog.Info("start job", "id", j.ID, "filepath", j.Filepath, "url", j.LoadUrl.String())

		// Download the image specified by the LoadUrl field.
		err := utils.DownloadImage(j.LoadUrl.String(), j.Filepath)

		// If there was an error during the download, log it and set the Message field of the ResultJob struct.
		if err != nil {
			result.Message = err.Error()
			slog.Error("load images", "id", j.ID, "filepath", j.Filepath, "url", j.LoadUrl.String(), "msg", err.Error())
		}

		// Send the ResultJob struct to the Results channel.
		wp.Results <- result
	}
}
