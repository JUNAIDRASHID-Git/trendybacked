package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Downloads an image from Google Drive and saves it locally
func DownloadImageFromDrive(fileID, savePath string) error {
	url := fmt.Sprintf("https://drive.google.com/uc?export=download&id=%s", fileID)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("google Drive returned status %d", resp.StatusCode)
	}

	// Create directory if it doesn't exist
	os.MkdirAll(filepath.Dir(savePath), os.ModePerm)

	outFile, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save image: %v", err)
	}

	return nil
}
