package splace

import (
	"context"
	"database/sql"
)

type Mode int

const (
	Equals Mode = iota
	Like
	Regexp
)

type Splace struct {
	db *sql.DB
}

func New(db *sql.DB) *Splace {
	return &Splace{db: db}
}

func (s *Splace) Replace(ctx context.Context, opt ReplaceOptions) *Replacer {
	r := &Replacer{ctx: ctx, db: s.db, opt: opt}
	r.start()
	return r
}
