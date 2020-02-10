package globals

import "time"

// channels
var Speakercommand = make(chan string)
var Speakerevent = make(chan string)
var Audiostate = make(chan Metadata)

type Metadata struct {
	Path     string
	Length   time.Duration
	Playtime time.Duration
	Finished bool
}
