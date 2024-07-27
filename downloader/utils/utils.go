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

func GetImageName(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func DownloadImages(imagesUrls []string, dirpath string) error {
	for _, url := range imagesUrls {
		err := DownloadImage(url, dirpath)
		if err != nil {
			return err
		}
	}
	return nil
}

func DownloadImage(url string, filepath string) error {
	slog.Info("download: ", "url", url, " in image: ", filepath)
	resp, err := http.Get(url)
	time.Sleep(500 * time.Millisecond)
	slog.Info("response", "status code", resp.StatusCode, "status", resp.Status, "content length", resp.ContentLength, "body", resp.Body)
	if err != nil {
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
	// save the image
	err = createImage(filepath, resp.Body)
	if err != nil {
		slog.Error(
			"create image error",
			"filepath", filepath,
			"error", err.Error(),
		)
		return err
	}
	return nil
}

func prepareDir(filepath string) (string, error) {
	parts := strings.Split(filepath, "/")
	dirpath := strings.Join(parts[0:len(parts)-1], "/")
	// fmt.Println(filepath, dirpath)
	err := createFolder(dirpath)
	return dirpath, err
}

func createFolder(path string) error {
	if _, err := os.ReadDir(path); err != nil {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			slog.Error("failed to create dir", "path", path, "msg", err.Error())
			return err
		}
	}
	return nil
}

func createImage(filepath string, data io.ReadCloser) error {
	dirpath, _ := prepareDir(filepath)
	slog.Info("saving image", "dirpath", dirpath)
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	n, err := io.Copy(file, data)
	if err != nil {
		return err
	}
	slog.Info("image saved", "filepath", filepath, "written", n)
	return nil
}

func CompressDirectory(zipfilename string, basePath string) error {
	prepareDir(zipfilename)
	outFile, err := os.Create(zipfilename)
	if err != nil {
		return err
	}

	w := zip.NewWriter(outFile)
	if err := addFilesToZip(w, basePath, ""); err != nil {
		_ = outFile.Close()
		slog.Error(
			"add file to zip failed",
			"basePath", basePath,
			"error", err.Error(),
		)
		return err
	}

	if err := w.Close(); err != nil {
		_ = outFile.Close()
		return errors.New("Warning: closing zipfile writer failed: " + err.Error())
	}

	if err := outFile.Close(); err != nil {
		return errors.New("Warning: closing zipfile failed: " + err.Error())
	}

	return nil
}

func addFilesToZip(w *zip.Writer, basePath, baseInZip string) error {
	files, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}

	for _, file := range files {
		fullfilepath := filepath.Join(basePath, file.Name())
		if _, err := os.Stat(fullfilepath); os.IsNotExist(err) {
			// ensure the file exists. For example a symlink pointing to a non-existing location might be listed but not actually exist
			continue
		}

		if file.Type()&os.ModeSymlink != 0 {
			// ignore symlinks alltogether
			continue
		}

		if file.IsDir() {
			if err := addFilesToZip(w, fullfilepath, filepath.Join(baseInZip, file.Name())); err != nil {
				return err
			}
		} else if file.Type().IsRegular() {
			dat, err := os.ReadFile(fullfilepath)
			if err != nil {
				return err
			}
			f, err := w.Create(filepath.Join(baseInZip, file.Name()))
			if err != nil {
				return err
			}
			_, err = f.Write(dat)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
