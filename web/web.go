package web

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"splace/splace"
	"splace/splace/querier"

	"github.com/labstack/echo"
)

type Options struct {
	Path  string
	Debug bool
	Port  int
}

type Server struct {
	opt    Options
	db     querier.Querier
	splace *splace.Splace
}

func New(opt Options) *Server {
	return &Server{opt: opt}
}

func (s *Server) Run() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", s.opt.Port))
	if err != nil {
		return err
	}
	defer ln.Close()

	e := echo.New()
	e.Debug = s.debug()

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
		db, err := sql.Open("mysql", u.String()[2:])
		if err != nil {
			return err
		}
		s.db = querier.NewDirect(req.Database, querier.Engine(req.Engine), db)
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

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
