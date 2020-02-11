package tui

import (
	"fmt"
	"mp3bak2/globals"
	"os"
	"path"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// global variables
var root string
var tracklist = make([]string, 0)
var songindex = 0

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

	//save working directory
	root, _ = os.Getwd()

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
