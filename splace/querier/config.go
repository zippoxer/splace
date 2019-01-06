package querier

import (
	"github.com/go-sql-driver/mysql"
)

type Config struct {
	Engine   Engine
	Addr     string
	User     string
	Pwd      string
	Database string
}

func parseDSN(engine Engine, dataSourceName string) (*Config, error) {
	switch engine {
	case MySQL:
		c, err := mysql.ParseDSN(dataSourceName)
		if err != nil {
			return nil, err
		}
		return &Config{
			Engine:   MySQL,
			Addr:     c.Addr,
			User:     c.User,
			Pwd:      c.Passwd,
			Database: c.DBName,
		}, nil
	}
	return nil, ErrUnsupportedEngine
}

func (c Config) String() (string, error) {
	switch c.Engine {
	case MySQL:
		cfg := mysql.NewConfig()
		cfg.Addr = c.Addr
		cfg.User = c.User
		cfg.Passwd = c.Pwd
		cfg.DBName = c.Database
		return cfg.FormatDSN(), nil
	}
	return "", ErrUnsupportedEngine
}
