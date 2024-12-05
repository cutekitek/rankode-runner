package isolate

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Qwerty10291/rankode-runner/internal/repository/dto"
	"github.com/Qwerty10291/rankode-runner/internal/repository/models"
	"github.com/Qwerty10291/rankode-runner/internal/runner"
	"github.com/Qwerty10291/rankode-runner/pkg/files"
	"github.com/pkg/errors"
)

const IsolatedExecPath = "isolate"

type IsolateRunnerConfig struct {
	MaxBoxCount       int
	RunnerScriptsPath string
}

type isolateRunner struct {
	Config         IsolateRunnerConfig
	availableBoxes chan int
}

func NewIsolateRunner(cfg IsolateRunnerConfig) (runner.Runner, error) {
	boxes := make(chan int, cfg.MaxBoxCount)
	for i := 0; i < cfg.MaxBoxCount; i++ {
		boxes <- i
	}

	return &isolateRunner{
		Config:         cfg,
		availableBoxes: boxes,
	}, nil
}

func (r *isolateRunner) Run(req *dto.RunRequest) (*dto.RunResult, error) {
	languageConfig, err := r.getLangConfig(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get language config")
	}

	// wait for box
	boxId := <-r.availableBoxes
	defer func() {
		r.availableBoxes <- boxId
	}()

	box, err := NewIsolatedBox(boxId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init box")
	}
	// defer box.Clean()
	err = r.initFiles(req, box, languageConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a code file")
	}

	// build
	if err := r.build(languageConfig, box); err != nil {
		if res, ok := err.(*runFailedError); ok {
			return &dto.RunResult{
				Status: models.AttemptStatusCompilationError,
				Error:  res.ErrorLogs,
			}, nil
		}
	}

	return nil, nil
}

func (i *isolateRunner) build(cfg *languageConfig, box *IsolatedBox) error {
	params := runParams{
		MaxFileSize: int64(cfg.BuildMaxFileSize),
		Timeout:     cfg.BuildTimeout,
		MemoryLimit: int64(cfg.BuildMemoryLimit),
	}
	cmd, err := box.Run(params, "builder")
	if err != nil {
		return errors.Wrap(err, "failed to start builder process")
	}
	cmd.Cmd.Stdout = os.Stdout
	if err := cmd.Cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &runFailedError{ErrorLogs: string(exitErr.Stderr), StatusCode: exitErr.ExitCode()}
		}
	}

	return nil
}

func (i *isolateRunner) initFiles(req *dto.RunRequest, box *IsolatedBox, lang *languageConfig) error {

	codeFile, err := box.CreateFile("code")
	if err != nil {
		return errors.Wrap(err, "failed to create code file")
	}
	defer codeFile.Close()
	if _, err := codeFile.WriteString(req.Code); err != nil {
		return errors.Wrap(err, "failed to create code file")
	}

	buildScript := filepath.Join(box.FilesDir, "builder")
	if err := files.CopyFile(lang.BuildScript, buildScript); err != nil {
		return errors.Wrap(err, "failed to create a build file")
	}

	runScript := filepath.Join(box.FilesDir, "runner")
	if err := files.CopyFile(lang.RunScript, runScript); err != nil {
		return errors.Wrap(err, "failed to create a run file")
	}
	
	return nil
}

func (i *isolateRunner) getLangConfig(req *dto.RunRequest) (*languageConfig, error) {
	path := filepath.Join(i.Config.RunnerScriptsPath, req.Image)
	cfg, err := NewLangConfigFromFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("language not found")
		}
		return nil, err
	}
	return cfg, nil
}
