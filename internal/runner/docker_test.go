package runner

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestBuildDockerTrustedAdminSpecSecurity(t *testing.T) {
	spec, err := BuildDockerTrustedAdminSpec("/tmp/job1/work/input/job.json", "/tmp/job1/work/output", DockerTrustedAdminConfig{
		Image:                "python-basic:test",
		ConnectorPackagePath: "/tmp/fixture",
	})
	if err != nil {
		t.Fatalf("build docker spec: %v", err)
	}
	if spec.Privileged {
		t.Fatal("container must not be privileged")
	}
	if spec.NetworkMode == "host" || spec.NetworkMode == "" {
		t.Fatalf("unexpected network mode: %s", spec.NetworkMode)
	}
	if spec.User == "" || strings.HasPrefix(spec.User, "0") || spec.User == "root" {
		t.Fatalf("container must not run as root, got %s", spec.User)
	}
	if !spec.ReadOnlyRootFS {
		t.Fatal("root filesystem must be read-only")
	}
	if spec.CPUQuota <= 0 || spec.MemoryBytes <= 0 || spec.PidsLimit <= 0 || spec.Timeout <= 0 {
		t.Fatalf("missing resource limits: %+v", spec)
	}
	var workMounts int
	var connectorMounts int
	for _, mount := range spec.Mounts {
		if strings.Contains(mount.Source, "docker.sock") || mount.Source == "/" {
			t.Fatalf("unsafe host mount: %+v", mount)
		}
		if mount.Target == "/work" {
			workMounts++
			if mount.ReadOnly {
				t.Fatalf("/work must be writable: %+v", mount)
			}
		}
		if mount.Target == "/connector" {
			connectorMounts++
			if !mount.ReadOnly {
				t.Fatalf("connector package must be read-only: %+v", mount)
			}
		}
	}
	if workMounts != 1 || connectorMounts != 1 || spec.WritableTarget != "/work" {
		t.Fatalf("unexpected mounts: %+v", spec)
	}
}

func TestDockerTrustedAdminExecutorRunsFixtureWithFakeDocker(t *testing.T) {
	client := &fakeDockerClient{}
	executor := DockerTrustedAdminExecutor{
		Client: client,
		Config: DockerTrustedAdminConfig{
			Image:                "python-basic:test",
			ConnectorPackagePath: t.TempDir(),
			Timeout:              time.Second,
		},
	}
	result := executor.Execute(context.Background(), "/tmp/job1/work/input/job.json", "/tmp/job1/work/output")
	if result.Err != nil {
		t.Fatalf("execute docker fixture: %v", result.Err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d", result.ExitCode)
	}
	if client.spec.Image != "python-basic:test" || !strings.Contains(strings.Join(client.spec.Command, " "), "fixture_connector.py") {
		t.Fatalf("unexpected docker spec: %+v", client.spec)
	}
}

type fakeDockerClient struct {
	spec DockerRunSpec
}

func (c *fakeDockerClient) Run(_ context.Context, spec DockerRunSpec) (string, int, error) {
	c.spec = spec
	return `{"type":"completed","message":"docker fixture completed"}` + "\n", 0, nil
}
