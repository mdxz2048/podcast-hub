package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var ErrRunnerDisabled = errors.New("runner mode disabled")

type DockerMount struct {
	Source   string
	Target   string
	ReadOnly bool
}

type DockerRunSpec struct {
	Image          string
	Command        []string
	ContainerName  string
	Privileged     bool
	NetworkMode    string
	User           string
	ReadOnlyRootFS bool
	Mounts         []DockerMount
	WritableTarget string
	CPUQuota       int64
	MemoryBytes    int64
	PidsLimit      int64
	Timeout        time.Duration
	StopGrace      time.Duration
}

type DockerTrustedAdminConfig struct {
	Image                string
	ConnectorPackagePath string
	Timeout              time.Duration
	StopGrace            time.Duration
	CPUQuota             int64
	MemoryBytes          int64
	PidsLimit            int64
	User                 string
}

type DockerClient interface {
	Run(ctx context.Context, spec DockerRunSpec) (stdout string, exitCode int, err error)
}

type DockerTrustedAdminExecutor struct {
	Client DockerClient
	Config DockerTrustedAdminConfig
}

func (e DockerTrustedAdminExecutor) Execute(ctx context.Context, inputPath string, outputDir string) ExecutionResult {
	spec, err := BuildDockerTrustedAdminSpec(inputPath, outputDir, e.Config)
	if err != nil {
		return ExecutionResult{ExitCode: 1, Err: err}
	}
	client := e.Client
	if client == nil {
		client = DockerCLIClient{}
	}
	stdout, exitCode, err := client.Run(ctx, spec)
	return ExecutionResult{Stdout: strings.NewReader(stdout), ExitCode: exitCode, Err: err}
}

func BuildDockerTrustedAdminSpec(inputPath string, outputDir string, cfg DockerTrustedAdminConfig) (DockerRunSpec, error) {
	if strings.TrimSpace(cfg.Image) == "" {
		return DockerRunSpec{}, errors.New("RUNNER_PYTHON_BASIC_IMAGE is required")
	}
	if strings.TrimSpace(cfg.ConnectorPackagePath) == "" {
		return DockerRunSpec{}, errors.New("connector package path is required")
	}
	workRoot := filepath.Clean(filepath.Join(outputDir, ".."))
	if filepath.Base(workRoot) != "work" {
		return DockerRunSpec{}, errors.New("output directory must be under work/output")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Minute
	}
	stopGrace := cfg.StopGrace
	if stopGrace <= 0 {
		stopGrace = 5 * time.Second
	}
	cpuQuota := cfg.CPUQuota
	if cpuQuota <= 0 {
		cpuQuota = 100000
	}
	memory := cfg.MemoryBytes
	if memory <= 0 {
		memory = 256 * 1024 * 1024
	}
	pids := cfg.PidsLimit
	if pids <= 0 {
		pids = 128
	}
	user := strings.TrimSpace(cfg.User)
	if user == "" {
		uid := os.Getuid()
		gid := os.Getgid()
		if uid == 0 {
			uid = 65532
			gid = 65532
		}
		user = strconv.Itoa(uid) + ":" + strconv.Itoa(gid)
	}
	return DockerRunSpec{
		Image:          cfg.Image,
		Command:        []string{"python", "/connector/fixture_connector.py", "/work/input/" + filepath.Base(inputPath), "/work/output"},
		ContainerName:  "podcast-hub-job-" + safeContainerSuffix(workRoot),
		Privileged:     false,
		NetworkMode:    "none",
		User:           user,
		ReadOnlyRootFS: true,
		Mounts: []DockerMount{
			{Source: workRoot, Target: "/work", ReadOnly: false},
			{Source: filepath.Clean(cfg.ConnectorPackagePath), Target: "/connector", ReadOnly: true},
		},
		WritableTarget: "/work",
		CPUQuota:       cpuQuota,
		MemoryBytes:    memory,
		PidsLimit:      pids,
		Timeout:        timeout,
		StopGrace:      stopGrace,
	}, nil
}

func safeContainerSuffix(value string) string {
	base := filepath.Base(filepath.Dir(value)) + "-" + filepath.Base(value)
	replacer := strings.NewReplacer("/", "-", "_", "-", ".", "-")
	safe := replacer.Replace(base)
	if safe == "" {
		return "fixture"
	}
	return safe
}

type DockerCLIClient struct{}

func (DockerCLIClient) Run(ctx context.Context, spec DockerRunSpec) (string, int, error) {
	createArgs := dockerCreateArgs(spec)
	containerIDBytes, err := exec.CommandContext(ctx, "docker", createArgs...).Output()
	if err != nil {
		return "", 1, fmt.Errorf("docker create: %w", err)
	}
	containerID := strings.TrimSpace(string(containerIDBytes))
	if containerID == "" {
		return "", 1, errors.New("docker create returned empty container id")
	}
	defer exec.Command("docker", "rm", "-f", containerID).Run()

	if err := exec.CommandContext(ctx, "docker", "start", containerID).Run(); err != nil {
		return dockerLogs(containerID), 1, fmt.Errorf("docker start: %w", err)
	}
	waitCtx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()
	waitCh := make(chan dockerWaitResult, 1)
	go func() {
		out, err := exec.CommandContext(waitCtx, "docker", "wait", containerID).Output()
		waitCh <- dockerWaitResult{output: strings.TrimSpace(string(out)), err: err}
	}()

	select {
	case result := <-waitCh:
		logs := dockerLogs(containerID)
		if result.err != nil {
			return logs, 1, fmt.Errorf("docker wait: %w", result.err)
		}
		exitCode, err := strconv.Atoi(strings.TrimSpace(result.output))
		if err != nil {
			return logs, 1, fmt.Errorf("parse docker exit code: %w", err)
		}
		if exitCode != 0 {
			return logs, exitCode, fmt.Errorf("docker container exited with code %d", exitCode)
		}
		return logs, exitCode, nil
	case <-waitCtx.Done():
		grace := strconv.Itoa(int(spec.StopGrace.Seconds()))
		if err := exec.Command("docker", "stop", "--time", grace, containerID).Run(); err != nil {
			_ = exec.Command("docker", "kill", containerID).Run()
		}
		return dockerLogs(containerID), 137, waitCtx.Err()
	}
}

type dockerWaitResult struct {
	output string
	err    error
}

func dockerCreateArgs(spec DockerRunSpec) []string {
	args := []string{
		"create",
		"--name", spec.ContainerName,
		"--privileged=false",
		"--network", spec.NetworkMode,
		"--user", spec.User,
		"--read-only",
		"--cpu-quota", strconv.FormatInt(spec.CPUQuota, 10),
		"--memory", strconv.FormatInt(spec.MemoryBytes, 10),
		"--pids-limit", strconv.FormatInt(spec.PidsLimit, 10),
		"--tmpfs", "/tmp:rw,noexec,nosuid,size=16m",
	}
	for _, mount := range spec.Mounts {
		mode := "rw"
		if mount.ReadOnly {
			mode = "ro"
		}
		args = append(args, "--volume", mount.Source+":"+mount.Target+":"+mode)
	}
	args = append(args, spec.Image)
	args = append(args, spec.Command...)
	return args
}

func dockerLogs(containerID string) string {
	var stdout bytes.Buffer
	cmd := exec.Command("docker", "logs", containerID)
	cmd.Stdout = &stdout
	_ = cmd.Run()
	return stdout.String()
}
