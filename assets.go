package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(mediaType string) string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic("Error generating file name.")
	}
	return base64.RawURLEncoding.EncodeToString(key) + getExtention(mediaType)
}

func getExtention(mediaType string) string {
	splitMediaType := strings.Split(mediaType, "/")
	exts, err := mime.ExtensionsByType(mediaType)
	if err != nil || len(exts) == 0 {
		return ".bin"
	}

	typeIndex := slices.Index(exts, "."+splitMediaType[1])
	if typeIndex != -1 {
		return exts[typeIndex]
	}

	return exts[0]
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func (cfg apiConfig) getObjectURL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	buf := bytes.Buffer{}
	cmd.Stdout = &buf

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffprobe error: %v", err)
	}

	var data struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		return "", fmt.Errorf("could not parse ffprobe output: %v", err)
	}

	if len(data.Streams) == 0 {
		return "", errors.New("Failed to get stream data")
	}

	displayRatio := int64(data.Streams[0].Width) / int64(data.Streams[0].Height)

	switch displayRatio {
	case 1:
		return "landscape/", nil
	case 0:
		return "portrait/", nil
	default:
		return "other/", nil
	}
}

func processVideoForFastStart(filePath string) (string, error) {
	outputFilePath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error processing video: %s, %v", stderr.String(), err)
	}

	fileInfo, err := os.Stat(outputFilePath)
	if err != nil {
		return "", fmt.Errorf("could not stat processed file: %v", err)
	}
	if fileInfo.Size() == 0 {
		return "", fmt.Errorf("processed file is empty")
	}

	return outputFilePath, nil
}
