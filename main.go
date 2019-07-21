package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	auth "go-tg-playlist-discover/internal/auth"
	"go-tg-playlist-discover/internal/config"
	"go-tg-playlist-discover/internal/formatter"
	"go-tg-playlist-discover/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/zmb3/spotify"
)

var client *spotify.Client
var user *spotify.PrivateUser
var checkedTracks = make(map[string][]spotify.ID, 0)

var playlistsToCheck = []string{"Discover Weekly", "Release Radar"}
var discoveryLaucnhed = false
var chatID int64

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
	bot, err := tgbotapi.NewBotAPI(config.Get(config.BotAPIKey))

	if err != nil {
		log.Fatal(err)
	}

	client = auth.GetClient()
	currentUser, err := client.CurrentUser()

	if err != nil {
		log.Fatal(err)
	}

	user = currentUser
	log.Println("You are logged in as user: ", user.ID)
	checkIntervalString := config.Get(config.CheckIntervalMinutes) + "m"

	checkIntervalDuration, err := time.ParseDuration(checkIntervalString)

	if err != nil {
		log.Fatal("Error in check interval configuration", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	initDiscovery := func() {
		checkTick := time.NewTicker(checkIntervalDuration)
		checkPlaylistData(bot)

		for {
			select {
			case <-checkTick.C:
				checkPlaylistData(bot)
			}
		}
	}

	for update := range updates {
		if update.Message.Command() == "start" && !discoveryLaucnhed {
			chatID = update.Message.Chat.ID
			go initDiscovery()
			discoveryLaucnhed = true
		}
	}
}

func checkPlaylistData(bot *tgbotapi.BotAPI) {
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

	playlistSummaries := fetchSongsFromPlaylists(user, playlistsIdsToCheck)

	for _, summary := range playlistSummaries {
		notCheckedTrackIds := make([]spotify.ID, 0)
		toNotify := make([]*types.PlaylistTrack, 0)

		for _, t := range summary.Tracks {
			if trackAlreadyChecked(summary.Name, t.ID) {
				log.Println("Track " + t.Name + " already checked...")
				continue
			}

			notCheckedTrackIds = append(notCheckedTrackIds, t.ID)
		}

		if len(notCheckedTrackIds) == 0 {
			continue
		}

		tracksExist, err := client.UserHasTracks(notCheckedTrackIds...)

		if err != nil {
			fmt.Println("Error occurred checking tracks existence in user library...", err)

			continue
		}

		for i, exists := range tracksExist {
			if !exists {
				toNotify = append(toNotify, summary.Tracks[i])
			}
		}

		if len(toNotify) == 0 {
			log.Println("No new tracks discovered in playlist " + summary.Name)
			continue
		}

		notifySummary := &types.PlaylistTracksSummary{
			ID:     summary.ID,
			Name:   summary.Name,
			Tracks: toNotify,
		}

		_, err = bot.Send(tgbotapi.NewMessage(chatID, formatter.FormatDiscoveryMessage(notifySummary)))

		if err != nil {
			fmt.Println("Failed to notify on updates")
		}

		for _, v := range notifySummary.Tracks {
			markTrackAsChecked(notifySummary.Name, v.ID)
		}

	}
}

func fetchSongsFromPlaylists(user *spotify.PrivateUser, ids []spotify.ID) []*types.PlaylistTracksSummary {
	playlistsSummaries := make([]*types.PlaylistTracksSummary, 0)

	for _, id := range ids {

		playlist, err := client.GetPlaylist(id)

		if err != nil {
			log.Fatal(err)
		}

		summary := &types.PlaylistTracksSummary{
			Name:   playlist.Name,
			ID:     playlist.ID,
			Tracks: make([]*types.PlaylistTrack, 0),
		}
		log.Printf("Playlist '%v'\n", playlist.Name)

		for _, track := range playlist.Tracks.Tracks {

			track := &types.PlaylistTrack{
				ID:     track.Track.ID,
				Name:   track.Track.Name,
				Artist: getDisplayArtistName(track),
			}
			summary.Tracks = append(summary.Tracks, track)
		}

		playlistsSummaries = append(playlistsSummaries, summary)
	}

	return playlistsSummaries
}

func getDisplayArtistName(track spotify.PlaylistTrack) (artist string) {
	var artistStrBuilder strings.Builder
	separator := ","

	for _, trackArtist := range track.Track.Artists {
		artistStrBuilder.WriteString(trackArtist.Name + separator)
	}

	return strings.TrimRight(artistStrBuilder.String(), separator)
}

func trackAlreadyChecked(playlist string, id spotify.ID) bool {
	checked := false
	_, ok := checkedTracks[playlist]

	if !ok {
		checkedTracks[playlist] = make([]spotify.ID, 0)
	}

	playlistCheckedTracks := checkedTracks[playlist]

	for _, trackID := range playlistCheckedTracks {
		if id.String() == trackID.String() {
			checked = true
			break
		}
	}

	return checked
}

func markTrackAsChecked(playlist string, id spotify.ID) {
	_, ok := checkedTracks[playlist]

	if !ok {
		checkedTracks[playlist] = make([]spotify.ID, 0)
	}

	checkedTracks[playlist] = append(checkedTracks[playlist], id)
}
