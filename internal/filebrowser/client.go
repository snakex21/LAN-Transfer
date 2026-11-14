package filebrowser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"lantransfer/internal/config"
)

func BrowseRemote(target, path string) (Listing, error) {
	endpoint := fmt.Sprintf("http://%s/api/browse?path=%s", config.TargetAddress(target), url.QueryEscape(path))
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(endpoint)
	if err != nil {
		return Listing{}, err
	}
	defer response.Body.Close()
	if response.StatusCode/100 != 2 {
		return Listing{}, fmt.Errorf("odbiorca zwrócił %s", response.Status)
	}
	var listing Listing
	if err := json.NewDecoder(response.Body).Decode(&listing); err != nil {
		return Listing{}, err
	}
	return listing, nil
}
