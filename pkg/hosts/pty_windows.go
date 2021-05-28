// +build windows

package hosts

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type PtyDialer struct {
	Stdin  io.ReadCloser
	Stdout io.Writer
	Stderr io.Writer
	Writer io.Writer

	Height int
	Weight int

	ctx  context.Context
	conn *os.File
	cmd  *exec.Cmd

	err error
}

func NewPtyDialer(cmd *exec.Cmd) (*PtyDialer, error) {
	if cmd == nil {
		return nil, errors.New("[pty-dialer] no cmd is specified")
	}

	return &PtyDialer{ctx: context.Background(), cmd: cmd}, nil
}

// Close close the pty connection.
func (d *PtyDialer) Close() error {
	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// SetIO set dialer's reader and writer.
func (d *PtyDialer) SetIO(stdout, stderr io.Writer, stdin io.ReadCloser) {
	d.Stdout = stdout
	d.Stderr = stderr
	d.Stdin = stdin
}

// SetWindowSize set dialer's default win size.
func (d *PtyDialer) SetWindowSize(height, weight int) *PtyDialer {
	d.Height = height
	d.Weight = weight
	return d
}

// OpenTerminal open pty websocket terminal.
func (d *PtyDialer) OpenTerminal() error {
	return fmt.Errorf("not support windows")
}

// ChangeWindowSize changes to the current win size.
func (d *PtyDialer) ChangeWindowSize(win *WindowSize) error {
	return fmt.Errorf("not support windows")
}

// Wait waits for the command to exit.
func (d *PtyDialer) Wait() error {
	return fmt.Errorf("not support windows")
}

// Write
func (d *PtyDialer) Write(b []byte) error {
	return fmt.Errorf("not support windows")
}
