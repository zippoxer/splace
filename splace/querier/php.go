package querier

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type cmdArgs map[string]interface{}

type PHP struct {
	url    string
	secret string
	cfg    Config

	handshake phpHandshake
	client    *http.Client
}

func NewPHP(url, secret string, cfg Config) (*PHP, error) {
	p := &PHP{
		url:    url,
		secret: secret,
		cfg:    cfg,
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	resp, err := p.cmd("handshake", cmdArgs{
		"Config": cfg,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&p.handshake); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *PHP) DiscoveredConfigs() []DiscoveredConfig {
	return p.handshake.DiscoveredConfigs
}

func (p *PHP) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	resp, err := p.cmd("exec", cmdArgs{
		"Query": query,
		"Args":  args,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result phpResult
	if err := json.NewDecoder(resp.Body).Decode(&result.affectedRows); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *PHP) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	resp, err := p.cmd("query", cmdArgs{
		"Query": query,
		"Args":  args,
	})
	if err != nil {
		return nil, err
	}

	return newPHPRows(resp.Body)
}

func (p *PHP) Dump(ctx context.Context, w io.Writer) error {
	resp, err := p.cmd("dump", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	return err
}

func (p *PHP) Config() Config {
	return p.cfg
}

func (p *PHP) Close() error {
	return nil
}

func (p *PHP) cmd(cmd string, args cmdArgs) (*http.Response, error) {
	form := args
	if form == nil {
		form = cmdArgs{}
	}
	form["Secret"] = p.secret
	form["Cmd"] = cmd
	form["Config"] = p.Config()
	data, err := json.Marshal(form)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Post(p.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if redirect := resp.Header.Get("Location"); redirect != "" {
		p.url = redirect
		return p.cmd(cmd, args)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("Can not access splace-proxy.php, try downloading & uploading it again.")
		case http.StatusNotFound:
		}

		var msg string
		if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
			return nil, err
		}
		return nil, errors.New(msg)
	}
	return resp, nil
}

type phpResult struct {
	affectedRows int64
}

func (r phpResult) RowsAffected() (int64, error) {
	return r.affectedRows, nil
}

type phpRows struct {
	columns []string
	rd      *bufio.Reader
	dec     *json.Decoder
	body    io.ReadCloser // Underlying reader kept for closing.

	err error
}

func newPHPRows(body io.ReadCloser) (*phpRows, error) {
	rd := bufio.NewReaderSize(body, 32*1024)
	dec := json.NewDecoder(rd)

	var columns []string
	if err := dec.Decode(&columns); err != nil {
		return nil, err
	}

	return &phpRows{
		columns: columns,
		rd:      rd,
		dec:     dec,
		body:    body,
	}, nil
}

func (r *phpRows) Columns() ([]string, error) {
	return r.columns, nil
}

func (r *phpRows) Next() bool {
	var msg string
	if err := r.dec.Decode(&msg); err != nil {
		r.err = err
		return false
	}
	switch msg {
	// Rows.
	case "R":
		return true

	// Error.
	case "E":
		var msg string
		if err := r.dec.Decode(&msg); err != nil {
			r.err = err
			return false
		}
		r.err = errors.New(msg)
		return false

	// Done.
	case "D":
		return false

	default:
		r.err = fmt.Errorf("invalid message '%s' received from PHP proxy", string(msg))
		return false
	}
}

func (r *phpRows) ScanStrings() ([]string, error) {
	var row []string
	err := r.dec.Decode(&row)
	return row, err
}

func (r *phpRows) Err() error {
	return r.err
}

func (r *phpRows) Close() error {
	return r.body.Close()
}

type phpHandshake struct {
	DiscoveredConfigs []DiscoveredConfig
}
