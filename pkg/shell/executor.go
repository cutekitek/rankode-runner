package shell

import (	
	"io"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type Command struct {
	Cmd *exec.Cmd
	StdIn io.WriteCloser
	StdOut io.ReadCloser
	StdErr io.ReadCloser
}

func NewCommand(command string, args ...string) (*Command, error) {
	cmd := exec.Command(command, args...)

	// stdin, err := cmd.StdinPipe()
	// if err != nil{
	// 	return nil, errors.Wrap(err, "failed to open stdin pipe")
	// }
	// stdout, err := cmd.StdoutPipe()
	// if err != nil{
	// 	return nil, errors.Wrap(err, "failed to open stdout pipe")
	// }
	stderr, err := cmd.StderrPipe()
	if err != nil{
		return nil, errors.Wrap(err, "failed to open stderr pipe")
	}
	return &Command{
		Cmd: cmd,
		// StdIn:  stdin,
		// StdOut: stdout,
		StdErr: stderr,
	},nil
}



func (c *Command) RunAndCollectStdout() (string, error) {
	data, err := c.Cmd.Output()
	return strings.TrimSpace(string(data)), err
}