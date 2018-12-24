package splace

import (
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
