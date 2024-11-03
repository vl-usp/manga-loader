package utils

import (
	"archive/zip"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GetImageName returns the name of the image from its path.
//
// It splits the path by "/" to get the parts, and then returns the last part,
// which is the name of the image.
//
// Parameters:
// - path: the path of the image.
//
// Returns:
// - string: the name of the image.
func GetImageName(path string) string {
	// Split the path by "/" to get the parts
	parts := strings.Split(path, "/")
	// Return the last part, which is the name of the image
	return parts[len(parts)-1]
}

// DownloadImages downloads multiple images from the given URLs and saves them to the specified directory.
//
// Parameters:
// - imagesUrls: the URLs of the images to download.
// - dirpath: the directory where the downloaded images will be saved.
//
// Returns:
// - error: an error if there was a problem downloading or saving any image.
func DownloadImages(imagesUrls []string, dirpath string) error {
	// Iterate over each URL and download the image
	for _, url := range imagesUrls {
		// Download the image and check for errors
		err := DownloadImage(url, dirpath)
		if err != nil {
			// If there is an error, return it immediately
			return err
		}
	}
	// If all images were downloaded successfully, return nil
	return nil
}

// DownloadImage downloads an image from the given URL and saves it to the specified filepath.
//
// Parameters:
// - url: the URL of the image to download.
// - filepath: the path where the downloaded image will be saved.
//
// Returns:
// - error: an error if there was a problem downloading or saving the image.
func DownloadImage(url string, filepath string) error {
	// Log the start of the download
	slog.Info("download: ", "url", url, " in image: ", filepath)

	// Send a GET request to the URL
	resp, err := http.Get(url)

	// Wait for a short time to prevent overloading the server
	time.Sleep(500 * time.Millisecond)

	// Log the response status and other details
	slog.Info("response", "status code", resp.StatusCode, "status", resp.Status, "content length", resp.ContentLength, "body", resp.Body)

	// Check for errors during the GET request
	if err != nil {
		// Log the error and return it
		slog.Error(
			"error downloading url",
			"url", url, "error", err.Error(),
			"status code", resp.StatusCode,
			"status", resp.Status,
			"content length", resp.ContentLength,
		)
		return err
	}
	defer resp.Body.Close()

	// Save the image
	err = createImage(filepath, resp.Body)
	if err != nil {
		// Log the error and return it
		slog.Error(
			"create image error",
			"filepath", filepath,
			"error", err.Error(),
		)
		return err
	}

	return nil
}

// prepareDir prepares the directory for a file specified by its filepath.
//
// It splits the filepath by "/" to get the directory path, then joins the parts
// excluding the last one to get the directory path. Finally, it calls createFolder
// to create the directory if it doesn't exist and returns the directory path.
//
// Parameters:
// - filepath: the path of the file.
//
// Returns:
// - dirpath: the path of the directory.
// - err: an error if it occurs during the directory creation process.
func prepareDir(filepath string) (string, error) {
	// Split the filepath by "/" to get the directory path
	parts := strings.Split(filepath, "/")
	// Join the parts excluding the last one to get the directory path
	dirpath := strings.Join(parts[0:len(parts)-1], "/")
	// Call createFolder to create the directory if it doesn't exist
	err := createFolder(dirpath)
	// Return the directory path and the error
	return dirpath, err
}

// createFolder creates a directory at the specified path if it doesn't exist.
//
// It checks if the directory already exists by attempting to read its contents.
// If the directory does not exist, it creates it using os.MkdirAll.
// If there is an error during the process, it logs the error and returns it.
// Otherwise, it returns nil.
func createFolder(path string) error {
	// Check if the directory already exists
	if _, err := os.ReadDir(path); err != nil {
		// If the directory does not exist, create it
		err = os.MkdirAll(path, 0755)
		if err != nil {
			// Log the error and return it
			slog.Error("failed to create dir", "path", path, "msg", err.Error())
			return err
		}
	}
	return nil
}

// createImage saves an image from an io.ReadCloser to a file specified by filepath.
//
// It creates the directory for the file if it doesn't exist, and then saves the
// image to the file. If there is an error during any of these steps, it returns
// the error. Otherwise, it returns nil.
func createImage(filepath string, data io.ReadCloser) error {
	// Prepare the directory for the file
	dirpath, _ := prepareDir(filepath)

	// Log the directory path
	slog.Info("saving image", "dirpath", dirpath)

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy the data from the ReadCloser to the file
	n, err := io.Copy(file, data)
	if err != nil {
		return err
	}

	// Log the file path and the number of bytes written
	slog.Info("image saved", "filepath", filepath, "written", n)

	// Return nil if everything was successful
	return nil
}

// CompressDirectory compresses a directory into a zip file.
//
// It creates a zip file at the specified zipfilename, and adds all files in the
// basePath directory to it.
// If the function encounters an error, it returns the error. Otherwise, it
// returns nil.
func CompressDirectory(zipfilename string, basePath string) error {
	// Ensure the directory for the zip file exists
	_, err := prepareDir(zipfilename)
	if err != nil {
		slog.Error("prepare dir error", "error", err.Error())
	}

	// Create the zip file
	outFile, err := os.Create(zipfilename)
	if err != nil {
		return err
	}

	// Create a new zip writer
	w := zip.NewWriter(outFile)

	// Add all files in the basePath directory to the zip file
	if err := addFilesToZip(w, basePath, ""); err != nil {
		// Close the file if adding files to the zip failed
		_ = outFile.Close()
		// Log the error
		slog.Error(
			"add file to zip failed",
			"basePath", basePath,
			"error", err.Error(),
		)
		return err
	}

	// Close the zip writer
	if err := w.Close(); err != nil {
		// Close the file if closing the zip writer failed
		_ = outFile.Close()
		// Return the error
		return errors.New("Warning: closing zipfile writer failed: " + err.Error())
	}

	// Close the file
	if err := outFile.Close(); err != nil {
		// Return the error
		return errors.New("Warning: closing zipfile failed: " + err.Error())
	}

	// Return nil if everything was successful
	return nil
}

// addFilesToZip adds files from a directory to a zip archive.
// It recursively adds all files in the directory to the zip archive.
//
// Parameters:
// - w: The zip writer to write the files to.
// - basePath: The base path of the directory to add files from.
// - baseInZip: The base path of the directory in the zip archive.
//
// Returns:
// - error: An error if any occurred during the process.
func addFilesToZip(w *zip.Writer, basePath, baseInZip string) error {
	// Read the files in the directory
	files, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}

	// Iterate over each file
	for _, file := range files {
		// Get the full path of the file
		fullfilepath := filepath.Join(basePath, file.Name())

		// Skip the file if it doesn't exist (symlink pointing to a non-existing location)
		if _, err := os.Stat(fullfilepath); os.IsNotExist(err) {
			continue
		}

		// Skip symlinks alltogether
		if file.Type()&os.ModeSymlink != 0 {
			continue
		}

		// If the file is a directory, recursively add its files to the zip archive
		if file.IsDir() {
			if err := addFilesToZip(w, fullfilepath, filepath.Join(baseInZip, file.Name())); err != nil {
				return err
			}
		} else if file.Type().IsRegular() {
			// If the file is a regular file, add it to the zip archive
			dat, err := os.ReadFile(fullfilepath)
			if err != nil {
				return err
			}
			// Create a new file in the zip archive with the relative path to the base path
			f, err := w.Create(filepath.Join(baseInZip, file.Name()))
			if err != nil {
				return err
			}
			// Write the file data to the zip file
			_, err = f.Write(dat)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
