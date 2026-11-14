package transfer

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func streamDirectory(root string, zw *zip.Writer) error {
	base := filepath.Dir(root)
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		header.Method = zip.Store
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		src, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.CopyBuffer(writer, src, make([]byte, 1024*1024))
		closeErr := src.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}
