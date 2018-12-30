package querier

import (
	"fmt"

	"github.com/go-sql-driver/mysql"
)

type dsn interface {
	Database() string
}

func parseDSN(engine Engine, dataSourceName string) (dsn, error) {
	switch engine {
	case MySQL:
		c, err := mysql.ParseDSN(dataSourceName)
		if err != nil {
			return nil, err
		}
		return mysqlDSN{c}, nil
	}
	return nil, fmt.Errorf("parseDSN: unsupported engine %s", engine)
}

type mysqlDSN struct {
	config *mysql.Config
}

func (d mysqlDSN) Database() string {
	return d.config.DBName
}
