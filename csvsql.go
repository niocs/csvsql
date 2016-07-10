package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/peterh/liner"
	"github.com/niocs/sflag"
	"path/filepath"
	"encoding/hex"
	"database/sql"
	"math/rand"
	"io/ioutil"
	"strings"
	"bufio"
	"fmt"
	"log"
	"os"
	"io"
)

var opt = struct {
	Usage     string    "Usage string"
	Load      string    "CSV file to load"
	MemDB     bool      "Create DB in :memory: instead of disk. Defaults to false"
	AskType   bool      "Asks type for each field. Uses TEXT otherwise."
	TableName string    "Sqlite table name.  Default is t1.|t1"
	RPI       int       "Rows per insert. Defaults to 100. Reduce if long rows.|100"
	Query     string    "Query to run. If not provided, enters interactive mode"
	OutFile   string    "File to write csv output to. Defaults to stdout.|/dev/stdout"
	OutDelim  string    "Output Delimiter to use. Defaults is comma.|,"
	WorkDir   string    "tmp dir to create db in. Defaults to /tmp/|/tmp/"
}{}

func Usage() {
	fmt.Println(`
Usage: csvsql  --Load    <csvfile>
              [--MemDB]             #Creates sqlite db in :memory: instead of disk.
              [--AskType]           #Asks type for each field. Uses TEXT otherwise.
              [--TableName]         #Sqlite table name.  Default is t1.
              [--Query]             #Query to run. If not provided, enters interactive mode.
              [--RPI]               #Rows per insert. Defaults to 100. Reduce if long rows.
              [--OutFile]           #File to write csv output to. Defaults to stdout.
              [--OutDelim]          #Output Delimiter to use. Defaults is comma.
              [--WorkDir <workdir>] #tmp dir to create db in. Defaults to /tmp/. 
The --WorkDir parameter is ignored if --MemDB is specified.
The --OutFile parameter is ignored if --Query is NOT specified.
`)
}

func TempFileName(_basedir, prefix string) (string,string) {
	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	tmpDir, err := ioutil.TempDir(_basedir, prefix)
	if err != nil {
		log.Println(err)
		panic(err)
	}
    return tmpDir, filepath.Join(tmpDir, prefix + hex.EncodeToString(randBytes))
}

func InsertToDB(db *sql.DB,queries <-chan string, done chan<- bool) {
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		panic(err)
	}
	for {
		query, more := <-queries
		if more {
			_, err := tx.Exec(query)
			if err != nil {
				log.Println(err)
				panic(err)
			}
		} else {
			tx.Commit()
			done <- true
			return
		}
	}
}

func main() {
	sflag.Parse(&opt)
	if _,err := os.Stat(opt.Load); os.IsNotExist(err) {
		fmt.Println(err)
		Usage()
		os.Exit(1)
	}
	var dbfile      string
	var dbdir       string
	var rowStr      string
	var fieldsSql   string
	var fieldNames  string
	var query       string
	var interactiveMode = len(opt.Query) == 0
	var queryChan       = make(chan string, 500)
	var doneChan        = make(chan bool)

	if opt.MemDB {
		dbfile = ":memory:"
	} else {
		dbdir, dbfile = TempFileName(opt.WorkDir, "csvsql")
		defer os.RemoveAll(dbdir)
	}

	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer db.Close()
	
	fp, err := os.Open(opt.Load)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	fpBuf := bufio.NewReader(fp)
	rowStr, err = fpBuf.ReadString('\n')
	if err != nil {
		log.Println(err)
		panic(err)
	}
	for _, field := range(strings.Split(rowStr, ",")) {
		if len(fieldsSql) > 0 {
			fieldsSql  += ","
			fieldNames += ","
		}
		fieldsSql  += field + " TEXT"
		fieldNames += field
	}
	go InsertToDB(db, queryChan, doneChan)
	query = fmt.Sprintf("CREATE TABLE %s (%s);",opt.TableName, fieldsSql)
	queryChan <- query
	if interactiveMode {
		fmt.Println("Loading csv into table '"+ opt.TableName + "'")
	}
	insCnt := 0
	for {
		fieldsSql = ""
		rowStr, err = fpBuf.ReadString('\n')
		rowStr = strings.TrimSpace(rowStr)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(err)
			panic(err)
		}
		for _, field := range(strings.Split(rowStr, ",")) {
			if len(fieldsSql) > 0 {
				fieldsSql += ","
			}
			fieldsSql += "\"" + field + "\""
		}
		if insCnt == 0 {
			query = "INSERT INTO " + opt.TableName + " VALUES (" + fieldsSql + ")"
		} else if insCnt == opt.RPI {
			query += ";"
			queryChan <- query
			insCnt = 0
			query = "INSERT INTO " + opt.TableName + " VALUES (" + fieldsSql + ")"
		} else {
			query += ",(" + fieldsSql + ")"
		}
			insCnt++
	}
	if len(query) > 0 {
		query += ";"
		queryChan <- query
	}
	close(queryChan)
	<-doneChan
	if interactiveMode { interactive(db) } else { execQuery(db) }
}

func execQuery(db *sql.DB) {
	rows, err := db.Query(opt.Query)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Println(err)
		panic(err)
	}
	fp, err := os.Create(opt.OutFile)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	fmt.Fprintln(fp,strings.Join(cols, opt.OutDelim))
	colLen  := len(cols)
	rowData := make([]string, colLen)
	rowInt  := make([]interface{}, colLen)
	for ii := 0; ii < colLen; ii ++ { rowInt[ii] = &rowData[ii] }
	for rows.Next() {
		rows.Scan(rowInt...)
		fmt.Fprintln(fp,strings.Join(rowData, opt.OutDelim))
	}
}

func interactive(db *sql.DB) {
	line := liner.NewLiner()
	line.SetCtrlCAborts(true)

	fmt.Println("Type \\q to exit")
	for {
		query, err := line.Prompt("sql>")
		if err != nil {
			log.Println(err)
			panic(err)
		}
		if query == "\\q" {break}
		rows, err := db.Query(query)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		defer rows.Close()
		cols, err := rows.Columns()
		if err != nil {
			log.Println(err)
			panic(err)
		}
		fmt.Println(strings.Join(cols, opt.OutDelim))
		colLen  := len(cols)
		rowData := make([]string, colLen)
		rowInt  := make([]interface{}, colLen)
		for ii := 0; ii < colLen; ii ++ { rowInt[ii] = &rowData[ii] }
		for rows.Next() {
			rows.Scan(rowInt...)
			fmt.Println(strings.Join(rowData, opt.OutDelim))
		}
	}

}
