package main

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

func startInterface() {
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

	// fill progress bar
	for i := 0; i < 200; i++ {
		fmt.Fprintf(progressbar, "%s", "▒")
	}

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

	var drawplaylist func()
	var changedir func()

	playsong := func() {
		file := tracklist[playlist.GetCurrentItem()]
		songindex = playlist.GetCurrentItem()
		drawplaylist()
		go audioplayer.Player(file)
	}

	drawplaylist = func() {
		playlist.Clear()
		for index, track := range tracklist {
			path, name := path.Split(track)
			if songindex == index {
				playlist.AddItem(name, path, '>', playsong)
			} else {
				playlist.AddItem(name, path, 0, playsong)
			}
		}
		playlist.SetCurrentItem(songindex)
		app.Draw()
	}

	nextsong := func() {
		if len(tracklist) > songindex+1 {
			songindex++
			drawplaylist()
			// println("starting: " + tracklist[songindex])
			go audioplayer.Player(tracklist[songindex])
		}
	}

	previoussong := func() {
		if songindex > 0 {
			songindex--
			drawplaylist()
			go audioplayer.Player(tracklist[songindex])
		}
	}

	addsong := func() {
		itemText, _ := filelist.GetItemText(filelist.GetCurrentItem())
		tracklist = append(tracklist, path.Join(root, itemText))
		drawplaylist()
		filelist.SetCurrentItem(filelist.GetCurrentItem() + 1)
	}

	drawprogressbar := func(playtime time.Duration, length time.Duration) {
		progressbar.Clear()
		_, _, width, _ := progressbar.GetInnerRect()
		fill := int(float64(width) * playtime.Seconds() / float64(length.Seconds()))
		for i := 0; i < fill; i++ {
			fmt.Fprintf(progressbar, "%s", "█")
		}
		for i := 0; i < width-fill; i++ {
			fmt.Fprintf(progressbar, "%s", "▒")
		}

	}

	changedir = func() {
		itemText, _ := directorylist.GetItemText(directorylist.GetCurrentItem())
		root = path.Join(root, itemText)
		directorylist.Clear()
		filelist.Clear()
		directorylist.AddItem("..", "", 0, changedir)
		files, _ := ioutil.ReadDir(root)
		for _, file := range files {
			if file.Name()[0] == '.' {
				continue
			}
			if file.IsDir() {
				directorylist.AddItem(file.Name(), "", 0, changedir)
			} else {
				if contains(formats, path.Ext(file.Name())) {
					filelist.AddItem(file.Name(), "", 0, addsong)
				}
			}
		}
	}

	directorylist.AddItem(".", "", 0, changedir)
	changedir()

	// all miscelanious globals that need to be tracked
	go func() {
		for {
			select {
			case data := <-globals.Audiostate:
				dir, name := path.Split(data.Path)
				infobox.SetCell(0, 1, tview.NewTableCell(name))
				infobox.SetCell(1, 1, tview.NewTableCell(dir))
				infobox.SetCell(2, 2, tview.NewTableCell(data.Length.String()))
				infobox.SetCell(2, 1, tview.NewTableCell(data.Playtime.String()))
				drawprogressbar(data.Playtime, data.Length)
				if data.Finished {
					nextsong()
				}
			}
			app.Draw()
		}
	}()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyF8 {
			globals.Speakercommand <- "pauze"
			return nil
		}

		// debug
		if event.Key() == tcell.KeyF7 {

			return nil
		}

		if event.Key() == tcell.KeyF9 {
			previoussong()
			return nil
		}

		if event.Key() == tcell.KeyF12 {
			nextsong()
			return nil
		}

		return event
	})

	filelist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(playlist)
			return nil
		}

		return event
	})

	playlist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(directorylist)
			return nil
		}

		return event
	})

	directorylist.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			app.SetFocus(filelist)
			return nil
		}
		return event
	})

	if err := app.SetRoot(flex, true).SetFocus(directorylist).Run(); err != nil {
		panic(err)
	}

}
