package tui

import (
	"fmt"
	"log"
	"github.com/MeesCode/mmjs/globals"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// A function that displays an error message and keeps the main thread alive
// These errors are meant to be displayed to the user and require interaction
// The user is instructed to manually kill the program
func DisplayError(err error){
	colorFocus := tcell.GetColor("#" + globals.Config.Highlight)

	text := tview.NewTextView()
	text.SetBorder(true).
		SetTitle(" Error ").
		SetBackgroundColor(tcell.ColorDefault).
		SetBorderColor(colorFocus)
	fmt.Fprintf(text, err.Error() + "\n\npress Ctrl+C to close the application")

	box := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(text, 5, 1, false).
			AddItem(nil, 0, 1, false), 50, 1, false).
		AddItem(nil, 0, 1, false)

	box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			log.Fatalln("Application shutdown after error display")
		}

		return event
	})

	if err := tview.NewApplication().SetRoot(box, true).SetFocus(text).Run(); err != nil {
		log.Fatalln("Could not open database error display")
	}
}
