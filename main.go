package main

import (
	"flag"
	"os"

	"github.com/gdamore/tcell"
	"github.com/pgavlin/femto"
	"github.com/rivo/tview"
)

// LogWnd is a global logging window
var LogWnd *tview.TextView

// TblWnd is a global window to show schema, table info
var TblWnd *tview.TreeView

// Query buffer
var Query *femto.Buffer

// Config Struct for connection string
type Config struct {
	driver string
	host   string
	port   string
	user   string
	pass   string
	db     string
}

// Conf to keep the connection details at global level
var Conf Config = Config{driver: "mysql", host: "localhost", port: "3306", user: "port", pass: "", db: ""}

func main() {

	//Parsing command line arguments. If not passed sqlite3 with test database will open
	flag.StringVar(&Conf.driver, "d", "mysql", "mysql|sqlite")
	flag.StringVar(&Conf.host, "h", "localhost", "host to connec to")
	flag.StringVar(&Conf.port, "t", "3306", "port to connect")
	flag.StringVar(&Conf.user, "u", "root", "db user")
	flag.StringVar(&Conf.pass, "p", "", "password")
	flag.StringVar(&Conf.db, "s", "", "Default Schema")
	flag.Parse()

	if !(Conf.driver == "mysql" || Conf.driver == "sqlite") {
		flag.Usage()
		os.Exit(0)
	}

	// A buffer to keep the query from and to mini editor
	Query = femto.NewBufferFromString("", "test.sql")

	// Window to show all the tables
	TblWnd = makeTableWnd()

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
		AddItem(TblWnd, 0, 1, false).
		AddItem(innerfx, 0, 4, false)
	resWnd.SetTitle(" Result ").SetBorder(true)
	// Create App
	app := tview.NewApplication()

	// Setup Shortcuts at App Level
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlT:
			app.SetFocus(TblWnd)
			return nil
		case tcell.KeyCtrlQ:
			app.SetFocus(qryEditor)
			return nil
		case tcell.KeyCtrlL:
			app.SetFocus(LogWnd)
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
