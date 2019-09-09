package formatter

import (
	"fmt"
	"go-tg-playlist-discover/internal/types"
	"strings"
)

func main() {

}

// FormatDiscoveryMessage - returns string reparesentation of Playlist Summary
func FormatDiscoveryMessage(summary *types.PlaylistTracksSummary) string {

	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("Playlist '%v' summary:\n", summary.Name))

	for i, t := range summary.Tracks {
		str := fmt.Sprintf("#%v. [%v - %v](%v)\n", i+1, t.Artist, t.Name, t.Link)
		builder.WriteString(str)
	}

	return builder.String()
}
