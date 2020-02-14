package tui

import (
	"fmt"
	"mp3bak2/database"
	"mp3bak2/globals"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// global variables
var (
	playlistFiles        = make([]globals.Track, 0)
	filelistFiles        = make([]globals.Track, 0)
	directorylistFolders = make([]globals.Folder, 0)
	songindex            = 0
)

type tui struct {
	app           *tview.Application
	directorylist *tview.List
	filelist      *tview.List
	playlist      *tview.List
	infobox       *tview.Table
	browseinfobox *tview.Table
	progressbar   *tview.TextView
	playtime      *tview.TextView
	totaltime     *tview.TextView
}

var myTui tui
var changedir func()

// Start : start the tui
func Start(root string, mode string) {

	// build interface
	app := tview.NewApplication()
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		screen.Clear()
		return false
	})

	directorylist := tview.NewList().ShowSecondaryText(false)
	directorylist.SetBorder(true).SetTitle(" Directories ").SetBackgroundColor(-1)

	infobox := tview.NewTable()
	infobox.SetBorder(false).SetBackgroundColor(-1)
	infobox.SetCell(0, 0, tview.NewTableCell("Title"))
	infobox.SetCell(1, 0, tview.NewTableCell("Artist"))
	infobox.SetCell(2, 0, tview.NewTableCell("Album"))
	infobox.SetCell(3, 0, tview.NewTableCell("Genre"))
	infobox.SetCell(4, 0, tview.NewTableCell("Year"))
	infobox.SetCell(5, 0, tview.NewTableCell("filename"))
	infobox.SetCell(6, 0, tview.NewTableCell("directory"))

	browseinfobox := tview.NewTable()
	browseinfobox.SetBorder(true).SetTitle(" Selection Info ").SetBackgroundColor(-1)
	browseinfobox.SetCell(0, 0, tview.NewTableCell("Title"))
	browseinfobox.SetCell(1, 0, tview.NewTableCell("Artist"))
	browseinfobox.SetCell(2, 0, tview.NewTableCell("Album"))
	browseinfobox.SetCell(3, 0, tview.NewTableCell("Genre"))
	browseinfobox.SetCell(4, 0, tview.NewTableCell("Year"))
	browseinfobox.SetCell(5, 0, tview.NewTableCell("filename"))
	browseinfobox.SetCell(6, 0, tview.NewTableCell("directory"))

	infoboxcontainer := tview.NewFlex()
	infoboxcontainer.SetBorder(true).SetTitle(" Play Info ").SetBackgroundColor(-1)
	infoboxcontainer.SetDirection(tview.FlexRow)

	playtime := tview.NewTextView()
	playtime.SetBorder(false).SetBackgroundColor(-1)

	totaltime := tview.NewTextView()
	totaltime.SetTextAlign(2)
	totaltime.SetBorder(false).SetBackgroundColor(-1)

	keybinds := tview.NewTable()
	keybinds.SetBorder(true).SetTitle(" Keybinds ").SetBackgroundColor(-1)
	keybinds.SetCell(0, 0, tview.NewTableCell("F3: search").SetExpansion(1).SetAlign(1))
	keybinds.SetCell(0, 1, tview.NewTableCell("|").SetExpansion(1).SetAlign(1))
	keybinds.SetCell(0, 2, tview.NewTableCell("F5: shuffle").SetExpansion(1).SetAlign(1))
	keybinds.SetCell(0, 3, tview.NewTableCell("|").SetExpansion(1).SetAlign(1))
	keybinds.SetCell(0, 4, tview.NewTableCell("F8: play/pause").SetExpansion(1).SetAlign(1))
	keybinds.SetCell(0, 5, tview.NewTableCell("|").SetExpansion(1).SetAlign(1))
	keybinds.SetCell(0, 6, tview.NewTableCell("F9: previous track").SetExpansion(1).SetAlign(1))
	keybinds.SetCell(0, 7, tview.NewTableCell("|").SetExpansion(1).SetAlign(1))
	keybinds.SetCell(0, 8, tview.NewTableCell("F12: next track").SetExpansion(1).SetAlign(1))

	progressbar := tview.NewTextView()
	progressbar.SetBorder(false).SetBackgroundColor(-1)

	filelist := tview.NewList().ShowSecondaryText(false)
	filelist.SetBorder(true).SetTitle(" Current directory ").SetBackgroundColor(-1)
	filelist.SetChangedFunc(func(i int, _, _ string, _ rune) {
		if len(filelistFiles) > 0 {
			updateInfoBox(filelistFiles[i], browseinfobox)
		}
	})

	playlist := tview.NewList()
	playlist.SetBorder(true).SetTitle(" Playlist ").SetBackgroundColor(-1)
	playlist.ShowSecondaryText(false)

	// save interface
	myTui = tui{
		app:           app,
		directorylist: directorylist,
		filelist:      filelist,
		playlist:      playlist,
		infobox:       infobox,
		progressbar:   progressbar,
		playtime:      playtime,
		totaltime:     totaltime,
		browseinfobox: browseinfobox,
	}

	// fill progress bar
	fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneBlock)
	for i := 0; i < 200; i++ {
		fmt.Fprintf(progressbar, "%c", tcell.RuneHLine)
	}
	fmt.Fprintf(myTui.playtime, "%s", "00:00:00")
	fmt.Fprintf(myTui.totaltime, "%s", "00:00:00")

	// define tui locations
	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(directorylist, 0, 1, false).
				AddItem(filelist, 0, 2, false).
				AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(infoboxcontainer.
						AddItem(infobox, 0, 1, false).
						AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
							AddItem(playtime, 9, 0, false).
							AddItem(progressbar, 0, 1, false).
							AddItem(totaltime, 9, 0, false), 1, 0, false), 11, 0, false).
					AddItem(browseinfobox, 9, 0, false).
					AddItem(playlist, 0, 2, false), 0, 2, false), 0, 1, false).
			AddItem(keybinds, 3, 0, false), 0, 1, false)

	// do some stuff depending on if we are in database or filesystem mode
	// and set the root folder as the current
	var folder globals.Folder
	if mode == "filesystem" {
		changedir = changedirFilesystem
		folder = globals.Folder{
			Id:       -1,
			Path:     root,
			ParentID: -1}
	} else {
		changedir = changedirDatabase
		folder = database.GetFolderByID(1)
	}

	directorylistFolders = append(directorylistFolders, folder)
	changedir()

	// listen for audio state updates
	go audioStateUpdater()

	//////////////////////////////////////////////////////////////////////////
	// the functions below are for handling user input not defined by tview //
	//////////////////////////////////////////////////////////////////////////

	// global
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF5:
			app.QueueUpdate(shuffle)
			return nil
		case tcell.KeyF8:
			globals.Speakercommand <- "pauze"
			return nil
		case tcell.KeyF7:
			globals.Speakercommand <- "change"
			return nil
		case tcell.KeyF9:
			app.QueueUpdate(previoussong)
			return nil
		case tcell.KeyF12:
			app.QueueUpdate(nextsong)
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
