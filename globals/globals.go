//////////////////////////////////////////////////////////////////////////////////
// go does not suppot circular dependencies, therefore this is it's own package //
//////////////////////////////////////////////////////////////////////////////////

package globals

import "time"

// Speakercommand : media controls for the music
var Speakercommand = make(chan string)

// Audiostate : updates to metadata for showing on the tui
var Audiostate = make(chan Metadata)

// Metadata : data from the audio player that is used by other components
type Metadata struct {
	Path     string
	Length   time.Duration
	Playtime time.Duration
	Finished bool
}
