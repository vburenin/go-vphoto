package models

type AlbumResponse struct {
	Albums        []*Album `json:"albums"`
	NextPageToken string   `json:"nextPageToken"`
}

type Album struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	ProductURL            string `json:"productUrl"`
	MediaItemsCount       string `json:"mediaItemsCount"`
	CoverPhotoBaseURL     string `json:"coverPhotoBaseUrl"`
	CoverPhotoMediaItemID string `json:"coverPhotoMediaItemId"`
}
