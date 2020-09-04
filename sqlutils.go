package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gdamore/tcell"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rivo/tview"
)

func renderRes(resWind *tview.Table, header []string, result [][]string, rowcnt, colcnt int) {
	resWind.Clear()
	bgHeaderColor := tcell.ColorBlue
	bgCellColor := tcell.ColorDarkGray
	for c := 0; c < colcnt; c++ {
		resWind.SetCell(0, c, tview.NewTableCell(header[c]).SetBackgroundColor(bgHeaderColor))
	}
	for r := 1; r <= rowcnt; r++ {
		for c := 0; c < colcnt; c++ {
			bodyCell := tview.NewTableCell(result[r-1][c])
			resWind.SetCell(r, c, bodyCell)
			if r%2 == 0 {
				bodyCell.SetBackgroundColor(bgCellColor)
			}
		}
	}

}

func err2Log(err error) {
	LogWnd.Write([]byte(fmt.Sprintf("[red]%v {Error}: ", time.Now().Format("2006/01/02 15:04:05Z07:00"))))
	LogWnd.Write([]byte(err.Error()))
	LogWnd.Write([]byte("[white]\n"))
}

func info2Log(msg string) {
	LogWnd.Write([]byte(fmt.Sprintf("%v {Infos}: ", time.Now().Format("2006/01/02 15:04:05Z07:00"))))
	LogWnd.Write([]byte(msg))
	LogWnd.Write([]byte("\n"))
}

func getColumnMeta(rows *sql.Rows) ([]string, []string, int, error) {

	var header []string
	var coltype []string
	colcnt := 0
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, 0, err
	}

	cols, err := rows.Columns()

	if err != nil {
		return nil, nil, 0, err
	}

	for i, v := range cols {
		header = append(header, v)
		coltype = append(coltype, colTypes[i].DatabaseTypeName())
		colcnt++
	}
	return header, coltype, colcnt, nil
}

func getResult(rows *sql.Rows, colcnt int) ([][]string, int, error) {
	var result [][]string
	var rowcnt int = 0
	readCols := make([]interface{}, colcnt)
	writeCols := make([]string, colcnt)
	for i := range writeCols {
		readCols[i] = &writeCols[i]
	}
	for rows.Next() {
		var temp []string
		err := rows.Scan(readCols...)
		if err != nil {
			return nil, 0, err
		}
		for i := 0; i < colcnt; i++ {
			temp = append(temp, writeCols[i])
		}
		result = append(result, temp)
		rowcnt++
	}
	return result, rowcnt, nil
}

func getTableInfo(t *tview.TreeView) {
	database, err := sql.Open("sqlite3", "./test.db")
	rows, err := database.Query("SELECT name FROM sqlite_schema WHERE type='table';")
	if err != nil {
		return
	}

	result, rowcnt, err := getResult(rows, 1)
	if err != nil {
		return
	}
	root := tview.NewTreeNode("Tables").SetSelectable(false)
	for i := 0; i < rowcnt; i++ {
		temp := tview.NewTreeNode(result[i][0]).SetSelectable(true).SetExpanded(false)
		temp.SetSelectedFunc(func() { temp.SetExpanded(!temp.IsExpanded()) })
		rows, err := database.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 0;", result[i][0]))
		if err != nil {
			return
		}
		header, coltype, cnt, err := getColumnMeta(rows)
		if err != nil {
			return
		}
		for i := 0; i < cnt; i++ {
			temp.AddChild(tview.NewTreeNode(fmt.Sprintf("%s %s", header[i], coltype[i])).SetSelectable(false).SetColor(tcell.ColorYellowGreen))
		}

		// Display ad a new nodes
		root.AddChild(temp)
	}
	t.SetRoot(root)
	database.Close()
}

func runQuery(resWind *tview.Table) {

	// Open Database
	database, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		err2Log(err)
		return
	}

	start := time.Now()

	// Execute Query
	rows, err := database.Query(Query.String())
	if err != nil {
		err2Log(err)
		return
	}

	// Get Column names, Column Types, Column Count
	header, _, colcnt, err := getColumnMeta(rows) // column type not needed
	if err != nil {
		err2Log(err)
		return
	}

	// Get the two dimensional result set
	result, rowcnt, err := getResult(rows, colcnt)
	if err != nil {
		err2Log(err)
		return
	}
	end := time.Now()

	// Show the result in the result window
	renderRes(resWind, header, result, rowcnt, colcnt)

	info2Log(fmt.Sprintf("Fetched %v rows", rowcnt))
	info2Log(fmt.Sprintf("Query executed in %v", end.Sub(start)))

	// Finally close the db connection
	database.Close()
}
