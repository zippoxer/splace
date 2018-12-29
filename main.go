package main

import (
	"flag"
	"log"
	"splace/web"

	_ "github.com/go-sql-driver/mysql"
)

var dev = flag.Bool("dev", false, "")

func main() {
	flag.Parse()

	// start := time.Now()
	// db, err := sql.Open("mysql", "root:@/quizard_web_dev")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// qr := querier.NewDirect("quizard_web_dev", querier.MySQL, db)
	// s := splace.New(qr)
	// tables, err := s.Tables(context.Background())
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // b, _ := json.MarshalIndent(tables, "", "\t")
	// // log.Println(string(b), time.Since(start))

	// // for k := range tables {
	// // 	if k != "qu_options" {
	// // 		delete(tables, k)
	// // 	}
	// // }

	// start := time.Now()
	// searcher := s.Search(context.Background(), splace.SearchOptions{
	// 	Search: "%quizard%",
	// 	Mode:   splace.Like,
	// 	Tables: tables,
	// 	Limit:  100,
	// })
	// log.Println("searching..")
	// ntables := 0
	// nrows := 0
	// stats := map[string]int{}
	// for k := range tables {
	// 	stats[k] = 0
	// }
	// for {
	// 	select {
	// 	case result := <-searcher.Results():
	// 		ntables++
	// 		// pp.Println("got result", result)
	// 		for row := range result.Rows {
	// 			_ = row
	// 			nrows++
	// 			// stats[result.Table]++
	// 			// pp.Println(row)
	// 		}
	// 	case err := <-searcher.Done():
	// 		// pp.Println(stats)
	// 		log.Fatalf("err: %v, took: %s, tables: %d, rows: %d", err, time.Since(start), ntables, nrows)
	// 	}
	// }
	// log.Fatal("done!", time.Since(start))

	app := web.New(web.Options{
		Path:  ".",
		Debug: *dev,
		Addr:  "127.0.0.1:30993",
	})
	go func() {
		log.Fatal(app.Run())
	}()
	select {}
}
