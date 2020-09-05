package main

import (
	"github.com/gdamore/tcell"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
)

// LogWnd is a global version of logWnd
var LogWnd *tview.TextView

// Query buffer
var Query *femto.Buffer

func main() {

	// A buffer to keep the query from and to mini editor
	Query = femto.NewBufferFromString("", "test.sql")

	// Window to show all the tables
	tblWnd := makeTableWnd()

	// Window to show all the error and info details
	LogWnd = makeLogWnd()

	// Window to display result sets from select query
	resWnd := makeResultWnd()

	// Editor primitive inside query window
	qryEditor := makeQueryWnd(resWnd)

	// All all primitives inside a flex
	innerfx := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(qryEditor, 0, 2, false).
		AddItem(resWnd, 0, 3, false).
		AddItem(LogWnd, 0, 1, false)

	// Create two flex box. 1. Table View 2. All the three boxes
	mainflex := tview.NewFlex().
		AddItem(tblWnd, 0, 1, false).
		AddItem(innerfx, 0, 4, false)
	resWnd.SetTitle(" Result ").SetBorder(true)

	// Create App
	app := tview.NewApplication()

	// Setup Shortcuts at App Level
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlT:
			app.SetFocus(tblWnd)
			return nil
		case tcell.KeyCtrlQ:
			app.SetFocus(qryEditor)
			return nil
		case tcell.KeyCtrlL:
			app.SetFocus(LogWnd)
			return nil
		case tcell.KeyF5:
			getTableInfo(tblWnd)
			return nil
		case tcell.KeyCtrlX:
			app.Stop()
			return nil
		}
		return event
	})

	//focus to Query Editor Window & Run the app.
	if err := app.SetRoot(mainflex, true).SetFocus(qryEditor).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
