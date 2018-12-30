package querier

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"path/filepath"

	"github.com/go-sql-driver/mysql"
)

type dumpError struct {
	err    error
	output []byte
}

func (e dumpError) Error() string {
	return fmt.Sprintf("%v; %s", e.err, string(e.output))
}

func mysqldump(ctx context.Context, config *mysql.Config, w io.Writer) error {
	host, port, err := net.SplitHostPort(config.Addr)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(
		ctx,
		filepath.Join("data", "mysqldump"),
		"-h", host,
		"--port", port,
		"-u", config.User,
		"--password="+config.Passwd,
		"--column-statistics=0",
		config.DBName)
	var errBuff bytes.Buffer
	cmd.Stderr = &errBuff
	cmd.Stdout = w
	err = cmd.Run()
	if err != nil {
		return &dumpError{
			err:    err,
			output: errBuff.Bytes(),
		}
	}
	return nil
}
