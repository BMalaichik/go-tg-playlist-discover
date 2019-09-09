package types

import "github.com/zmb3/spotify"

// PlaylistTrack -
type PlaylistTrack struct {
	ID     spotify.ID
	Artist string `json:"artist"`
	Name   string `json:"name"`
	Link   string `json:"uri"`
}

// PlaylistTracksSummary -
type PlaylistTracksSummary struct {
	ID     spotify.ID
	Name   string `json:"name"`
	Tracks []*PlaylistTrack
}

func main() {

}
