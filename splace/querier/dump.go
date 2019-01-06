package querier

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"path/filepath"
)

type dumpError struct {
	err    error
	output []byte
}

func (e dumpError) Error() string {
	return fmt.Sprintf("%v; %s", e.err, string(e.output))
}

func mysqldump(ctx context.Context, cfg Config, w io.Writer) error {
	host, port, err := net.SplitHostPort(cfg.Addr)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(
		ctx,
		filepath.Join("data", "mysqldump"),
		"-h", host,
		"--port", port,
		"-u", cfg.User,
		"--password="+cfg.Pwd,
		"--column-statistics=0",
		cfg.Database)
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
