package web

import (
	"bytes"
	"compress/gzip"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/zippoxer/splace/splace"
	"github.com/zippoxer/splace/splace/querier"
	"github.com/zippoxer/splace/web/sse"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type Options struct {
	Path  string
	Debug bool
	Addr  string
}

type Server struct {
	opt Options

	addr   net.Addr
	db     querier.Querier
	splace *splace.Splace

	// Secret is a cryptographically generated random for this session.
	secret string
}

func New(opt Options) *Server {
	secret, err := uuid.NewV4()
	if err != nil {
		panic("can't generated uuid: " + err.Error())
	}
	return &Server{
		opt:    opt,
		secret: secret.String(),
	}
}

func (s *Server) Addr() net.Addr {
	return s.addr
}

func (s *Server) Run() error {
	ln, err := net.Listen("tcp", s.opt.Addr)
	if err != nil {
		return err
	}
	s.addr = ln.Addr()
	defer ln.Close()

	e := echo.New()
	e.Debug = s.debug()
	// e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8080", "http://" + s.addr.String()},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	templateFile := "web/app/dist/index.html"
	if s.debug() {
		templateFile = "web/index.html"
	}
	t := &Template{
		templates: template.Must(template.ParseFiles(filepath.Join(s.opt.Path, templateFile))),
	}
	e.Renderer = t

	e.HTTPErrorHandler = s.httpErrorHandler

	e.Static("/static", filepath.Join(s.opt.Path, "web/app/dist/static"))
	e.GET("/", s.index)
	e.POST("/connect", s.connect)
	e.GET("/search", s.search)
	e.GET("/replace", s.replace)
	e.GET("/dump", s.dump)
	e.GET("/download-php-proxy", s.downloadPhpProxy)

	return http.Serve(ln, e)
}

func (s *Server) debug() bool {
	return s.opt.Debug
}

func (s *Server) httpErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	var resp = struct {
		Error string
	}{
		err.Error(),
	}
	if err := c.JSON(code, resp); err != nil {
		c.Logger().Error(err)
	}
	c.Logger().Error(err)
}

func (s *Server) index(c echo.Context) error {
	bundleURL := "/app.js"
	if s.debug() {
		bundleURL = "http://localhost:8080/app.js"
	}
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"Dev":       s.debug(),
		"APIURL":    s.addr.String(),
		"BundleURL": bundleURL,
	})
}

type connectReq struct {
	Driver string
	Engine int

	Host     string
	Database string
	User     string
	Pwd      string

	URL string
}

type connectResp struct {
	Tables splace.TableMap
}

func (s *Server) connect(c echo.Context) error {
	var req connectReq
	if err := c.Bind(&req); err != nil {
		return err
	}

	config := querier.Config{
		Engine:   querier.Engine(req.Engine),
		Addr:     req.Host,
		Database: req.Database,
		User:     req.User,
		Pwd:      req.Pwd,
	}
	switch req.Driver {
	case "direct":
		var err error
		s.db, err = querier.NewDirect(config)
		if err != nil {
			return err
		}
	case "php":
		qr, err := querier.NewPHP(req.URL, s.secret, config)
		if err != nil {
			return err
		}
		s.db = qr
	}

	s.splace = splace.New(s.db)
	tables, err := s.splace.Tables(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, connectResp{
		Tables: tables,
	})
}

func (s *Server) dump(c echo.Context) error {
	filename := fmt.Sprintf("%s--%s.sql.gz",
		s.db.Config().Database,
		time.Now().Format("2006-02-01--15-04-05"))
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Response().Header().Set("Content-Type", "application/sql")

	w := gzip.NewWriter(c.Response().Writer)
	defer w.Close()

	err := s.db.Dump(c.Request().Context(), w)
	if err != nil {
		return err
	}

	return w.Flush()
}

