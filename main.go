package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/zippoxer/splace/web"

	_ "github.com/go-sql-driver/mysql"
)

var dev = flag.Bool("dev", false, "")

func main() {
	flag.Parse()

	app := web.New(web.Options{
		Path:  ".",
		Debug: *dev,
		Addr:  "127.0.0.1:30993",
	})
	go func() {
		log.Fatal(app.Run())
	}()

	if !*dev {
		if err := openBrowser("http://localhost:30993"); err != nil {
			log.Println(err)
		}
	}

	select {}
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
