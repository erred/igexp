package insta

import (
	"io/ioutil"
	"net/http"
)

// Media represents a single media item
type Media struct {
	ID   string
	Type int
	URL  string
}

// Get uses the provided client to retrieve the media
func (m *Media) Get(client *http.Client) ([]byte, error) {
	res, err := client.Get(m.URL)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}
