package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"os"
	"splace/splace"
	"splace/splace/querier"
	"splace/web"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zserge/webview"
)

var dev = flag.Bool("dev", false, "")

func main() {
	flag.Parse()

	db, err := sql.Open("mysql", "root:@/quizard_web_dev")
	if err != nil {
		log.Fatal(err)
	}
	s := splace.New(querier.NewDirect("quizard_web_dev", querier.MySQL, db))
	tables, err := s.Tables(context.Background())
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "\t")
	e.Encode(tables)
	log.Println(err)

	app := web.New(web.Options{
		Path:  ".",
		Debug: *dev,
		Port:  30993,
	})
	go func() {
		log.Fatal(app.Run())
	}()

	wv := webview.New(webview.Settings{
		Title:     "splace",
		URL:       "http://127.0.0.1:30993",
		Width:     800,
		Height:    600,
		Resizable: *dev,
		Debug:     *dev,
	})
	wv.Run()
}
