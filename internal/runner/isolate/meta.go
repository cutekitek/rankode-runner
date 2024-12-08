package isolate

import (
	"bufio"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Qwerty10291/rankode-runner/pkg/utils"
	"github.com/pkg/errors"
)

type exitStatus int

const (
	exitStatusOk exitStatus = iota
	exitStatusTimeout exitStatus = iota
	exitStatusOutOfMemory exitStatus = iota
	exitStatusRuntimeError exitStatus = iota
)

type metaFile struct {
	file *os.File
}

type metaData struct {
	RunTime time.Duration
	Memory int64
	StatusCode int
	Status exitStatus
}

func (m metaFile) Collect() (*metaData, error) {
	s := bufio.NewReader(m.file)
	meta := &metaData{}
	defer os.Remove(m.file.Name())
	for {
		line, err := s.ReadString('\n')
		if err == io.EOF {
			return meta, nil
		}
		line = strings.TrimSpace(line)
		args := strings.Split(line, ":")
		if len(args) != 2 {
			return nil, errors.New("invalid meta file")
		}
		switch args[0] {
		case "cg-mem":
			meta.Memory = int64(utils.MustParseInt(args[1]))
		case "exitcode":
			meta.StatusCode = utils.MustParseInt(args[1])
		case "status":
			switch args[1] {
			case "RE":
				meta.Status = exitStatusRuntimeError
			case "TO":
				meta.Status = exitStatusTimeout
			}
		case "cg-oom-killed":
			meta.Status = exitStatusOutOfMemory
		case "time":
			meta.RunTime = time.Duration(float64(time.Second) * utils.MustParseFloat64(args[1]))
		}
	}

}

