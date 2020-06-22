package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/vburenin/go-vphoto/authstate"
	"github.com/vburenin/go-vphoto/models"
	"github.com/vburenin/go-vphoto/settings"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/flosch/pongo2"
	log "github.com/sirupsen/logrus"
)

func listAllAlbums(c *http.Client) ([]*models.Album, error) {
	nextPageToken := ""
	baseUrl, _ := url.Parse("https://photoslibrary.googleapis.com/v1/albums")
	albums := make([]*models.Album, 0, 25)

	for ;; {
		params := url.Values{}
		if nextPageToken != "" {
			params.Add("pageToken", nextPageToken)
		}
		params.Add("pageSize", "50")
		baseUrl.RawQuery = params.Encode()

		resp, err := c.Get(baseUrl.String())
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		decodedResponse := &models.AlbumResponse{}
		err = json.Unmarshal(body, decodedResponse)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal response: %s", err)
		}
		albums = append(albums, decodedResponse.Albums...)
		if decodedResponse.NextPageToken == "" {
			break
		}
		nextPageToken = decodedResponse.NextPageToken
	}
	return albums, nil
}

func ListAlbums(w http.ResponseWriter, r *http.Request) {
	c := authstate.GetHttpClient()
	if c == nil {
		log.Warn("client is not initialized")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	albums, _ := listAllAlbums(c)

	settings.AlbumTemplate.ExecuteWriter(pongo2.Context{
		"albums": albums,
	}, w)
}