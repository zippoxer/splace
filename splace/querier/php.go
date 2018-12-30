package querier

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

type PHP struct {
	url    string
	secret string

	info   phpInfo
	client *http.Client
}

func NewPHP(url, secret string) (*PHP, error) {
	p := &PHP{
		url:    url,
		secret: secret,
		client: &http.Client{},
	}
	resp, err := p.cmd(phpCmd{
		Cmd: "info",
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&p.info)
	return p, err
}

func (p *PHP) Exec(ctx context.Context, query string, args ...interface{}) (Result, error) {
	resp, err := p.cmd(phpCmd{
		Cmd:   "exec",
		Query: query,
		Args:  args,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(data))
	}

	var result phpResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (p *PHP) Query(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return nil, nil
}

func (p *PHP) Dump(ctx context.Context, w io.Writer) error {
	return nil
}

func (p *PHP) Engine() Engine {
	return p.info.Engine
}

func (p *PHP) Database() string {
	return p.info.Database
}

func (p *PHP) cmd(cmd phpCmd) (*http.Response, error) {
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	return p.client.Post(p.url, "application/json", bytes.NewBuffer(data))
}

type phpCmd struct {
	Cmd   string
	Query string
	Args  []interface{}
}

type phpResult struct {
	Affected int64
}

func (r phpResult) RowsAffected() (int64, error) {
	return r.Affected, nil
}

type phpInfo struct {
	Engine   Engine
	Database string
}
