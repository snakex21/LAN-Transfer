package transfer

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Receiver struct {
	Directory string
}

func (r Receiver) HandleHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	multipartReader, err := req.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	destinationDir := r.Directory
	for {
		part, err := multipartReader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if part.FormName() == "destination" {
			data, err := io.ReadAll(io.LimitReader(part, 32*1024))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			requested := string(data)
			if requested != "" {
				info, err := os.Stat(requested)
				if err != nil || !info.IsDir() {
					http.Error(w, "nieprawidłowy folder docelowy", http.StatusBadRequest)
					return
				}
				destinationDir = requested
			}
			continue
		}

		name := filepath.Base(part.FileName())
		if name == "" || name == "." {
			continue
		}
		if err := receivePart(destinationDir, name, part); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"ok":true}`)
}

func receivePart(directory, name string, source io.Reader) error {
	destination, err := os.Create(filepath.Join(directory, name))
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(destination, source)
	closeErr := destination.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
