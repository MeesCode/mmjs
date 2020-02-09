package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// global variables
var root = "/home/mees"
var tracklist = make([]string, 0)
var songindex = 0

// channels
var speakercommand = make(chan string)
var speakerevent = make(chan string)
var timer = make(chan time.Duration)
var data = make(chan metadata)
var trackfinished = make(chan bool)

type metadata struct {
	path     string
	length   time.Duration
	playtime time.Duration
}

func main() {
	app := tview.NewApplication()

	directorylist := tview.NewList().ShowSecondaryText(false)
	directorylist.SetBorder(true).SetTitle(" Directories ")

	filelist := tview.NewList().ShowSecondaryText(false)
	filelist.SetBorder(true).SetTitle(" Current directory ")

	playlist := tview.NewList()
	playlist.SetBorder(true).SetTitle(" Playlist ")
	playlist.ShowSecondaryText(false)

	infobox := tview.NewTable()
	infobox.SetBorder(true).SetTitle(" Info ")
	infobox.SetCell(0, 0, tview.NewTableCell("filename"))
	infobox.SetCell(1, 0, tview.NewTableCell("directory"))
	infobox.SetCell(2, 0, tview.NewTableCell("playtime"))

	progresscontainer := tview.NewBox()
	progresscontainer.SetBorder(false)
	progressbar := tview.NewTextView()

	go waitforspeakercommand()

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(directorylist, 0, 1, false).
				AddItem(filelist, 0, 2, false).
				AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
						AddItem(infobox, 0, 1, false).
						AddItem(progresscontainer, 0, 0, false).
						AddItem(progressbar, 1, 0, false), 0, 1, false).
					AddItem(playlist, 0, 2, false), 0, 2, false), 0, 1, false).
			AddItem(tview.NewBox().SetBorder(true).SetTitle(" Keybinds "), 3, 0, false), 0, 1, false)

	var drawplaylist func()
	var changedir func()

	playsong := func() {
		speakercommand <- "stop"
		file := tracklist[playlist.GetCurrentItem()]
		songindex = playlist.GetCurrentItem()
		waitforspeakerevent()
		drawplaylist()
		go initspeaker(file)
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
		app.Draw()
		playlist.SetCurrentItem(songindex)
	}

	nextsong := func() {
		if len(tracklist) > songindex+1 {
			speakercommand <- "stop"
			waitforspeakerevent()
			songindex++
			drawplaylist()
			go initspeaker(tracklist[songindex])
		}
	}

	previoussong := func() {
		if songindex > 0 {
			speakercommand <- "stop"
			waitforspeakerevent()
			songindex--
			drawplaylist()
			go initspeaker(tracklist[songindex])
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
		_, _, width, _ := progresscontainer.GetInnerRect()
		fill := int(float64(width)*playtime.Seconds()/float64(length.Seconds()) - 2)
		for i := 0; i < fill; i++ {
			fmt.Fprintf(progressbar, "%s", "█")
		}
		for i := 0; i < width-fill; i++ {
			fmt.Fprintf(progressbar, "%s", "-")
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
				filelist.AddItem(file.Name(), "", 0, addsong)
			}
		}
	}

	directorylist.AddItem(".", "", 0, changedir)
	changedir()

	// all miscelanious channels that ned to be tracked
	go func() {
		for {
			select {
			case data := <-data:
				dir, name := path.Split(data.path)
				infobox.SetCell(0, 1, tview.NewTableCell(name))
				infobox.SetCell(1, 1, tview.NewTableCell(dir))
				infobox.SetCell(2, 2, tview.NewTableCell(data.length.String()))
				infobox.SetCell(2, 1, tview.NewTableCell(data.playtime.String()))
				drawprogressbar(data.playtime, data.length)
			case <-trackfinished:
				nextsong()
			}
			app.Draw()
		}
	}()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyF8 {
			speakercommand <- "pauze"
			return nil
		}

		// debug
		if event.Key() == tcell.KeyF7 {
			_, _, w, _ := progresscontainer.GetInnerRect()
			print(w)
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

func insert(a []interface{}, c interface{}, i int) []interface{} {
	return append(a[:i], append([]interface{}{c}, a[i:]...)...)
}

func waitforspeakerevent() {
	event := <-speakerevent
	for {
		if event != "stopped" {
			event = <-speakerevent
		}
		return
	}
}

func waitforspeakercommand() {
	command := <-speakercommand
	for {
		if command == "stop" {
			speakerevent <- "stopped"
			return
		}
		command = <-speakercommand
	}
}

func initspeaker(file string) {
	f, err := os.Open(file)

	if err != nil {
		println(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		println(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{Paused: false, Streamer: beep.Seq(streamer, beep.Callback(func() {
		trackfinished <- true
	}))}
	speaker.Play(ctrl)

	speaker.Lock()
	length := format.SampleRate.D(streamer.Len()).Round(time.Second)
	speaker.Unlock()

	data <- metadata{path: file, length: length, playtime: time.Duration(0)}

	for {
		select {
		case command := <-speakercommand:
			switch command {
			case "pauze":
				speaker.Lock()
				ctrl.Paused = !ctrl.Paused
				speaker.Unlock()
				break
			case "stop":
				speaker.Close()
				speakerevent <- "stopped"
				return
			}
		case <-time.After(time.Second):
			speaker.Lock()
			data <- metadata{path: file, length: length, playtime: format.SampleRate.D(streamer.Position()).Round(time.Second)}
			speaker.Unlock()
		}
	}
}
