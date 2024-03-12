package task_test

import (
	"testing"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
)

func TestDockerCommand(t *testing.T) {
	dt := task.NewDockerTask().
		WithDockerImage("alpine").
		WithMount("/tmp", "/tmp").
		WithEntrypoint("sh").
		WithArguments("-c", "echo hello world").
		WithLabel("project", "test-project").
		WithInteractive().
		WithTty().
		WithDetach()
	expected := `docker run --detach --interactive --tty --entrypoint "sh" -v "/tmp:/tmp" --label "project=test-project" alpine "-c" "echo hello world"`
	actual := dt.DockerRunCommand()
	if actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
}
