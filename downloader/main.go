package main

import (
	"downloader/manga"
	"downloader/utils"
	"downloader/wp"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

const (
	outputDir = "output"
	tmpDir    = "tmp"
)

type Input struct {
	workerCnt int
	from      int
	to        int
	filepath  string
	ext       string
}

func parseInput() (*Input, error) {
	filepathPtr := flag.String("json", "", "path to data file")
	workerCntPtr := flag.Int("workers", 8, "worker count")
	fromPtr := flag.Int("from", 1, "worker count")
	toPtr := flag.Int("to", 1, "worker count")
	extPtr := flag.String("ext", "cbz", "extension of output file")
	flag.Parse()

	if *filepathPtr == "" {
		return nil, fmt.Errorf("path to data file not specified")
	}

	if *extPtr != "cbz" || *extPtr != "cbr" || *extPtr != "zip" {
		return nil, fmt.Errorf("extension %s not supported", *extPtr)
	}

	return &Input{
		workerCnt: *workerCntPtr,
		from:      *fromPtr,
		to:        *toPtr,
		filepath:  *filepathPtr,
		ext:       *extPtr,
	}, nil
}

func main() {
	i, err := parseInput()
	if err != nil {
		slog.Error("input error: " + err.Error())
		panic("input error")
	}
	m, err := manga.New(i.filepath, i.from, i.to)
	if err != nil {
		slog.Error("manga init error", "msg", err.Error(), "filepath", i.filepath, "from", i.from, "to", i.to)
		panic("manga init error")
	}

	workerPool := wp.New(i.workerCnt, len(m.Chapters))
	jobs := wp.GetLoadJobs(m, tmpDir)
	workerPool.AddJobs(jobs)
	slog.Info("jobs added", "count", len(jobs))
	workerPool.Start()
	slog.Info("worker pool started")
	workerPool.CollectResults()
	workerPool.Wait()

	outputFilename := fmt.Sprintf("%s/%s %d-%d %s.%s", outputDir, m.Name, i.from, i.to, "Volumes", i.ext)
	if i.from == i.to {
		outputFilename = fmt.Sprintf("%s/%s %d %s.%s", outputDir, m.Name, i.from, "Volume", i.ext)
	}
	slog.Info("compressing", "filename", outputFilename, "dir", tmpDir)
	err = utils.CompressDirectory(outputFilename, tmpDir)
	if err != nil {
		slog.Error("compressing error", "msg", err.Error(), "filename", outputFilename, "dir", outputDir)
	}

	os.RemoveAll(tmpDir)
}
