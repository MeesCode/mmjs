//////////////////////////////////////////////////////////////////////////////////
// go does not suppot circular dependencies, therefore this is it's own package //
//////////////////////////////////////////////////////////////////////////////////

package globals

import "time"

var (
	Speakercommand = make(chan string)
	Speakerevent   = make(chan bool)
	Playfile       = make(chan string)
	Audiostate     = make(chan Metadata)
	Formats        = []string{".wav", ".mp3", ".ogg", ".flac"}
)

// Metadata : data from the audio player that is used by other components
type Metadata struct {
	Path     string
	Length   time.Duration
	Playtime time.Duration
	Finished bool
}

// helper function to check if an array cointains a specific string
func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
