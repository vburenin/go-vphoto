package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/flosch/pongo2"
	log "github.com/sirupsen/logrus"
	"github.com/vburenin/go-vphoto/authstate"
	"github.com/vburenin/go-vphoto/downloader"
	"github.com/vburenin/go-vphoto/models"
	"github.com/vburenin/go-vphoto/settings"
	"github.com/vburenin/nsync"
)

var albumLocks = nsync.NewNamedMutex()

func listAllPictures(albumID string, c *http.Client) ([]*models.MediaItem, error) {
	albumLocks.Lock(albumID)
	defer albumLocks.Unlock(albumID)

	albumPicCache := path.Join(settings.CacheLocation, albumID+".json")
	if s, _ := os.Stat(albumPicCache); s != nil && s.Size() > 0 {
		data, err := ioutil.ReadFile(albumPicCache)
		pictures := make([]*models.MediaItem, 0, 25)
		if err == nil {
			err = json.Unmarshal(data, &pictures)
			if err == nil {
				return pictures, nil
			}
		}
		log.Errorf("could not read album pictures cache %s: %s", albumID, err)
	}
	nextPageToken := ""
	baseUrl, _ := url.Parse("https://photoslibrary.googleapis.com/v1/mediaItems:search")
	pictures := make([]*models.MediaItem, 0, 25)

	params := map[string]interface{}{
		"pageSize": 100,
		"albumId":  albumID,
	}
	for ; ; {
		if nextPageToken != "" {
			params["pageToken"] = nextPageToken
		}
		data, _ := json.Marshal(params)
		resp, err := c.Post(baseUrl.String(), "application/json", bytes.NewReader(data))

		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		decodedResponse := &models.PicturesResponse{}
		err = json.Unmarshal(body, decodedResponse)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal response: %s", err)
		}
		for _, item := range decodedResponse.MediaItems {
			downloader.Downloader.AddToDownload(item)
		}
		pictures = append(pictures, decodedResponse.MediaItems...)
		log.Infof("received %d pictures", len(pictures))
		if decodedResponse.NextPageToken == "" {
			break
		}
		nextPageToken = decodedResponse.NextPageToken
	}
	data, err := json.Marshal(pictures)
	if err != nil {
		log.Errorf("failed to marshal album data %s: %s", albumID, err)
	}

	err = ioutil.WriteFile(albumPicCache, data, 0644)
	if err != nil {
		log.Error("failed to create pictures cache for album %s: %s", albumID, err)
	}

	return pictures, nil
}

func SlideShow(w http.ResponseWriter, r *http.Request) {

	albumID := r.FormValue("albumId")
	log.Debugf("Listing pictures for album: %s", albumID)

	pics, _ := listAllPictures(albumID, authstate.GetHttpClient())
	rand.Shuffle(len(pics), func(i, j int) { pics[i], pics[j] = pics[j], pics[i] })
	output := make([]string, 0, len(pics))
	for _, p := range pics {
		output = append(output, p.ID)
	}

	piclist, _ := json.Marshal(output)

	settings.PictureTemplate.ExecuteWriter(pongo2.Context{
		"allPics": string(piclist),
	}, w)

}

func LoadPicture(w http.ResponseWriter, r *http.Request) {

	picId := r.FormValue("picId")
	downloader.Downloader.GetPicNow(picId)

	http.Redirect(w, r, path.Join("/cache", picId),
		http.StatusTemporaryRedirect)

}
