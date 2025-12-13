package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"os"
	"path/filepath"
	"time"

	"github.com/criyle/go-sandbox/container"
	"github.com/criyle/go-sandbox/pkg/cgroup"
	"github.com/criyle/go-sandbox/pkg/mount"
	"github.com/criyle/go-sandbox/pkg/rlimit"
	"github.com/criyle/go-sandbox/runner"
	"github.com/cutekitek/rankode-runner/internal/repository/dto"
	"github.com/cutekitek/rankode-runner/internal/repository/models"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

var rootCG cgroup.Cgroup

func init() {
	container.Init()
	t := cgroup.DetectType()
	if t == cgroup.TypeV2 {
		cgroup.EnableV2Nesting()
	}

	ct, err := cgroup.GetAvailableController()
	if err != nil {
		panic(fmt.Sprintf("cgroup.GetAvailableController: %s", err))
	}
	rootCG, err = cgroup.New("rankode", ct)

	if err != nil {
		panic(fmt.Sprintf("cgroup.New: %s", err))
	}
}

type SandboxRunnerConfig struct {
	RunnerScriptsPath  string
	ContainersPoolSize int
}

type sandboxContainerEnv struct {
	container.Environment
	WorkDir string
}

type SandboxRunner struct {
	Config     SandboxRunnerConfig
	containers chan *sandboxContainerEnv
}

type containerRunner struct {
	container.Environment
	container.ExecveParam
}

func (r *containerRunner) Run(c context.Context) runner.Result {
	return r.Execve(c, r.ExecveParam)
}

func NewSandboxRunner(cfg SandboxRunnerConfig) (*SandboxRunner, error) {
	return &SandboxRunner{
		Config:     cfg,
		containers: make(chan *sandboxContainerEnv, cfg.ContainersPoolSize),
	}, nil
}

func (r *SandboxRunner) Init() error {
	return r.prepareContainers()
}

func (r *SandboxRunner) Close() {
	closed := 0
	for closed < r.Config.ContainersPoolSize {
		c := <-r.containers
		c.Destroy()
		os.Remove(c.WorkDir)
		closed++
	}
}

func (r *SandboxRunner) Run(req *dto.RunRequest) (*dto.RunResult, error) {
	langConfig, err := r.getLangConfig(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get language config")
	}

	container := <-r.containers
	container.Reset()
	defer func() {
		r.containers <- container
	}()

	if err := r.initFiles(req, container, langConfig); err != nil {
		return nil, errors.Wrap(err, "failed to init files")
	}

	if err := container.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping container: %w", err)
	}

	// Build
	if len(langConfig.BuildCmd) > 0 {
		if err := r.build(langConfig, container); err != nil {
			if res, ok := err.(*runFailedError); ok {
				return &dto.RunResult{
					Status: models.AttemptStatusBuildFailed,
					Error:  res.ErrorLogs + err.Error(),
				}, nil
			}
			return nil, errors.Wrap(err, "build failed")
		}
	}

	// Run test cases
	result, err := r.runTestCases(req, container, langConfig)
	if err != nil {
		return &dto.RunResult{
			Status: models.AttemptStatusInternalError,
			Error:  "failed to run a test",
		}, nil
	}

	return result, nil
}

func (r *SandboxRunner) getLangConfig(req *dto.RunRequest) (*languageConfig, error) {
	path := filepath.Join(r.Config.RunnerScriptsPath, req.Image)
	cfg, err := NewLangConfigFromFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("language not found")
		}
		return nil, err
	}
	return cfg, nil
}

func (r *SandboxRunner) initFiles(req *dto.RunRequest, env container.Environment, lang *languageConfig) error {
	codeFile := "/w/code"
	if lang.CodeFile != "" {
		codeFile = "/w/" + lang.CodeFile
	}
	
	files, err := env.Open([]container.OpenCmd{
		{Path: codeFile, Flag: os.O_WRONLY | os.O_CREATE | os.O_TRUNC, Perm: 0777},
	})
	if err != nil {
		return fmt.Errorf("failed to open files in container: %w", err)
	}
	defer func() {
		for _, f := range files {
			f.Close()
		}
	}()

	if _, err := io.Copy(files[0], strings.NewReader(req.Code)); err != nil {
		return fmt.Errorf("failed to copy code file: %w", err)
	}

	return nil
}

type runFailedError struct {
	ErrorLogs  string
	StatusCode int
}

func (r *runFailedError) Error() string {
	return fmt.Sprintf("failed to run(%d)", r.StatusCode)
}

