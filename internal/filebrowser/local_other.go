//go:build !windows

package filebrowser

import "lantransfer/internal/i18n"

func PickFiles() ([]string, error) {
	return nil, i18n.New("err.pickerWindowsOnly")
}

func PickFolder() (string, error) {
	return "", i18n.New("err.pickerWindowsOnly")
}
