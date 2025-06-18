package isolate

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cutekitek/rankode-runner/pkg/shell"
	"github.com/pkg/errors"
)

type IsolatedBox struct {
	BoxId    int
	FilesDir string
}

type runParams struct {
	// kilobytes
	MaxFileSize int64
	Timeout     time.Duration
	// kilobytes
	MemoryLimit int64
	BindFiles   map[string]string
}

type runnableWithMeta struct {
	*shell.Command
	Meta metaFile
}

func NewIsolatedBox(boxId int) (*IsolatedBox, error) {
	cmd, err := shell.NewCommand(IsolatedExecPath, "--init", "--cg", fmt.Sprintf("-b %d", boxId))
	if err != nil {
		return nil, err
	}
	baseDir, err := cmd.RunAndCollectStdout()
	if err != nil {
		return nil, errors.Wrap(err, "failed to init isolate box")
	}
	filesDir := filepath.Join(baseDir, "box")
	return &IsolatedBox{
		BoxId:    boxId,
		FilesDir: filesDir,
	}, nil
}

func (b *IsolatedBox) Run(params runParams, command string, args ...string) (*runnableWithMeta, error) {
	isolateArgs := []string{
		"--cg",
		fmt.Sprintf("-b %d", b.BoxId),
		"-s",
	}
	for k, v := range params.BindFiles {
		isolateArgs = append(isolateArgs, fmt.Sprintf("--dir %s=%s", k, v))
	}
	metafile, err := os.CreateTemp("", "boxmeta")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a meta file")
	}
	isolateArgs = append(isolateArgs,
		"--meta="+metafile.Name(),
		fmt.Sprintf("--fsize=%d", params.MaxFileSize),
		fmt.Sprintf("--time=%f", params.Timeout.Seconds()),
		fmt.Sprintf("--cg-mem=%d", params.MemoryLimit),
		"--processes=12",
		"--run",
		"--", command)
	cmd, err := shell.NewCommand(IsolatedExecPath, append(isolateArgs, args...)...)
	if err != nil {
		return nil, err
	}
	return &runnableWithMeta{Command: cmd, Meta: metaFile{metafile}}, nil
}

func (b *IsolatedBox) CreateFile(filename string) (*os.File, error) {
	return os.Create(filepath.Join(b.FilesDir, filename))
}

func (b *IsolatedBox) Clean() {
	shell.NewCommand(IsolatedExecPath, "--cleanup", "--cg", fmt.Sprintf("-b %d", b.BoxId))
}
