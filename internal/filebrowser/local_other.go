//go:build !windows

package filebrowser

import "errors"

func PickFiles() ([]string, error) {
	return nil, errors.New("systemowe okno wyboru plików jest obecnie dostępne tylko w Windows")
}

func PickFolder() (string, error) {
	return "", errors.New("systemowe okno wyboru folderu jest obecnie dostępne tylko w Windows")
}
