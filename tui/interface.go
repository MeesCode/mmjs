// Package tui provides all means to draw and interact with the user interface.
package tui

import (
	"fmt"
	"log"
	"mmjs/audioplayer"
	"mmjs/database"
	"mmjs/globals"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// global variables
var (
	playlistFiles        = make([]globals.Track, 0)
	filelistFiles        = make([]globals.Track, 0)
	directorylistFolders = make([]globals.Folder, 0)
	songindex            = 0 // the index of the currently playing track
	myTui                tui
	changedir            func()
	search               func()
	addFolder            func()
)

// a big struct that hold all interface elements as to not occupy too much
// from the global namespace.
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
	keybinds      *tview.TextView
	mainFlex      *tview.Flex
	searchinput   *tview.InputField
	playlistinput *tview.InputField
}

// Start builds the user interface, defines the keybinds and sets initial values.
// This function will not stop until Ctrl-C is pressed, after which it will shut
// down gracefully.
func Start(mode string) {

	// build interface
	app := tview.NewApplication()
	app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		screen.Clear()
		return false
	})

	directorylist := tview.NewList().ShowSecondaryText(false)
	directorylist.SetBorder(true).SetTitle(" Directories ").SetBackgroundColor(-1)
	directorylist.SetWrapAround(false)

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

	keybinds := tview.NewTextView()
	keybinds.SetBorder(false).SetTitle(" Keybinds ").SetBackgroundColor(-1)
	keybinds.SetTextAlign(2)
	if mode == "database" {
		fmt.Fprintf(keybinds, "F2: clear | F3: search | F5: shuffle | F6: save playlist "+
			"| F7: open playlist | F8: play/pause | F9: previous | F12: next ")
	} else {
		fmt.Fprintf(keybinds, "F2: clear | F3: search | F5: shuffle "+
			" | F8: play/pause | F9: previous | F12: next ")
	}

	searchinput := tview.NewInputField().
		SetLabel("Enter a search term: ").
		SetFieldTextColor(-1).
		SetFieldBackgroundColor(-1).
		SetLabelColor(-1).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				search()
			}
			if key == tcell.KeyEsc {
				closeSearch()
			}
		})
	searchinput.SetBorder(false).SetTitle(" Search ").SetBackgroundColor(-1)

	playlistinput := tview.NewInputField().
		SetLabel("Enter name of new playlist ").
		SetFieldTextColor(-1).
		SetFieldBackgroundColor(-1).
		SetLabelColor(-1).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				savePlaylist()
			}
			if key == tcell.KeyEsc {
				closePlaylist()
			}
		})
	searchinput.SetBorder(false).SetTitle(" Playlist name ").SetBackgroundColor(-1)

	progressbar := tview.NewTextView()
	progressbar.SetBorder(false).SetBackgroundColor(-1)

	filelist := tview.NewList().ShowSecondaryText(false)
	filelist.SetBorder(true).SetTitle(" Current directory ").SetBackgroundColor(-1)
	filelist.SetWrapAround(false)
	filelist.SetChangedFunc(func(i int, _, _ string, _ rune) {
		if len(filelistFiles) > 0 {
			updateInfoBox(filelistFiles[i], browseinfobox)
		}
	})

	playlist := tview.NewList()
	playlist.SetBorder(true).SetTitle(" Playlist ").SetBackgroundColor(-1)
	playlist.ShowSecondaryText(false).SetWrapAround(false)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)

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
		keybinds:      keybinds,
		mainFlex:      mainFlex,
		searchinput:   searchinput,
		playlistinput: playlistinput,
	}

	// fill progress bar
	// time.Duration(1) is far below 1 second so it will render as 0
	drawprogressbar(time.Duration(0), time.Duration(1))

	// define tui locations
	flex := tview.NewFlex().
		AddItem(mainFlex.
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(directorylist, 0, 1, false).
				AddItem(filelist, 0, 1, false).
				AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(infoboxcontainer.
						AddItem(infobox, 0, 1, false).
						AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
							AddItem(playtime, 9, 0, false).
							AddItem(progressbar, 0, 1, false).
							AddItem(totaltime, 9, 0, false), 1, 0, false), 11, 0, false).
					AddItem(browseinfobox, 9, 0, false).
					AddItem(playlist, 0, 1, false), 0, 1, false), 0, 1, false).
			AddItem(keybinds, 1, 0, false), 0, 1, false)

	// do some stuff depending on if we are in database or filesystem mode
	// and set the root folder as the current
	var folder globals.Folder
	if mode == "filesystem" {
		addFolder = addFolderFilesystem
		changedir = changedirFilesystem
		search = searchFilesystem
		folder = globals.Folder{
			ID:       -1,
			Path:     globals.Root,
			ParentID: -1}
	} else {
		addFolder = addFolderDatabase
		changedir = changedirDatabase
		search = searchDatabase
		folder = database.GetFolderByID(1)
	}

	directorylistFolders = append(directorylistFolders, folder)
	changedir()

	// listen for audio state updates
	go audioStateUpdater()

	//////////////////////////////////////////////////////////////////////////
	// the functions below are for handling user input not defined by tview //
	//////////////////////////////////////////////////////////////////////////

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if mode == "database" {
			switch event.Key() {
			case tcell.KeyF6:
				openPlaylistInput()
				return nil
			case tcell.KeyF7:
				showPlaylists()
				return nil
			}
		}

		switch event.Key() {
		case tcell.KeyF2:
			clear()
			return nil
		case tcell.KeyF3:
			openSearch()
			return nil
		case tcell.KeyF5:
			shuffle()
			return nil
		case tcell.KeyF8:
			_, _, playing := audioplayer.GetPlaytime()
			if !playing {
				playsong()
			} else {
				audioplayer.Pause()
			}
			return nil
		case tcell.KeyF9:
			previoussong()
			return nil
		case tcell.KeyF12:
			nextsong()
			return nil
		case tcell.KeyCtrlC: // gracefull shutdown
			audioplayer.Stop()
			app.Stop()
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
		case tcell.KeyInsert:
			insertsong()
			return nil
		case tcell.KeyRight:
			app.SetFocus(playlist)
			return nil
		case tcell.KeyLeft:
			app.SetFocus(directorylist)
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
		case tcell.KeyDelete:
			deletesong()
			return nil
		case tcell.KeyRight:
			app.SetFocus(directorylist)
			return nil
		case tcell.KeyLeft:
			app.SetFocus(filelist)
			return nil
		case tcell.KeyRune:
			if event.Rune() == '-' {
				moveUp()
				return nil
			}
			if event.Rune() == '+' {
				moveDown()
				return nil
			}
		}
		return event
	})

	// directory list
	directorylist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		// alt-enter is the onlty key combination in this system
		// it's only here for legacy reasons
		if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModAlt {
			addFolder()
			return nil
		}

		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(filelist)
			return nil
		case tcell.KeyRune:
			jump(event.Rune())
			return nil
		case tcell.KeyRight:
			app.SetFocus(filelist)
			return nil
		case tcell.KeyLeft:
			app.SetFocus(playlist)
			return nil
		case tcell.KeyBackspace:
			goback()
			return nil
		case tcell.KeyBackspace2:
			goback()
			return nil
		}
		return event
	})

	// finished, draw to screen
	if err := app.SetRoot(flex, true).SetFocus(directorylist).Run(); err != nil {
		log.Fatalln("Could not start the user interface", err)
	}

}
