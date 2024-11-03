package main

import (
	"flag"
	"fmt"
	"log/slog"
	"mangalib-loader/internal/loader"

	"github.com/joho/godotenv"
)

type Input struct {
	name      string
	ext       string
	workerCnt int
	volume    int
	debugMode bool
}

// parseInput parses the command line flags and returns an Input struct.
func parseInput() (*Input, error) {
	// Define command line flags
	mangeNamePtr := flag.String("name", "", "manga name from url")
	workerCntPtr := flag.Int("workers", 8, "worker count")
	volNumPtr := flag.Int("vol_num", 1, "starting volume number")
	extPtr := flag.String("ext", "cbz", "extension of output file")
	debugMode := flag.Bool("debug", false, "debug mode")

	// Parse command line flags
	flag.Parse()

	// Check if the path to the data file is specified
	if *mangeNamePtr == "" {
		return nil, fmt.Errorf("manga name from url must be provided")
	}

	// Check if the "from" flag is greater than the "to" flag
	if volNumPtr == nil {
		return nil, fmt.Errorf("number of volume must be provided")
	}

	// Check if the extension of the output file is supported
	if *extPtr != "cbz" && *extPtr != "cbr" && *extPtr != "zip" {
		return nil, fmt.Errorf("extension %s not supported", *extPtr)
	}

	// Return an Input struct with the parsed values
	return &Input{
		workerCnt: *workerCntPtr,
		volume:    *volNumPtr,
		name:      *mangeNamePtr,
		ext:       *extPtr,
		debugMode: *debugMode,
	}, nil
}

func init() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
	}
}

func main() {
	// Parse input from command line flags
	i, err := parseInput()
	if err != nil {
		// Log and panic if there is an error parsing input
		slog.Error("input error: " + err.Error())
		panic("input error")
	}

	// Get manga data from the API
	l, err := loader.New(i.name, i.workerCnt, i.volume, i.ext)
	if err != nil {
		slog.Error("loader init error: " + err.Error())
		panic("loader init error")
	}

	err = l.Load()
	if err != nil {
		slog.Error("load error: " + err.Error())
		panic("load error")
	}
}
