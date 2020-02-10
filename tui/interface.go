package tui

import (
	"fmt"
	"io/ioutil"
	"mp3bak2/audioplayer"
	"mp3bak2/globals"
	"path"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// global variables
var root = "/home/mees"
var tracklist = make([]string, 0)
var songindex = 0
var formats = []string{".wav", ".mp3", ".ogg", ".weba", ".webm", ".flac"}

type tui struct {
	app           *tview.Application
	directorylist *tview.List
	filelist      *tview.List
	playlist      *tview.List
	infobox       *tview.Table
	progressbar   *tview.TextView
}

var myTui tui

// Start : start the tui
func Start() {

	// build interface
	app := tview.NewApplication()

	directorylist := tview.NewList().ShowSecondaryText(false)
	directorylist.SetBorder(true).SetTitle("[ Directories ]")

	filelist := tview.NewList().ShowSecondaryText(false)
	filelist.SetBorder(true).SetTitle("[ Current directory ]")

	playlist := tview.NewList()
	playlist.SetBorder(true).SetTitle("[ Playlist ]")
	playlist.ShowSecondaryText(false)

	infobox := tview.NewTable()
	infobox.SetBorder(true).SetTitle("[ Info ]")
	infobox.SetCell(0, 0, tview.NewTableCell("filename"))
	infobox.SetCell(1, 0, tview.NewTableCell("directory"))
	infobox.SetCell(2, 0, tview.NewTableCell("playtime"))

	progressbar := tview.NewTextView()
	progressbar.SetBorder(false)

	// save interface
	myTui = tui{
		app:           app,
		directorylist: directorylist,
		filelist:      filelist,
		playlist:      playlist,
		infobox:       infobox,
		progressbar:   progressbar,
	}

	// fill progress bar
	for i := 0; i < 200; i++ {
		fmt.Fprintf(progressbar, "%s", "▒")
	}

	// define tui locations
	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(directorylist, 0, 1, false).
				AddItem(filelist, 0, 2, false).
				AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
						AddItem(infobox, 0, 1, false).
						AddItem(progressbar, 1, 0, false), 0, 1, false).
					AddItem(playlist, 0, 2, false), 0, 2, false), 0, 1, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle("[ Keybinds ]"), 3, 0, false), 0, 1, false)

	// set menu to current folser
	directorylist.AddItem(".", "", 0, changedir)
	changedir()

	// update the audio state
	go func() {
		for {
			data := <-globals.Audiostate
			dir, name := path.Split(data.Path)
			infobox.SetCell(0, 1, tview.NewTableCell(name))
			infobox.SetCell(1, 1, tview.NewTableCell(dir))
			infobox.SetCell(2, 2, tview.NewTableCell(data.Length.String()))
			infobox.SetCell(2, 1, tview.NewTableCell(data.Playtime.String()))
			drawprogressbar(data.Playtime, data.Length)
			if data.Finished {
				nextsong()
			}
			app.Draw()
		}
	}()

	//////////////////////////////////////////////////////////////////////////////////
	// the functions below are for handling user input that is not defined by tview //
	//////////////////////////////////////////////////////////////////////////////////

	// global
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF8:
			globals.Speakercommand <- "pauze"
			return nil
		case tcell.KeyF7:
			//debug
			return nil
		case tcell.KeyF9:
			previoussong()
			return nil
		case tcell.KeyF12:
			nextsong()
			return nil
		}
		return event
	})

	// file list
	filelist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(playlist)
			return nil
		}
		return event
	})

	// playlist
	playlist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(directorylist)
			return nil
		}
		return event
	})

	// directory list
	directorylist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(filelist)
			return nil
		}
		return event
	})

	// finished, draw to screen
	if err := app.SetRoot(flex, true).SetFocus(directorylist).Run(); err != nil {
		panic(err)
	}

}

// play the song currently selected on the playlist
func playsong() {
	file := tracklist[myTui.playlist.GetCurrentItem()]
	songindex = myTui.playlist.GetCurrentItem()
	drawplaylist()
	go audioplayer.Play(file)
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
		// println("starting: " + tracklist[Songindex])
		go audioplayer.Play(tracklist[songindex])
	}
}

// go to the previous song (if available)
func previoussong() {
	if songindex > 0 {
		songindex--
		drawplaylist()
		go audioplayer.Play(tracklist[songindex])
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
	for i := 0; i < fill; i++ {
		fmt.Fprintf(myTui.progressbar, "%s", "█")
	}
	for i := 0; i < width-fill; i++ {
		fmt.Fprintf(myTui.progressbar, "%s", "▒")
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
			if contains(formats, path.Ext(file.Name())) {
				myTui.filelist.AddItem(file.Name(), "", 0, addsong)
			}
		}
	}
}

// helper function to check if an array cointains a specific string
func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
