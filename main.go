package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	auth "go-tg-playlist-discover/internal/auth"

	"github.com/zmb3/spotify"
)

// PlaylistTrack -
type PlaylistTrack struct {
	id     spotify.ID
	Artist string `json:"artist"`
	Name   string `json:"name"`
}

// PlaylistTracksSummary -
type PlaylistTracksSummary struct {
	id     spotify.ID
	Name   string `json:"name"`
	Tracks []*PlaylistTrack
}

var client *spotify.Client
var user *spotify.PrivateUser

var playlistsToCheck = []string{"Discover Weekly", "Release Radar"}

func playlistMatches(pl string) bool {
	matches := false
	for _, v := range playlistsToCheck {
		if v == pl {
			matches = true
			break
		}
	}

	return matches
}

func main() {
	client = auth.GetClient()
	currentUser, err := client.CurrentUser()

	if err != nil {
		log.Fatal(err)
	}

	user = currentUser

	log.Println("You are logged in as user: ", user.ID)

	fetchPlaylistData()
}

func fetchPlaylistData() {
	playlists, err := client.GetPlaylistsForUser(user.ID)

	if err != nil {
		log.Fatal(err)
	}

	playlistsIdsToCheck := make([]spotify.ID, 0)

	for _, pl := range playlists.Playlists {
		if playlistMatches(pl.Name) {
			playlistsIdsToCheck = append(playlistsIdsToCheck, pl.ID)
		}
	}

	fetchSongsFromPlaylists(user, playlistsIdsToCheck)
}

func fetchSongsFromPlaylists(user *spotify.PrivateUser, ids []spotify.ID) {
	for _, id := range ids {

		playlist, err := client.GetPlaylist(id)

		if err != nil {
			log.Fatal(err)
		}

		newTrackssummary := &PlaylistTracksSummary{
			Name: playlist.Name,
			id:   playlist.ID,
		}
		log.Printf("Playlist '%v'\n", playlist.Name)

		for _, track := range playlist.Tracks.Tracks {

			track := &PlaylistTrack{
				id:     track.Track.ID,
				Name:   track.Track.Name,
				Artist: getDisplayArtistName(track),
			}
			newTrackssummary.Tracks = append(newTrackssummary.Tracks, track)
		}
		data, err := json.Marshal(newTrackssummary)

		os.Stdout.Write(data)
	}
}

func getDisplayArtistName(track spotify.PlaylistTrack) (artist string) {
	var artistStrBuilder strings.Builder
	separator := ","

	for _, trackArtist := range track.Track.Artists {
		artistStrBuilder.WriteString(trackArtist.Name + separator)
	}

	return strings.TrimRight(artistStrBuilder.String(), separator)
}
