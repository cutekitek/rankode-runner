package isolate

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Qwerty10291/rankode-runner/pkg/shell"
	"github.com/pkg/errors"
)

type IsolatedBox struct {
	BoxId        int
	FilesDir     string
	metadataPath string
}

type runParams struct {
	// kilobytes
	MaxFileSize int64
	Timeout     time.Duration
	// kilobytes
	MemoryLimit int64
	BindFiles   map[string]string
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
	metaPath, err := os.CreateTemp("", fmt.Sprintf("boxmeta_%d", boxId))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create meta file")
	}
	return &IsolatedBox{
		BoxId:        boxId,
		FilesDir:     filesDir,
		metadataPath: metaPath.Name(),
	}, nil
}

func (b *IsolatedBox) Run(params runParams, command string, args ...string) (*shell.Command, error) {
	isolateArgs := []string{
		"--cg",
		fmt.Sprintf("-b %d", b.BoxId),
		"-v",
		"-s",
		"--dir=/etc:noexec",
	}
	for k, v := range params.BindFiles {
		isolateArgs = append(isolateArgs, fmt.Sprintf("--dir %s=%s", k, v))
	}

	isolateArgs = append(isolateArgs,
		fmt.Sprintf("--fsize=%d", params.MaxFileSize),
		fmt.Sprintf("--time=%f", params.Timeout.Seconds()),
		fmt.Sprintf("--cg-mem=%d", params.MaxFileSize),
		"--processes=10",
		"--run",
		"--", command)
	return shell.NewCommand(IsolatedExecPath, append(isolateArgs, args...)...)
}

func (b *IsolatedBox) CreateFile(filename string) (*os.File, error) {
	return os.Create(filepath.Join(b.FilesDir, filename))
}

func (b *IsolatedBox) Clean() {
	shell.NewCommand(IsolatedExecPath, "--cleanup", "--cg", fmt.Sprintf("-b %d", b.BoxId))
}
