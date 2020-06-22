package handlers

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vburenin/go-vphoto/authstate"
	"io/ioutil"
	"net/http"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := authstate.GoogleOauthConfig.AuthCodeURL(authstate.OauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}


func GoogleCallBackHandler(w http.ResponseWriter, r *http.Request) {
	content, err := getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		log.Errorf("unknown response: %s", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	http.Redirect(w, r, "/albums", http.StatusTemporaryRedirect)
	fmt.Fprintf(w, "Content: %s\n", content)
}


func getUserInfo(state string, code string) ([]byte, error) {
	if state != authstate.OauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := authstate.GoogleOauthConfig.Exchange(context.Background(), code)

	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err)
	}
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err)
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err)
	}

	if err := authstate.StoreToken(token); err != nil{
		return nil, fmt.Errorf("could not store token: %s", err)
	}

	return contents, nil
}
