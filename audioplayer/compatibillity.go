// Package audioplayer controls the audio.
package audioplayer

import (
	"log"
	"os/exec"
	"path"

	"github.com/MeesCode/mmjs/database"
	"github.com/MeesCode/mmjs/globals"
)

// convertFile takes a file and tries to convert it to 
// a file that might be compatible. Some newer mp3 versions
// and or sample rates are not supported by default.
// they will be converted to flac.
// if database mode is enabled the database will be updated 
// to reflect the new file extention 
// this function requires ffmpeg to be installed
func convertFile(track globals.Track) bool {

    cmd := exec.Command("./audioplayer/file_convert.sh", path.Clean(path.Join(globals.Root, track.Path)))
    _, err := cmd.Output()

	// command failed
	if err != nil {
		log.Println("File conversion failed", err)
		return false
	}

	// command succeeded

	// update databse
	if globals.Config.Mode == "database" {
		 // create new filename
		new_path := track.Path + ".flac"
		database.UpdatePath(new_path, track.ID)
	}

	// update file in-memory
	Playlist[Songindex].Path = Playlist[Songindex].Path + ".flac"

	return true
}