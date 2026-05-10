package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"mime"
	"os"
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
