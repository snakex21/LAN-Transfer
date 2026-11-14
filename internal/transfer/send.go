package transfer

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"lantransfer/internal/config"
)

func SendPath(path, target, selfID string) error {
	return SendPathTo(path, target, selfID, "")
}

func SendPathTo(path, target, selfID, destination string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	ip, err := resolveTarget(target, selfID)
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	multipartWriter := multipart.NewWriter(writer)
	go func() {
		var writeErr error
		defer func() {
			if writeErr == nil {
				writeErr = multipartWriter.Close()
			}
			_ = writer.CloseWithError(writeErr)
		}()

		if destination != "" {
			field, err := multipartWriter.CreateFormField("destination")
			if err != nil {
				writeErr = err
				return
			}
			if _, err := io.WriteString(field, destination); err != nil {
				writeErr = err
				return
			}
		}

		uploadName := filepath.Base(path)
		if info.IsDir() {
			uploadName += ".zip"
		}
		part, err := multipartWriter.CreateFormFile("file", uploadName)
		if err != nil {
			writeErr = err
			return
		}

		if info.IsDir() {
			zipWriter := zip.NewWriter(part)
			writeErr = streamDirectory(path, zipWriter)
			if closeErr := zipWriter.Close(); writeErr == nil {
				writeErr = closeErr
			}
			return
		}

		file, err := os.Open(path)
		if err != nil {
			writeErr = err
			return
		}
		defer file.Close()
		_, writeErr = io.CopyBuffer(part, file, make([]byte, 1024*1024))
	}()

	url := fmt.Sprintf("http://%s/api/receive", config.TargetAddress(ip))
	request, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	response, err := (&http.Client{}).Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode/100 != 2 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("odbiorca zwrócił %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	return nil
}
