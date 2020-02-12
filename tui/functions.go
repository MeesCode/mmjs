package tui

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"mp3bak2/globals"
	"path"
	"strings"
	"time"

	"github.com/gdamore/tcell"
)

// play the song currently selected on the playlist
func playsong() {
	songindex = myTui.playlist.GetCurrentItem()
	drawplaylist()
	globals.Playfile <- tracklist[myTui.playlist.GetCurrentItem()]
	<-globals.Speakerevent
}

// draw the playlist
func drawplaylist() {
	myTui.playlist.Clear()
	for index, track := range tracklist {
		path, name := path.Split(track)
		if songindex == index {
			myTui.playlist.AddItem(name, path, '>', playsong)
		} else {
			myTui.playlist.AddItem(name, path, 0, playsong)
		}
	}
	myTui.playlist.SetCurrentItem(songindex)
	myTui.app.Draw()
}

// go to the next song (if available)
func nextsong() {
	if len(tracklist) > songindex+1 {
		songindex++
		drawplaylist()
		globals.Playfile <- tracklist[songindex]
		<-globals.Speakerevent
	}
}

// go to the previous song (if available)
func previoussong() {
	if songindex > 0 {
		songindex--
		drawplaylist()
		globals.Playfile <- tracklist[songindex]
		<-globals.Speakerevent
	}
}

// add a song to the playlist
func addsong() {
	itemText, _ := myTui.filelist.GetItemText(myTui.filelist.GetCurrentItem())
	tracklist = append(tracklist, path.Join(root, itemText))
	drawplaylist()
	myTui.filelist.SetCurrentItem(myTui.filelist.GetCurrentItem() + 1)
}

// draw the progressbar
func drawprogressbar(playtime time.Duration, length time.Duration) {
	myTui.progressbar.Clear()
	_, _, width, _ := myTui.progressbar.GetInnerRect()
	fill := int(float64(width) * playtime.Seconds() / float64(length.Seconds()))
	for i := 0; i < fill-1; i++ {
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneHLine)
	}
	fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneBlock)
	for i := 0; i < width-fill; i++ {
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneHLine)
	}

}

// navigate the file manager
func changedir() {
	itemText, _ := myTui.directorylist.GetItemText(myTui.directorylist.GetCurrentItem())
	root = path.Join(root, itemText)
	myTui.directorylist.Clear()
	myTui.filelist.Clear()
	myTui.directorylist.AddItem("..", "", 0, changedir)
	files, _ := ioutil.ReadDir(root)
	for _, file := range files {
		if file.Name()[0] == '.' {
			continue
		}
		if file.IsDir() {
			myTui.directorylist.AddItem(file.Name(), "", 0, changedir)
		} else {
			if globals.Contains(globals.Formats, strings.ToLower(path.Ext(file.Name()))) {
				myTui.filelist.AddItem(file.Name(), "", 0, addsong)
			}
		}
	}
}

// go to the previous song (if available)
func shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tracklist), func(i, j int) { tracklist[i], tracklist[j] = tracklist[j], tracklist[i] })
	songindex = 0
	globals.Playfile <- tracklist[songindex]
	<-globals.Speakerevent
	drawplaylist()
}
