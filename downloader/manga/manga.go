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

// New creates a new Manga instance from a JSON file.
// It filters the chapters by volume number.
//
// Parameters:
// - filepath: the path to the JSON file.
// - from: the starting volume number (inclusive).
// - to: the ending volume number (inclusive).
//
// Returns:
// - A pointer to the Manga instance.
// - An error if the loading or filtering process fails.
func New(filepath string, from int, to int) (*Manga, error) {
	// Create a new Manga instance with an empty Chapters slice.
	m := &Manga{
		Chapters: make([]Chapter, 0),
	}

	// Load the Manga instance from the JSON file.
	err := m.load(filepath)
	if err != nil {
		return nil, err
	}

	// Filter the chapters by volume number.
	m.filterByVolume(from, to)

	// Print the number of loaded chapters.
	fmt.Println("loaded", len(m.Chapters), "chapters for", m.Name)

	// Return the Manga instance and nil error.
	return m, nil
}

// load reads the JSON file from the specified filepath and unmarshals it into the Manga instance.
// It returns an error if the file cannot be read or the unmarshaling process fails.
//
// Parameters:
// - filepath: The path to the JSON file.
//
// Returns:
// - error: An error if the file cannot be read or the unmarshaling process fails.
func (m *Manga) load(filepath string) error {
	// Read the JSON file from the specified filepath.
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into the Manga instance.
	err = json.Unmarshal(data, &m)
	if err != nil {
		return err
	}

	// Return nil if the loading process was successful.
	return nil
}

// filterByVolume filters the Manga's Chapters slice by the specified volume numbers.
// It creates a new slice of Chapters with only the chapters that have a volume number
// within the specified range (inclusive).
//
// Parameters:
// - from: The starting volume number (inclusive).
// - to: The ending volume number (inclusive).
func (m *Manga) filterByVolume(from int, to int) {
	// Create a new slice to store the filtered chapters.
	// We preallocate the capacity of the slice to be the difference between the ending and starting volume numbers.
	items := make([]Chapter, 0, to-from+1)

	// Iterate through each chapter in the Manga's Chapters slice.
	for _, item := range m.Chapters {
		// Check if the chapter's volume number is within the specified range.
		if item.VolumeNum >= from && item.VolumeNum <= to {
			// If it is, append the chapter to the new slice.
			items = append(items, item)
		}
	}

	// Replace the Manga's Chapters slice with the filtered slice.
	m.Chapters = items
}