func (r *SandboxRunner) build(cfg *languageConfig, cenv container.Environment) error {
	params := RunParams{
		ContainerEnv: cenv,
		Args:         cfg.BuildCmd,
		MaxFileSize:  int64(cfg.BuildMaxFileSize),
		Timeout:      cfg.BuildTimeout,
		MemoryLimit:  int64(cfg.BuildMemoryLimit),
	}

	res, err := r.ExecuteInSandbox(params)
	if err != nil {
		return errors.Wrap(err, "failed to execute builder")
	}

	if res.Status != runner.StatusNormal {
		return &runFailedError{ErrorLogs: string(res.Output), StatusCode: res.ExitStatus}
	}
	if res.ExitStatus != 0 {
		return &runFailedError{ErrorLogs: string(res.Output), StatusCode: res.ExitStatus}
	}

	return nil
}

func (r *SandboxRunner) runTestCases(req *dto.RunRequest, cenv container.Environment, cfg *languageConfig) (*dto.RunResult, error) {
	result := &dto.RunResult{
		Status: models.AttemptStatusSuccessful,
	}
	for _, input := range req.Input {
		params := RunParams{
			ContainerEnv:  cenv,
			Args:          cfg.RunCmd,
			MaxFileSize:   int64(req.MaxOutputSize),
			Timeout:       time.Duration(req.Timeout) * time.Millisecond,
			MemoryLimit:   int64(req.MemoryLimit),
			Input:         input,
			MaxOutputSize: int64(req.MaxOutputSize),
		}

		res, err := r.ExecuteInSandbox(params)
		if err != nil {
			return nil, errors.Wrap(err, "failed to execute runner")
		}

		caseStatus := dto.RunCaseResult{
			Output: string(res.Output),
			Status: models.TestCaseStatusComplete,
		}

		result.MemoryUsage = int(res.Memory)
		result.ExecutionTime += res.Time

		if res.Status != runner.StatusNormal {
			switch res.Status {
			case runner.StatusMemoryLimitExceeded:
				caseStatus.Status = models.TestCaseStatusOutOfMemory
			case runner.StatusTimeLimitExceeded:
				caseStatus.Status = models.TestCaseStatusTimeout
			case runner.StatusOutputLimitExceeded:
				caseStatus.Status = models.TestCaseStatusOutputOverflow
			default:
				caseStatus.Status = models.TestCaseStatusRunningError
			}
			result.Error = string(res.Error)
			result.Status = models.AttemptStatusRunFailed
			result.Output = append(result.Output, caseStatus)
			return result, nil
		}

		result.Output = append(result.Output, caseStatus)
	}

	return result, nil
}

type RunParams struct {
	ContainerEnv  container.Environment
	Args          []string
	MaxFileSize   int64
	Timeout       time.Duration
	MemoryLimit   int64
	Input         string
	MaxOutputSize int64
}

type executionResult struct {
	Status     runner.Status
	ExitStatus int
	Time       time.Duration
	Memory     runner.Size
	Error      []byte
	Output     []byte
}

