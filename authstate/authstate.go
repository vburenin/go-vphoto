package authstate

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sync"
)


var (
	mu = sync.Mutex{}
	httpClient *http.Client
	StoredTokenFile = "tokendata.json"

	OauthStateString = RandStringRunes(64)
	GoogleOauthConfig *oauth2.Config
	Token = &oauth2.Token{}
)

func init() {
	GoogleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/photoslibrary.readonly",
		},
		Endpoint: google.Endpoint,
	}
	s, err := os.Stat(StoredTokenFile)
	if err != nil || s.IsDir() {
		if s.IsDir() {
			panic("token file is a directory: " + StoredTokenFile)
		}
	}
	if s != nil {
		data, err := ioutil.ReadFile(StoredTokenFile)
		if err != nil {
			log.Errorf("could not read token file %s: %s", StoredTokenFile, err)
		}
		err = json.Unmarshal(data, Token)
		if err != nil {
			log.Errorf("could not decode token data %s: %s", StoredTokenFile, err)
		}
	}
	SetHttpClient(GoogleOauthConfig.Client(context.Background(), Token))
}


var letterRunes = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
func RandStringRunes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}


func GetHttpClient() *http.Client {
	mu.Lock()
	defer mu.Unlock()
	return httpClient
}

func SetHttpClient(client *http.Client) {
	mu.Lock()
	httpClient = client
	mu.Unlock()
}

func StoreToken(token *oauth2.Token) error {
	jtoken, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("could not marshal token: %s", err)
	}

	f, err := os.Create(StoredTokenFile)
	if err != nil {
		panic("could not store token data: " + err.Error())
	}

	defer f.Close()
	if _, err := f.Write(jtoken); err != nil {
		return fmt.Errorf("could not save token: %s", err)
	}

	SetHttpClient(GoogleOauthConfig.Client(context.Background(), token))

	return nil
}