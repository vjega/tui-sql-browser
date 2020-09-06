package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/rivo/tview"
)

func isDDL(s string) bool {
	ddl := [4]string{"USE", "CREATE", "ALTER", "DROP"}
	res := false
	for _, v := range ddl {
		res = strings.HasPrefix(strings.ToUpper(s), v)
		if res {
			return res
		}
	}
	return res
}

func opensqlitedb() *sql.DB {
	if Conf.db == "" {
		Conf.db = "test.db"
	}
	conn, err := sql.Open("sqlite3", Conf.db)
	if err != nil {
		panic(err)
	}
	return conn
}

func openmysqldb() *sql.DB {
	var dsn string
	if Conf.host == "localhost" {
		dsn = fmt.Sprintf("%s:%s@/%s", Conf.user, Conf.pass, Conf.db)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", Conf.user, Conf.pass, Conf.host, Conf.port, Conf.db)
	}
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	return conn
}

// db, err := sql.Open("mysql", "user:password@/dbname")

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

func getCurrentDB(conn *sql.DB) string {
	dbcurrdb, err := conn.Query("SELECT IFNULL(DATABASE(),'')")
	if err != nil {
		panic(err)
	}

	currdbres, _, err := getResult(dbcurrdb, 1)
	if err != nil {
		panic(err)
	}

	return currdbres[0][0]
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
	values := make([]interface{}, colcnt)
	valuePtrs := make([]interface{}, colcnt)
	columns, _ := rows.Columns()
	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		rows.Scan(valuePtrs...)
		var temp []string
		for i := range columns {
			val := values[i]

			b, ok := val.([]byte)
			var v string
			if ok {
				v = string(b)
			} else {
				v = "NULL"
			}
			temp = append(temp, v)
			//fmt.Println(col, v)
		}
		result = append(result, temp)
		rowcnt++
	}
	return result, rowcnt, nil
}

func getSQLiteTableInfo(t *tview.TreeView) {
	var conn *sql.DB
	if Conf.driver == "sqlite" {
		conn = opensqlitedb()
	}
	rows, err := conn.Query("SELECT name FROM sqlite_schema WHERE type='table';")
	if err != nil {
		panic(err)
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
		rows, err := conn.Query(fmt.Sprintf("SELECT * FROM %s LIMIT 0;", result[i][0]))
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
}

func getMySQLTableInfo(t *tview.TreeView) {

	// Run the query to get the list of databases
	var conn *sql.DB
	if Conf.driver == "mysql" {
		conn = openmysqldb()
	}

	dbrows, err := conn.Query("SHOW SCHEMAS;")
	if err != nil {
		panic(err)
	}

	dbresult, rowcnt, err := getResult(dbrows, 1)
	if err != nil {
		panic(err)
	}

	// Get the current active db if it is not set
	if Conf.db == "" {
		Conf.db = getCurrentDB(conn)
	}

	root := tview.NewTreeNode("Databases").SetSelectable(false)
	for i := 0; i < rowcnt; i++ {
		if dbresult[i][0] == "mysql" || dbresult[i][0] == "information_schema" ||
			dbresult[i][0] == "performance_schema" || dbresult[i][0] == "sys" {
			continue
		}
		var temp *tview.TreeNode
		if Conf.db == dbresult[i][0] {
			temp = tview.NewTreeNode("(*)" + dbresult[i][0])
		} else {
			temp = tview.NewTreeNode(dbresult[i][0])
		}
		temp.SetSelectable(true).SetExpanded(false).SetSelectedFunc(func() { temp.SetExpanded(!temp.IsExpanded()) })
		tblRows, err := conn.Query(fmt.Sprintf("SELECT Table_name as TablesName from information_schema.tables"+
			" WHERE table_schema = '%s';", dbresult[i][0]))
		if err != nil {
			return
		}
		tblresult, rowcnt2, err := getResult(tblRows, 1)
		if err != nil {
			return
		}
		for j := 0; j < rowcnt2; j++ {
			temp2 := tview.NewTreeNode(tblresult[j][0]).SetSelectable(true).SetExpanded(false)
			temp2.SetSelectedFunc(func() { temp2.SetExpanded(!temp2.IsExpanded()) })
			clmRows, err := conn.Query(fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT 0;", dbresult[i][0], tblresult[j][0]))
			if err != nil {
				return
			}
			_ = clmRows
			header, coltype, cnt, err := getColumnMeta(clmRows)
			if err != nil {
				return
			}
			for i := 0; i < cnt; i++ {
				temp2.AddChild(tview.NewTreeNode(fmt.Sprintf("%s %s", header[i], coltype[i])).SetSelectable(false).SetColor(tcell.ColorYellowGreen))
			}
			temp.AddChild(temp2)
		}

		// Display ad a new nodes
		root.AddChild(temp)
	}
	t.SetRoot(root)
}

func runQuery(resWind *tview.Table) {
	var conn *sql.DB

	if Conf.driver == "mysql" {
		conn = openmysqldb()
	}
	if Conf.driver == "sqlite" {
		conn = opensqlitedb()
	}

	start := time.Now()

	// Execute Query
	rows, err := conn.Query(Query.String())
	if err != nil {
		err2Log(err)
		return
	}

	if err != nil {
		panic(err)
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

	// Update table info for DDL queries
	if isDDL(Query.String()) {
		if Conf.driver == "mysql" {
			Conf.db = getCurrentDB(conn)
			getMySQLTableInfo(TblWnd)
		}
		if Conf.driver == "sqlite" {
			getSQLiteTableInfo(TblWnd)
		}
	}
	info2Log(fmt.Sprintf("Fetched %v rows", rowcnt))
	info2Log(fmt.Sprintf("Query executed in %v", end.Sub(start)))
}
