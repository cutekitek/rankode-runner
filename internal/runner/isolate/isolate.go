package isolate

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Qwerty10291/rankode-runner/internal/repository/dto"
	"github.com/Qwerty10291/rankode-runner/internal/repository/models"
	"github.com/Qwerty10291/rankode-runner/internal/runner"
	"github.com/Qwerty10291/rankode-runner/pkg/files"
	"github.com/Qwerty10291/rankode-runner/pkg/shell"
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
	defer box.Clean()
	err = r.initFiles(req, box, languageConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a code file")
	}

	// build
	if err := r.build(languageConfig, box); err != nil {
		if res, ok := err.(*runFailedError); ok {
			return &dto.RunResult{
				Status: models.AttemptStatusBuildFailed,
				Error:  res.ErrorLogs,
			}, nil
		}
	}

	result, err := r.run(req, box)
	if err != nil {

		return &dto.RunResult{
			Status: models.AttemptStatusInternalError,
			Error:  "failed to run a test",
		}, nil
	}

	return result, nil
}

func (i *isolateRunner) run(req *dto.RunRequest, box *IsolatedBox) (*dto.RunResult, error) {
	params := runParams{
		MaxFileSize: int64(req.MaxFilesSize),
		Timeout:     req.Timeout,
		MemoryLimit: int64(req.MemoryLimit),
	}
	result := &dto.RunResult{
		Status: models.AttemptStatusSuccessful,
	}

	for _, input := range req.Input {
		cmd, err := box.Run(params, "runner")
		if err != nil {
			return nil, errors.Wrap(err, "failed to run test")
		}
		output, err := i.runProcessWithStdin(cmd.Command, input, int64(req.MaxOutputSize))
		meta, _ := cmd.Meta.Collect()

		result.MemoryUsage = int(meta.Memory)
		result.ExecutionTime += meta.RunTime
		caseStatus := dto.RunCaseResult{Output: output, Status: models.TestCaseStatusComplete}
		
		if err != nil {
			switch meta.Status {
			case exitStatusOutOfMemory:
				caseStatus.Status = models.TestCaseStatusOutOfMemory
			case exitStatusRuntimeError:
				caseStatus.Status = models.TestCaseStatusRunningError
			case exitStatusTimeout:
				caseStatus.Status = models.TestCaseStatusTimeout
			}

			if errors.Is(err, OutputOverflowError) {
				caseStatus.Status = models.TestCaseStatusOutputOverflow
			}
			
			result.Status = models.AttemptStatusRunFailed
			result.Output = append(result.Output, caseStatus)
			return result, nil
		}

		result.Output = append(result.Output, caseStatus)
	}
	return result, nil
}

func (i *isolateRunner) runProcessWithStdin(cmd *shell.Command, input string, maxBufferSize int64) (string, error) {
	stdinPipe, err := cmd.Cmd.StdinPipe()
	if err != nil {
		return "", errors.Wrap(err, "failed to open stdin pipe:")
	}
	stdoutPipe, err := cmd.Cmd.StdoutPipe()
	if err != nil {
		return "", errors.Wrap(err, "failed to open stdout pipe:")
	}
	errChan := make(chan error)
	var outputBuffer bytes.Buffer

	go func() {
		defer stdoutPipe.Close()
		for {
			_, err := io.CopyN(&outputBuffer, stdoutPipe, 1024)
			if err != nil {
				if err == io.EOF {
					continue
				}
				return
			}
			if outputBuffer.Len() > int(maxBufferSize) {
				errChan <- OutputOverflowError
				cmd.Cmd.Process.Kill()
			}
		}
	}()

	if err := cmd.Cmd.Start(); err != nil {
		return "", errors.Wrap(err, "failed to start runner process")
	}

	go func() {
		defer stdinPipe.Close()
		io.WriteString(stdinPipe, input)
	}()

	go func() {
		errChan <- cmd.Cmd.Wait()
	}()

	err = <-errChan
	return outputBuffer.String(), err
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
	if out, err := cmd.Cmd.Output(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &runFailedError{ErrorLogs: string(exitErr.Stderr) + string(out), StatusCode: exitErr.ExitCode()}
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
