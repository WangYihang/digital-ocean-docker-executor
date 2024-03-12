package task

import (
	"fmt"
	"strings"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
)

type Mount struct {
	Source string
	Target string
}

func NewMount(source, target string) *Mount {
	return &Mount{
		Source: source,
		Target: target,
	}
}

func (m *Mount) String() string {
	return fmt.Sprintf("%s:%s", m.Source, m.Target)
}

type Label struct {
	Key   string
	Value string
}

func NewLabel(key, value string) *Label {
	return &Label{
		Key:   key,
		Value: value,
	}
}

func (l *Label) String() string {
	return fmt.Sprintf("%s=%s", l.Key, l.Value)
}

type DockerTask struct {
	DockerImage string
	Entrypoint  string
	Mounts      []*Mount
	Arguments   []string
	Labels      []*Label
	Interactive bool
	Tty         bool
	Detach      bool
}

func NewDockerTask() *DockerTask {
	return &DockerTask{
		DockerImage: "ubuntu:latest",
		Arguments:   []string{},
		Mounts:      []*Mount{},
		Labels:      []*Label{},
		Interactive: false,
		Tty:         false,
		Detach:      false,
	}
}

func (dt *DockerTask) String() string {
	return fmt.Sprintf("DockerTask{DockerImage: %s, Entrypoint: %s, Mounts: %v, Arguments: %v, Labels: %v, Interactive: %t, Tty: %t, Detach: %t}",
		dt.DockerImage,
		dt.Entrypoint,
		dt.Mounts,
		dt.Arguments,
		dt.Labels,
		dt.Interactive,
		dt.Tty,
		dt.Detach,
	)
}

func (dt *DockerTask) WithDockerImage(image string) *DockerTask {
	dt.DockerImage = image
	return dt
}

func (dt *DockerTask) WithMount(source, target string) *DockerTask {
	dt.Mounts = append(dt.Mounts, NewMount(source, target))
	return dt
}

func (dt *DockerTask) WithEntrypoint(entrypoint string) *DockerTask {
	dt.Entrypoint = entrypoint
	return dt
}

func (dt *DockerTask) WithArguments(arguments ...string) *DockerTask {
	dt.Arguments = append(dt.Arguments, arguments...)
	return dt
}

func (dt *DockerTask) WithLabel(key, value string) *DockerTask {
	dt.Labels = append(dt.Labels, NewLabel(key, value))
	return dt
}

func (dt *DockerTask) WithInteractive() *DockerTask {
	dt.Interactive = true
	return dt
}

func (dt *DockerTask) WithTty() *DockerTask {
	dt.Tty = true
	return dt
}

func (dt *DockerTask) WithDetach() *DockerTask {
	dt.Detach = true
	return dt
}

// Docker PUll command
func (dt *DockerTask) DockerPullCommand() string {
	parts := []string{
		"docker",
		"pull",
		dt.DockerImage,
	}
	return strings.Join(parts, " ")
}

func (dt *DockerTask) DockerRunCommand() string {
	parts := []string{
		"docker",
		"run",
	}
	if dt.Detach {
		parts = append(parts, "--detach")
	}
	if dt.Interactive {
		parts = append(parts, "--interactive")
	}
	if dt.Tty {
		parts = append(parts, "--tty")
	}
	if dt.Entrypoint != "" {
		parts = append(parts, []string{
			"--entrypoint",
			fmt.Sprintf("%q", dt.Entrypoint),
		}...)
	}
	for _, mount := range dt.Mounts {
		parts = append(parts, "-v", fmt.Sprintf(
			"%q",
			mount,
		))
	}
	for _, label := range dt.Labels {
		parts = append(parts, "--label", fmt.Sprintf(
			"%q",
			label,
		))
	}
	parts = append(parts, dt.DockerImage)
	for _, arg := range dt.Arguments {
		parts = append(parts, fmt.Sprintf("%q", arg))
	}
	return strings.Join(parts, " ")
}

// Check if any of the container project is already running on the current server
func (dt *DockerTask) DockerPsAllRelatedContainersCommand() string {
	parts := []string{
		"docker",
		"ps",
		"--quiet",
		"--filter",
		fmt.Sprintf(
			"%q",
			fmt.Sprintf("label=dode.task=%s", config.Cfg.Task.Label),
		),
	}
	return strings.Join(parts, " ")
}

// Check if the current task instance is already running on the current server
func (dt *DockerTask) DockerPsAllTaskContainersCommand() string {
	parts := []string{
		"docker",
		"ps",
		"--all",
		"--quiet",
	}
	for _, label := range dt.Labels {
		parts = append(parts, "--filter", fmt.Sprintf(
			"%q",
			fmt.Sprintf("label=%s", label),
		))
	}
	return strings.Join(parts, " ")
}

// Check if the current task instance is already running on the current server
func (dt *DockerTask) DockerPsRunningTaskContainersCommand() string {
	parts := []string{
		"docker",
		"ps",
		"--quiet",
	}
	for _, label := range dt.Labels {
		parts = append(parts, "--filter", fmt.Sprintf(
			"%q",
			fmt.Sprintf("label=%s", label),
		))
	}
	return strings.Join(parts, " ")
}

func (dt *DockerTask) GetOutputFilePaths() []string {
	return []string{
		"/etc/passwd",
		"/etc/hosts",
	}
}