func (s *Server) downloadPhpProxy(c echo.Context) error {
	data, err := ioutil.ReadFile(filepath.Join("data", "splace-proxy.php"))
	if err != nil {
		return err
	}
	rawHash := sha512.Sum512([]byte(s.secret))
	secretHash := hex.EncodeToString(rawHash[:])
	data = bytes.Replace(data,
		[]byte("<Splace Secret Hash Placeholder>"),
		[]byte("'"+secretHash+"'"),
		-1)

	c.Response().Header().Set("Content-Disposition", "attachment; filename=splace-proxy.php")
	c.Response().Header().Set("Content-Type", "application/php")
	_, err = c.Response().Write(data)
	return err
}

func (s *Server) search(c echo.Context) error {
	var options splace.SearchOptions

	if err := json.Unmarshal([]byte(c.QueryParam("options")), &options); err != nil {
		return err
	}

	searcher := s.splace.Search(c.Request().Context(), options)
	stream := sse.Open(c.Response().Writer)
	defer stream.Close()

	var wg sync.WaitGroup
	lastSendRows := time.Now()
	for {
		select {
		case result := <-searcher.Results():
			stream.Send("table", struct {
				Table string
				SQL   string
				Start time.Time
			}{
				Table: result.Table,
				SQL:   result.SQL,
				Start: result.Start,
			})

			wg.Add(1)
			go func(result splace.SearchResult) {
				defer wg.Done()

				buffLimit := 500
				buff := make([][]string, buffLimit)
				buffPos := 0
				buffSize := 0
				sendRows := func() {
					if buffPos > 0 {
						stream.Send("rows", []interface{}{
							result.Table,
							buffPos,
							buff[:buffPos],
						})
						buffPos = 0
						buffSize = 0
						lastSendRows = time.Now()
					}
				}

				rowCount := 0
				sendRowCount := func() {
					if rowCount > 0 {
						stream.Send("rows", []interface{}{
							result.Table,
							rowCount,
						})
						rowCount = 0
						lastSendRows = time.Now()
					}
				}
				for row := range result.Rows {
					if buffPos == buffLimit {
						rowCount++
						if time.Since(lastSendRows) > time.Millisecond*200 {
							sendRowCount()
						}
						continue
					}

					buff[buffPos] = row
					buffPos++

					for _, v := range row {
						buffSize += len(v)
					}

					// If rows buffer is full or the update interval
					// has passed, send the rows.
					if buffSize >= 128*1024 ||
						time.Since(lastSendRows) > time.Millisecond*200 {
						sendRows()
					}
				}
				sendRows()
				sendRowCount()
			}(result)

		case err := <-searcher.Done():
			wg.Wait()

			var msg struct {
				Error *string
			}
			if err != nil {
				s := err.Error()
				msg.Error = &s
			}
			stream.Send("done", msg)

			return stream.Close()

		case err := <-stream.Err():
			wg.Wait()
			return err
		}
	}
}

func (s *Server) replace(c echo.Context) error {
	var options splace.ReplaceOptions

	if err := json.Unmarshal([]byte(c.QueryParam("options")), &options); err != nil {
		return err
	}

	replacer := s.splace.Replace(c.Request().Context(), options)
	stream := sse.Open(c.Response().Writer)
	defer stream.Close()

	var wg sync.WaitGroup
	for {
		select {
		case result := <-replacer.Results():
			stream.Send("table", struct {
				Table string
				SQL   string
				Start time.Time
			}{
				Table: result.Table,
				SQL:   result.SQL,
				Start: result.Start,
			})

			wg.Add(1)
			go func() {
				defer wg.Done()
				for n := range result.AffectedRows {
					stream.Send("affected_rows", []interface{}{result.Table, n})
				}
			}()

		case err := <-replacer.Done():
			wg.Wait()

			var msg struct {
				Error *string
			}
			if err != nil {
				s := err.Error()
				msg.Error = &s
			}
			stream.Send("done", msg)

			return stream.Close()

		case err := <-stream.Err():
			wg.Wait()
			return err
		}
	}
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
