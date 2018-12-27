package main

import (
	"flag"
	"log"
	"splace/web"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zserge/webview"
)

var dev = flag.Bool("dev", false, "")

func main() {
	flag.Parse()

	app := web.New(web.Options{
		Path:  ".",
		Debug: *dev,
		Port:  30993,
	})
	go func() {
		log.Fatal(app.Run())
	}()
	// select {}

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
