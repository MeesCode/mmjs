package database

import (
	"fmt"
	"log"
	"github.com/MeesCode/mmjs/globals"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func displayError(errortext string){
	colorFocus := tcell.GetColor("#" + globals.Config.Highlight)

	text := tview.NewTextView()
	text.SetBorder(true).
		SetTitle(" Error ").
		SetBackgroundColor(tcell.ColorDefault).
		SetBorderColor(colorFocus)
	fmt.Fprintf(text, errortext)

	box := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(text, 5, 1, false).
			AddItem(nil, 0, 1, false), 50, 1, false).
		AddItem(nil, 0, 1, false)

	if err := tview.NewApplication().SetRoot(box, true).SetFocus(text).Run(); err != nil {
		log.Fatalln("Could not open database error display")
	}
}