func (r *SandboxRunner) ExecuteInSandbox(params RunParams) (*executionResult, error) {
	var err error
	var cg cgroup.Cgroup

	cg, err = rootCG.Random("sandbox")
	if err != nil {
		return nil, fmt.Errorf("cgroup.Random: %w", err)
	}
	defer cg.Destroy()

	if params.MemoryLimit > 0 {
		_ = cg.SetMemoryLimit(uint64(runner.Size(params.MemoryLimit)))
	}

	cgDir, err := cg.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open cg fd: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), params.Timeout)
	defer cancel()

	stdinR, stdinW, _ := os.Pipe()
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()

	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	var syncFunc func(pid int) error
	if cg != nil {
		syncFunc = func(pid int) error {
			if err := cg.AddProc(pid); err != nil {
				return err
			}
			go pipeWriter(ctx, stdinW, params.Input)
			go pipeReader(wg, ctx, cancel, stdoutR, stdout, params.MaxOutputSize)
			go pipeReader(wg, ctx, cancel, stderrR, stderr, params.MaxOutputSize)
			return nil
		}
	}

	// RLimits
	rlims := rlimit.RLimits{
		CPU:      uint64(params.Timeout.Seconds()) + 1,
		CPUHard:  uint64(params.Timeout.Seconds()) + 2,
		FileSize: uint64(params.MaxFileSize),
		Stack:    128 * 1024 * 1024,
		Data:     uint64(params.MemoryLimit),
		OpenFile: 2048,
	}

	rs := containerRunner{
		Environment: params.ContainerEnv,
		ExecveParam: container.ExecveParam{
			Args:     params.Args,
			Env:      []string{"PATH=/usr/local/bin:/usr/bin:/bin", "GOCACHE=/tmp/gocache"},
			Files:    []uintptr{stdinR.Fd(), stdoutW.Fd(), stderrW.Fd()},
			RLimits:  rlims.PrepareRLimit(),
			SyncFunc: syncFunc,
			CgroupFD: cgDir.Fd(),
		},
	}

	res := rs.Run(ctx)
	stdinR.Close()
	stdinW.Close()
	stdoutR.Close()
	stdoutW.Close()
	stderrR.Close()
	stdoutW.Close()
	wg.Wait()

	execRes := &executionResult{
		Status:     res.Status,
		ExitStatus: res.ExitStatus,
		Time:       res.Time,
		Memory:     res.Memory,
		Output:     stdout.Bytes(),
		Error:      stderr.Bytes(),
	}

	if useCGroup := (cg != nil); useCGroup {
		if cpu, err := cg.CPUUsage(); err == nil {
			execRes.Time = time.Duration(cpu)
		}
		if mem, err := cg.MemoryMaxUsage(); err == nil {
			execRes.Memory = runner.Size(mem)
		}
	}

	if stdout.Len() > int(params.MaxOutputSize) || stderr.Len() > int(params.MaxOutputSize) {
		execRes.Status = runner.StatusOutputLimitExceeded
	}

	slog.Debug("execution result", "status", execRes.Status, "exitStatus", execRes.ExitStatus, "memory", execRes.Memory, "error", res.Error, "output", execRes.Output, "stderr", execRes.Error, "time", execRes.Time)

	return execRes, nil
}

func pipeReader(wg *sync.WaitGroup, ctx context.Context, cancelF context.CancelFunc, pipe *os.File, out io.Writer, maxSize int64) {
	var copied int64
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := io.CopyN(out, pipe, 1024)
			copied += n
			if maxSize > 0 && copied >= maxSize {
				cancelF()
				return
			}
			if err != nil {
				return
			}
		}
	}
}

func pipeWriter(ctx context.Context, pipe *os.File, in string) {
	buf := make([]byte, 1024)
	reader := strings.NewReader(in)
	defer pipe.Close()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			io.CopyBuffer(pipe, reader, buf)
			if reader.Len() == 0 {
				return
			}
		}
	}
}

func (r *SandboxRunner) PrepareContainer(workdir string) (container.Environment, error) {
	mb := mount.NewBuilder().
		WithBind("/bin", "bin", true).
		WithBind("/lib", "lib", true).
		WithBind("/lib64", "lib64", true).
		WithBind("/usr", "usr", true).
		WithBind("/etc/ld.so.cache", "/etc/ld.so.cache", true).
		WithProc().
		WithBind("/dev/null", "dev/null", false).
		WithTmpfs("tmp", "size=128m,nr_inodes=4k").
		WithTmpfs("w", "size=32m,nr_inodes=4k").
		FilterNotExist()

	mounts := mb.FilterNotExist()

	cloneFlag := unix.CLONE_NEWIPC | unix.CLONE_NEWNET | unix.CLONE_NEWNS | unix.CLONE_NEWPID | unix.CLONE_NEWUSER | unix.CLONE_NEWUTS

	b := container.Builder{
		Root:          workdir,
		WorkDir:       "/w",
		Mounts:        mounts.Mounts,
		Stderr:        os.Stderr,
		CredGenerator: newCredGen(),
		CloneFlags:    uintptr(cloneFlag),
	}
	return b.Build()
}

func (r *SandboxRunner) prepareContainers() error {
	for i := 0; i < r.Config.ContainersPoolSize; i++ {
		workDir, err := os.MkdirTemp("", "rankode-container-")
		if err != nil {
			return errors.Wrap(err, "failed to create temp dir")
		}
		c, err := r.PrepareContainer(workDir)

		if err != nil {
			return errors.Wrap(err, "failed to create container")
		}
		r.containers <- &sandboxContainerEnv{
			Environment: c,
			WorkDir:     workDir,
		}
	}
	return nil
}

type credGen struct {
	cur uint32
}

func newCredGen() *credGen {
	return &credGen{cur: 10000}
}

func (c *credGen) Get() syscall.Credential {
	n := atomic.AddUint32(&c.cur, 1)
	return syscall.Credential{
		Uid: n,
		Gid: n,
	}
}
