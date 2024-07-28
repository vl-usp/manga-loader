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
	debugMode bool
}

// parseInput parses the command line flags and returns an Input struct.
// It returns an error if the path to the data file is not specified,
// if the "from" flag is greater than the "to" flag, or if the extension
// of the output file is not supported.
func parseInput() (*Input, error) {
	// Define command line flags
	filepathPtr := flag.String("json", "", "path to data file")
	workerCntPtr := flag.Int("workers", 8, "worker count")
	fromPtr := flag.Int("from", 1, "starting volume number")
	toPtr := flag.Int("to", 1, "ending volume number")
	extPtr := flag.String("ext", "cbz", "extension of output file")
	debugMode := flag.Bool("debug", false, "debug mode")

	// Parse command line flags
	flag.Parse()

	// Check if the path to the data file is specified
	if *filepathPtr == "" {
		return nil, fmt.Errorf("path to data file not specified")
	}

	// Check if the "from" flag is greater than the "to" flag
	if *fromPtr > *toPtr {
		return nil, fmt.Errorf("from (%d) cannot be greater than to (%d)", *fromPtr, *toPtr)
	}

	// Check if the extension of the output file is supported
	if *extPtr != "cbz" && *extPtr != "cbr" && *extPtr != "zip" {
		return nil, fmt.Errorf("extension %s not supported", *extPtr)
	}

	// Return an Input struct with the parsed values
	return &Input{
		workerCnt: *workerCntPtr,
		from:      *fromPtr,
		to:        *toPtr,
		filepath:  *filepathPtr,
		ext:       *extPtr,
		debugMode: *debugMode,
	}, nil
}

func main() {
	// Parse input from command line flags
	i, err := parseInput()
	if err != nil {
		// Log and panic if there is an error parsing input
		slog.Error("input error: " + err.Error())
		panic("input error")
	}

	// Initialize manga struct from the data file
	m, err := manga.New(i.filepath, i.from, i.to)
	if err != nil {
		// Log and panic if there is an error initializing manga struct
		slog.Error("manga init error", "msg", err.Error(), "filepath", i.filepath, "from", i.from, "to", i.to)
		panic("manga init error")
	}

	// Create a new worker pool with specified worker count and chapter count
	workerPool := wp.New(i.workerCnt, len(m.Chapters))

	// Get jobs for loading chapters
	jobs := wp.GetLoadJobs(m, tmpDir)

	// Add jobs to the worker pool
	workerPool.AddJobs(jobs)

	// Log the number of jobs added
	slog.Info("jobs added", "count", len(jobs))

	// Start the worker pool
	workerPool.Start()

	// Log that the worker pool has started
	slog.Info("worker pool started")

	// Collect results from the worker pool
	workerPool.CollectResults()

	// Wait for the worker pool to finish
	workerPool.Wait()

	// Generate output filename based on the input and manga struct
	outputFilename := fmt.Sprintf("%s/%s %d-%d %s.%s", outputDir, m.Name, i.from, i.to, "Volumes", i.ext)
	if i.from == i.to {
		outputFilename = fmt.Sprintf("%s/%s %d %s.%s", outputDir, m.Name, i.from, "Volume", i.ext)
	}

	// Compress the temporary directory into the output file
	slog.Info("compressing", "filename", outputFilename, "dir", tmpDir)
	err = utils.CompressDirectory(outputFilename, tmpDir)
	if err != nil {
		// Log and error if there is an error compressing the directory
		slog.Error("compressing error", "msg", err.Error(), "filename", outputFilename, "dir", outputDir)
	}

	// Remove the temporary directory if debug mode is not enabled
	if !i.debugMode {
		os.RemoveAll(tmpDir)
	}
}
