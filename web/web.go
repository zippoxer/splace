package web

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

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
}

func New(opt Options) *Server {
	return &Server{opt: opt}
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
	e.Use(middleware.Logger())
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
	e.GET("/dump", s.dump)

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

	URL    string
	Secret string
}

type connectResp struct {
	Tables map[string][]string
}

func (s *Server) connect(c echo.Context) error {
	var req connectReq
	if err := c.Bind(&req); err != nil {
		return err
	}

	switch req.Driver {
	case "direct":
		u := &url.URL{
			Host: fmt.Sprintf("tcp(%s)", req.Host),
			Path: req.Database,
			User: url.UserPassword(req.User, req.Pwd),
		}
		dsn := u.String()[2:]
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return err
		}
		s.db, err = querier.NewDirect(querier.Engine(req.Engine), dsn, db)
		if err != nil {
			return err
		}
	case "php":
		qr, err := querier.NewPHP(req.URL, req.Secret)
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
		s.db.Database(),
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

func (s *Server) search(c echo.Context) error {
	var options splace.SearchOptions

	if err := json.Unmarshal([]byte(c.QueryParam("options")), &options); err != nil {
		return err
	}

	searcher := s.splace.Search(c.Request().Context(), options)
	stream := sse.Open(c.Response().Writer)

	ticker := time.NewTicker(time.Millisecond * 150)
	defer ticker.Stop()
	buff := make([][]string, 0, 1024)
	sendRows := func() {
		if len(buff) > 0 {
			stream.Send("rows", buff)
			buff = make([][]string, 0, 1024)
		}
	}
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

			for row := range result.Rows {
				buff = append(buff, row)

				// If rows buffer is full or the update interval
				// has passed, send the rows.
				if len(buff) == cap(buff) {
					sendRows()
				} else {
					select {
					case <-ticker.C:
						sendRows()
					default:
					}
				}
			}
			sendRows()
		case err := <-searcher.Done():
			stream.Send("done", struct {
				Error error
			}{
				err,
			})
			return stream.Wait()
		}
	}
}

func (s *Server) replace(c echo.Context) error {
	return nil
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
