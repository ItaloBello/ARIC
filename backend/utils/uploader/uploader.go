package uploader

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UploadResult struct {
	FileName string
	FilePath string
	URL      string
}

type UploadConfig struct {
	DestFolder   string
	BaseURL      string
	MaxSize      int64
	AllowedTypes []string
}

func SaveFile(file *multipart.FileHeader, config UploadConfig) (*UploadResult, error) {

	if file.Size > config.MaxSize {
		return nil, errors.New("arquivo muito grande")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := false
	for _, t := range config.AllowedTypes {
		if ext == t {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, errors.New("tipo de arquivo não permitido")
	}

	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(config.DestFolder, fileName)

	if err := os.MkdirAll(config.DestFolder, 0755); err != nil {
		return nil, err
	}

	if err := save(file, filePath); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", config.BaseURL, fileName)

	return &UploadResult{
		FileName: fileName,
		FilePath: filePath,
		URL:      url,
	}, nil
}

func save(file *multipart.FileHeader, path string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = dst.ReadFrom(src)
	return err
}

func DeleteFile(url string) error {
	if url == "" {
		return nil
	}
	path := url
	if strings.HasPrefix(path, "/") {
		path = "." + path
	}
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
