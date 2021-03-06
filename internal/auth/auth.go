package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"go-tg-playlist-discover/internal/config"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var redirectURI = config.Get(config.HOST) + "/callback"
var clientChannel = make(chan *spotify.Client, 1)
var authenticator = spotify.NewAuthenticator(
	redirectURI,
	spotify.ScopeUserLibraryRead,
	spotify.ScopePlaylistModifyPrivate,
	spotify.ScopePlaylistReadPrivate,
)
var client *spotify.Client
var user *spotify.PrivateUser
var state = config.Get(config.STATE)
var credentialsFilePath = config.Get(config.CredentialsPath)

// GetClient - returns spotify client
func GetClient() *spotify.Client {
	http.HandleFunc("/playlist-discover/callback", completeAuth)

	http.HandleFunc("/playlist-discover/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("PongPong!!!"))
	})

	go http.ListenAndServe(":"+config.Get(config.PORT), nil)
	go initClient()

	return <-clientChannel
}

func initClient() {
	credsFile, err := os.Open(credentialsFilePath)

	if err != nil {
		log.Println("no credentials file found, pending callback auth...")
		url := authenticator.AuthURL(state)

		log.Println("Please visit following URL: ", url)
		return
	}

	credsBytes, _ := ioutil.ReadAll(credsFile)

	token := new(oauth2.Token)
	decodeErr := json.Unmarshal(credsBytes, token)

	if decodeErr != nil {
		log.Fatal(err)
	}

	client := spotify.NewAuthenticator("").NewClient(token)

	newToken, _ := client.Token()

	if newToken.AccessToken != token.AccessToken {
		fmt.Println("updating token data...")
		writeCreds(newToken)
	}

	fmt.Println("Login completed")

	clientChannel <- &client
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	token, err := authenticator.Token(state, r)

	if err != nil {
		http.Error(w, "Could not get token", http.StatusForbidden)
	}

	if responseState := r.FormValue("state"); responseState != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", responseState, state)
	}

	writeCreds(token)
	client := authenticator.NewClient(token)

	fmt.Println(w, "Login completed")

	clientChannel <- &client

}

func writeCreds(token *oauth2.Token) {
	f, err := os.Create(credentialsFilePath)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	b, err := json.Marshal(token)

	if err != nil {
		log.Fatal(err)
	}

	f.Write(b)
}
