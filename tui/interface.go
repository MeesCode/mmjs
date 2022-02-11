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
	filelistFiles        = make([]globals.Track, 0)
	directorylistFolders = make([]globals.Folder, 0)
	myTui                tui
	changedir            func()
	search               func()
	searchQuery          func(string)
	addFolder            func()
)

var (
	colorFocus   = tcell.ColorBlue
	colorUnfocus = tcell.ColorWhite
)

// a big struct that hold all interface elements as to not occupy too much
// from the global namespace.
type tui struct {
	app            *tview.Application
	pages          *tview.Pages
	directorylist  *tview.List
	filelist       *tview.List
	playlist       *tview.List
	infobox        *tview.Table
	browseinfobox  *tview.Table
	progressbar    *tview.TextView
	playtime       *tview.TextView
	totaltime      *tview.TextView
	keybinds       *tview.TextView
	main           *tview.Flex
	searchbox      *tview.Flex
	searchinput    *tview.InputField
	confirmbox     *tview.Flex
	confirmfalse   *tview.Button
	confirmcontent *tview.Flex
	playlistinput  *tview.InputField
	playlistbox    *tview.Flex
	keybindstext   *tview.TextView
	keybindsbox    *tview.Flex
}

// Start builds the user interface, defines the keybinds and sets initial values.
// This function will not stop until Ctrl-C is pressed, after which it will shut
// down gracefully.
func Start() {

	log.Println("start interface")

	colorFocus = tcell.GetColor("#" + globals.Config.Highlight)

	// build interface
	app := tview.NewApplication()
	app.SetBeforeDrawFunc(func(s tcell.Screen) bool {
		s.Clear()
		return false
	})

	directorylist := tview.NewList().ShowSecondaryText(false)
	directorylist.SetBorder(true).SetTitle(" Directories ")
	directorylist.SetWrapAround(false)
	directorylist.SetBorderColor(colorFocus)
	directorylist.SetBackgroundColor(tcell.ColorDefault)

	infobox := tview.NewTable()
	infobox.SetBorder(false)
	infobox.SetBackgroundColor(tcell.ColorDefault)
	infobox.SetCell(0, 0, tview.NewTableCell("Title"))
	infobox.SetCell(1, 0, tview.NewTableCell("Artist"))
	infobox.SetCell(2, 0, tview.NewTableCell("Album"))
	infobox.SetCell(3, 0, tview.NewTableCell("Genre"))
	infobox.SetCell(4, 0, tview.NewTableCell("Year"))
	infobox.SetCell(5, 0, tview.NewTableCell("Filename"))
	infobox.SetCell(6, 0, tview.NewTableCell("Directory"))

	browseinfobox := tview.NewTable()
	browseinfobox.SetBackgroundColor(tcell.ColorDefault)
	browseinfobox.SetBorder(true).SetTitle(" Selection Info ")
	browseinfobox.SetCell(0, 0, tview.NewTableCell("Title"))
	browseinfobox.SetCell(1, 0, tview.NewTableCell("Artist"))
	browseinfobox.SetCell(2, 0, tview.NewTableCell("Album"))
	browseinfobox.SetCell(3, 0, tview.NewTableCell("Genre"))
	browseinfobox.SetCell(4, 0, tview.NewTableCell("Year"))
	browseinfobox.SetCell(5, 0, tview.NewTableCell("Filename"))
	browseinfobox.SetCell(6, 0, tview.NewTableCell("Directory"))

	infoboxcontainer := tview.NewFlex()
	infoboxcontainer.SetBackgroundColor(tcell.ColorDefault)
	infoboxcontainer.SetBorder(true).SetTitle(" Play Info ")
	infoboxcontainer.SetDirection(tview.FlexRow)

	playtime := tview.NewTextView()
	playtime.SetBackgroundColor(tcell.ColorDefault)
	playtime.SetBorder(false)

	totaltime := tview.NewTextView()
	totaltime.SetBackgroundColor(tcell.ColorDefault)
	totaltime.SetTextAlign(2)
	totaltime.SetBorder(false)

	keybinds := tview.NewTextView()
	keybinds.SetBorder(true).SetTitle(" Keybinds ")
	keybinds.SetBackgroundColor(tcell.ColorDefault)
	keybinds.SetTextAlign(1)
	fmt.Fprintf(keybinds, "F1: help | F3: search | F8: play/pause | F9: previous | F12: next ")

	searchinput := tview.NewInputField().
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				search()
			}
		})
	searchinput.SetBackgroundColor(tcell.ColorDefault)
	searchinput.SetBorder(true).SetTitle(" Search ")

	searchbox := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(searchinput, 3, 1, false).
			AddItem(nil, 0, 1, false), 60, 1, false).
		AddItem(nil, 0, 1, false)

	confirmtrue  := tview.NewButton("confirm")
	confirmfalse := tview.NewButton("cancel")

	confirmtext := tview.NewTextView()
	confirmtext.SetBackgroundColor(tcell.ColorDefault)
	fmt.Fprintf(confirmtext, "Are you sure you want to \ndelete the playlist?")

	confirmcontent := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(confirmtext, 3, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(confirmfalse, 15, 1, false).
			AddItem(nil, 0, 1, false).
			AddItem(confirmtrue, 15, 1, false), 1, 1, false)
	confirmcontent.SetBackgroundColor(tcell.ColorDefault)
	confirmcontent.SetBorder(true).SetTitle(" Confirm ")

	confirmbox := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(confirmcontent, 6, 1, false).
			AddItem(nil, 0, 1, false), 32, 1, false).
		AddItem(nil, 0, 1, false)

	keybindstext := tview.NewTextView()
	if globals.Config.Mode == "database" {
		fmt.Fprintf(keybindstext,
			`[application]
F1:  show all
F2:  clear
F3:  search
F4:  popular
F5:  shuffle
F6:  show playlists
F7:  save playlist
F8:  play/pause
F9:  previous
F10: random
F12: next

[terminal]
F11:    toggle fullscreen
Crtl-:  zoom out
Ctrl+:  zoom in
Crtl+C: close application

[playlist]
Enter:     play selected track
Delete:    remove selected track
plus (+):  move track down
minus (-): move track up

[directories]
Enter:     enter folder
Alt+Enter: add entire folder
Backspace: previous folder

[file selection]
Enter:  add track
Intert: add as next track

[contextual]
Esc: go back`)
	} else {
		fmt.Fprintf(keybindstext,
			`[application]
F1:  show all
F2:  clear
F3:  search
F5:  shuffle
F8:  play/pause
F9:  previous
F12: next

[terminal]
F11:    toggle fullscreen
Crtl-:  zoom out
Ctrl+:  zoom in
Crtl+C: close application

[playlist]
Enter:     play selected track
Delete:    remove selected track
plus (+):  move track down
minus (-): move track up

[directories]
Enter:     enter folder
Alt+Enter: add entire folder
Backspace: previous folder

[file selection]
Enter:  add track
Intert: add as next track

[contextual]
Esc: go back`)
	}

	keybindsbox := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(keybindstext, 38, 1, false).
			AddItem(nil, 0, 1, false), 50, 1, false).
		AddItem(nil, 0, 1, false)
	keybindstext.SetBackgroundColor(tcell.ColorDefault)
	keybindstext.SetBorder(true).SetTitle(" All keybindings ")

	playlistinput := tview.NewInputField().
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				savePlaylist()
			}
		})
	playlistinput.SetBackgroundColor(tcell.ColorDefault)
	playlistinput.SetBorder(true).SetTitle(" Name for new playlist ")

	playlistbox := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(playlistinput, 3, 1, false).
			AddItem(nil, 0, 1, false), 60, 1, false).
		AddItem(nil, 0, 1, false)

	progressbar := tview.NewTextView()
	progressbar.SetBorder(false)
	progressbar.SetDynamicColors(true)
	progressbar.SetBackgroundColor(tcell.ColorDefault)

	filelist := tview.NewList().ShowSecondaryText(false)
	filelist.SetBorder(true).SetTitle(" Current directory ")
	filelist.SetWrapAround(false)
	filelist.SetBackgroundColor(tcell.ColorDefault)
	filelist.SetChangedFunc(func(i int, _, _ string, _ rune) {
		if len(filelistFiles) > 0 {
			updateInfoBox(filelistFiles[i], browseinfobox)
		}
	})

	playlist := tview.NewList()
	playlist.SetBorder(true).SetTitle(" Playlist ")
	playlist.SetBackgroundColor(tcell.ColorDefault)
	playlist.ShowSecondaryText(false).SetWrapAround(false)

	main := tview.NewFlex().
		SetDirection(tview.FlexRow).
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
		AddItem(keybinds, 3, 0, false)

	pages := tview.NewPages().
		AddPage("main", main, true, true)

	// save interface
	myTui = tui{
		app:            app,
		pages:          pages,
		directorylist:  directorylist,
		filelist:       filelist,
		playlist:       playlist,
		infobox:        infobox,
		progressbar:    progressbar,
		playtime:       playtime,
		totaltime:      totaltime,
		browseinfobox:  browseinfobox,
		keybinds:       keybinds,
		main:           main,
		searchbox:      searchbox,
		searchinput:    searchinput,
		confirmbox:     confirmbox,
		confirmfalse:   confirmfalse,
		confirmcontent: confirmcontent,
		playlistbox:    playlistbox,
		playlistinput:  playlistinput,
		keybindsbox:    keybindsbox,
		keybindstext:   keybindstext,
	}

	// do some stuff depending on if we are in database or filesystem mode
	// and set the root folder as the current
	var folder globals.Folder
	if globals.Config.Mode == "filesystem" {
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

	// global input captures
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		// playlist functionallity is only available in database mode
		// for abvious reasons
		if globals.Config.Mode == "database" {
			switch event.Key() {
			case tcell.KeyF6:
				if !myTui.main.HasFocus() { return nil }
				showPlaylists()
				return nil
			case tcell.KeyF7:
				if pages.HasPage("playlist") {
					closeModals()
				} else {
					if !myTui.main.HasFocus() { return nil }
					openPlaylistInput()
				}
				return nil
			case tcell.KeyF4:
				if !myTui.main.HasFocus() { return nil }
				getPopular()
				focusWithColor(filelist)
				return nil
			case tcell.KeyF10:
				if !myTui.main.HasFocus() { return nil }
				getRandom()
				focusWithColor(filelist)
				return nil
			}
		}

		switch event.Key() {
		case tcell.KeyF1:
			if pages.HasPage("keybinds") {
				closeModals()
			} else {
				if !myTui.main.HasFocus() { return nil }
				openKeybinds()
			}
			return nil
		case tcell.KeyF2:
			if pages.HasPage("confirm") {
				closeModals()
			} else {
				if !myTui.main.HasFocus() { return nil }
				clearplaylist()
			}
			return nil
		case tcell.KeyF3:
			if pages.HasPage("search") {
				closeModals()
			} else {
				if !myTui.main.HasFocus() { return nil }
				openSearch()
			}
			return nil
		case tcell.KeyF5:
			if !myTui.main.HasFocus() { return nil }
			audioplayer.Shuffle()
			drawplaylist()
			return nil
		case tcell.KeyF8:
			// if no song is loaded, play the first song
			if len(audioplayer.Playlist) > 0 {
				if !audioplayer.WillPlay() {
					audioplayer.PlaySong(audioplayer.Songindex)
				} else {
					audioplayer.TogglePause()
				}
			}
			return nil
		case tcell.KeyF9:
			previoussong()
			return nil
		case tcell.KeyF12:
			nextsong()
			return nil
		case tcell.KeyCtrlC: // gracefull shutdown
			audioplayer.Close()
			app.Stop()
			return nil
		case tcell.KeyEsc:
			closeModals()
			return nil
		}
		return event
	})

	// file list
	filelist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyInsert:
			insertsong()
			return nil
		case tcell.KeyRight, tcell.KeyTab:
			focusWithColor(playlist)
			return nil
		case tcell.KeyLeft, tcell.KeyBacktab:
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
		case tcell.KeyLeft, tcell.KeyBacktab:
			focusWithColor(filelist)
			if myTui.filelist.GetItemCount() > 0 {
				updateInfoBox(filelistFiles[myTui.filelist.GetCurrentItem()], browseinfobox)
			}
			return nil
		case tcell.KeyRune:
			if event.Rune() == '-' {
				moveUp()
				return nil
			}
			if event.Rune() == '+' || event.Rune() == '=' {
				moveDown()
				return nil
			}
		}
		return event
	})

	// directory list
	directorylist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		// alt-enter is the only key combination in this system
		// it's only here for legacy reasons
		if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModAlt {
			addFolder()
			return nil
		}

		switch event.Key() {
		case tcell.KeyRune:
			jump(event.Rune())
			return nil
		case tcell.KeyRight, tcell.KeyTab:
			focusWithColor(filelist)
			if myTui.filelist.GetItemCount() > 0 {
				updateInfoBox(filelistFiles[myTui.filelist.GetCurrentItem()], browseinfobox)
			}
			return nil
		case tcell.KeyLeft, tcell.KeyBacktab:
			focusWithColor(playlist)
			return nil
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			goback()
			return nil
		}
		return event
	})

	// confirm buttons
	confirmfalse.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRight, tcell.KeyTab, tcell.KeyLeft, tcell.KeyBacktab:
			myTui.app.SetFocus(confirmtrue)
			return nil
		case tcell.KeyEnter:
			closeModals()
			return nil
		}
		return event
	})
	confirmtrue.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRight, tcell.KeyTab, tcell.KeyLeft, tcell.KeyBacktab:
			myTui.app.SetFocus(confirmfalse)
			return nil
		case tcell.KeyEnter:
			audioplayer.Clear()
			closeModals()
			drawplaylist()
			return nil
		}
		return event
	})

	// finished, draw to screen
	if err := app.SetRoot(pages, true).SetFocus(directorylist).Run(); err != nil {
		log.Fatalln("Could not start the user interface", err)
	}

}
