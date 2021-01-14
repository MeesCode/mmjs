// Package tui provides all means to draw and interact with the user interface.
package tui

import (
	"fmt"
	"log"
	"github.com/MeesCode/mmjs/audioplayer"
	"github.com/MeesCode/mmjs/database"
	"github.com/MeesCode/mmjs/globals"

	"github.com/gdamore/tcell/v2"
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
	searchQuery          func(string)
	addFolder            func()
)

const (
	colorFocus   = tcell.ColorMaroon
	colorUnfocus = tcell.ColorWhite
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

	directorylist := tview.NewList().ShowSecondaryText(false)
	directorylist.SetBorder(true).SetTitle(" Directories ")
	directorylist.SetWrapAround(false)
	directorylist.SetBorderColor(colorFocus)

	infobox := tview.NewTable()
	infobox.SetBorder(false)
	infobox.SetCell(0, 0, tview.NewTableCell("Title"))
	infobox.SetCell(1, 0, tview.NewTableCell("Artist"))
	infobox.SetCell(2, 0, tview.NewTableCell("Album"))
	infobox.SetCell(3, 0, tview.NewTableCell("Genre"))
	infobox.SetCell(4, 0, tview.NewTableCell("Year"))
	infobox.SetCell(5, 0, tview.NewTableCell("filename"))
	infobox.SetCell(6, 0, tview.NewTableCell("directory"))

	browseinfobox := tview.NewTable()
	browseinfobox.SetBorder(true).SetTitle(" Selection Info ")
	browseinfobox.SetCell(0, 0, tview.NewTableCell("Title"))
	browseinfobox.SetCell(1, 0, tview.NewTableCell("Artist"))
	browseinfobox.SetCell(2, 0, tview.NewTableCell("Album"))
	browseinfobox.SetCell(3, 0, tview.NewTableCell("Genre"))
	browseinfobox.SetCell(4, 0, tview.NewTableCell("Year"))
	browseinfobox.SetCell(5, 0, tview.NewTableCell("filename"))
	browseinfobox.SetCell(6, 0, tview.NewTableCell("directory"))

	infoboxcontainer := tview.NewFlex()
	infoboxcontainer.SetBorder(true).SetTitle(" Play Info ")
	infoboxcontainer.SetDirection(tview.FlexRow)

	playtime := tview.NewTextView()
	playtime.SetBorder(false)

	totaltime := tview.NewTextView()
	totaltime.SetTextAlign(2)
	totaltime.SetBorder(false)

	keybinds := tview.NewTextView()
	keybinds.SetBorder(true).SetTitle(" Keybinds ")
	keybinds.SetTextAlign(1)
	if mode == "database" {
		fmt.Fprintf(keybinds, "F2: clear | F3: search | F5: shuffle | F6: save playlist "+
			"| F7: open playlist | F8: play/pause | F9: previous | F12: next ")
	} else {
		fmt.Fprintf(keybinds, "F2: clear | F3: search | F5: shuffle "+
			" | F8: play/pause | F9: previous | F12: next ")
	}

	searchinput := tview.NewInputField().
		SetLabel("Enter a search term: ").
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				search()
			}
			if key == tcell.KeyEsc {
				closeSearch()
			}
		})
	searchinput.SetBorder(true).SetTitle(" Search ")

	playlistinput := tview.NewInputField().
		SetLabel("Enter name of new playlist: ").
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				savePlaylist()
			}
			if key == tcell.KeyEsc {
				closePlaylist()
			}
		})
	playlistinput.SetBorder(true).SetTitle(" Playlist name ")

	progressbar := tview.NewTextView()
	progressbar.SetBorder(false)
	progressbar.SetDynamicColors(true)

	filelist := tview.NewList().ShowSecondaryText(false)
	filelist.SetBorder(true).SetTitle(" Current directory ")
	filelist.SetWrapAround(false)
	filelist.SetChangedFunc(func(i int, _, _ string, _ rune) {
		if len(filelistFiles) > 0 {
			updateInfoBox(filelistFiles[i], browseinfobox)
		}
	})

	playlist := tview.NewList()
	playlist.SetBorder(true).SetTitle(" Playlist ")
	playlist.ShowSecondaryText(false).SetWrapAround(false)
	playlist.SetChangedFunc(func(i int, _, _ string, _ rune) {
		if len(playlistFiles) > 0 {
			updateInfoBox(playlistFiles[i], browseinfobox)
		}
	})

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
	fmt.Fprintf(myTui.playtime, "00:00:00")
	fmt.Fprintf(myTui.totaltime, "00:00:00")
	fmt.Fprintf(myTui.progressbar, "%s%c%s", "[crimson]", tcell.RuneBlock, "[white]")
	for i := 0; i < 200; i++ {
		fmt.Fprintf(myTui.progressbar, "%c", tcell.RuneHLine)
	}

	// define tui locations
	flex := tview.NewFlex().
		AddItem(mainFlex.
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(directorylist, 0, 1, false).
				AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(filelist, 0, 1, false).
					AddItem(browseinfobox, 9, 0, false), 0, 1, false).
				AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(infoboxcontainer.
						AddItem(infobox, 0, 1, false).
						AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
							AddItem(playtime, 9, 0, false).
							AddItem(progressbar, 0, 1, false).
							AddItem(totaltime, 9, 0, false), 1, 0, false), 11, 0, false).
					AddItem(playlist, 0, 1, false), 0, 1, false), 0, 1, false).
			AddItem(keybinds, 3, 0, false), 0, 1, false)

	// do some stuff depending on if we are in database or filesystem mode
	// and set the root folder as the current
	var folder globals.Folder
	if mode == "filesystem" {
		addFolder = addFolderFilesystem
		changedir = changedirFilesystem
		search = searchFilesystem
		searchQuery = searchFilesystemQuery
		folder = globals.Folder{
			ID:       -1,
			Path:     globals.Root,
			ParentID: -1}
	} else {
		addFolder = addFolderDatabase
		changedir = changedirDatabase
		search = searchDatabase
		searchQuery = searchDatabaseQuery
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
			focusWithColor(playlist)
			return nil
		case tcell.KeyInsert:
			insertsong()
			return nil
		case tcell.KeyRight:
			focusWithColor(playlist)
			return nil
		case tcell.KeyLeft:
			focusWithColor(directorylist)
			return nil
		}
		return event
	})

	// playlist
	playlist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			focusWithColor(directorylist)
			return nil
		case tcell.KeyDelete:
			deletesong()
			return nil
		case tcell.KeyRight:
			focusWithColor(directorylist)
			return nil
		case tcell.KeyLeft:
			focusWithColor(filelist)
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
			focusWithColor(filelist)
			return nil
		case tcell.KeyRune:
			jump(event.Rune())
			return nil
		case tcell.KeyRight:
			focusWithColor(filelist)
			return nil
		case tcell.KeyLeft:
			focusWithColor(playlist)
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
