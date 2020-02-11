//////////////////////////////////////////////////////////////////////////////////
// go does not suppot circular dependencies, therefore this is it's own package //
//////////////////////////////////////////////////////////////////////////////////

package globals

import "time"

// Speakercommand : media controls for the music
var Speakercommand = make(chan string)

// Audiostate : updates to metadata for showing on the tui
var Audiostate = make(chan Metadata)

// Formats : file formats supported by the program
var Formats = []string{".wav", ".mp3", ".ogg", ".weba", ".webm", ".flac"}

// Metadata : data from the audio player that is used by other components
type Metadata struct {
	Path     string
	Length   time.Duration
	Playtime time.Duration
	Finished bool
}
