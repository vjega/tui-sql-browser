package main

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/pgavlin/femto"
	"github.com/pgavlin/femto/runtime"
	"github.com/rivo/tview"
)

func makeResultWnd() *tview.Table {
	resWnd := tview.NewTable()
	return resWnd
}

func makeTableWnd() *tview.TreeView {
	res := tview.NewTreeView()
	if Conf.driver == "sqlite" {
		getSQLiteTableInfo(res)
	}
	if Conf.driver == "mysql" {
		getMySQLTableInfo(res)
	}
	res.SetTitle(" [blue:white]^T[white:black] Tables [blue:white] ").SetBorder(true)
	return res

}

func makeQueryWnd(resBox *tview.Table) *femto.View {

	var colorscheme femto.Colorscheme

	if monokai := runtime.Files.FindFile(femto.RTColorscheme, "monokai"); monokai != nil {
		if data, err := monokai.Data(); err == nil {
			colorscheme = femto.ParseColorscheme(string(data))
		}
	}

	editor := femto.NewView(Query)
	editor.SetColorscheme(colorscheme)
	editor.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlG:
			runQuery(resBox)
			return nil
		case tcell.KeyCtrlL:
			// load from disk
			return nil
		}
		return event
	})

	editor.SetBorder(true).SetTitle(fmt.Sprintf(" [blue:white]%s[white:black:] [blue:white]^Q[white:black] Query Window  [blue:white]^G[white:black] RUN QUERY [red:white]^X[white:black] EXIT", Conf.driver))

	return editor
}

func makeLogWnd() *tview.TextView {
	logWnd := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true)
	logWnd.SetBorder(true).SetTitle(" [blue:white]^L[white:black] Log Window ")
	return logWnd
}
