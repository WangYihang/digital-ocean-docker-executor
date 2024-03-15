package zmap_task

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/executor/secureshell"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	"github.com/charmbracelet/log"
)

type ZmapTask struct {
	e           *secureshell.SSHExecutor
	containerID string
	labels      map[string]interface{}
	arguments   *ZMapArguments
}

func Generate() <-chan *ZmapTask {
	out := make(chan *ZmapTask)
	go func() {
		defer close(out)
		shards := 254
		port := 80
		for shard := range shards {
			if shard >= 4 {
				break
			}
			out <- NewZmapTask(port, shard, shards)
		}
	}()
	return out
}

func NewZmapTask(port, shard, shards int) *ZmapTask {
	filename := fmt.Sprintf("zmap-%d-%d-%d", port, shard, shards)
	z := &ZmapTask{
		arguments: NewZmapArguments().
			WithTargetPort(port).
			WithShard(shard).
			WithShards(shards).
			WithBandWidth("1M").
			WithSubnet("104.245.0.0/8").
			WithOutputFileName(fmt.Sprintf("%s.json", filename)).
			WithStatusUpdateFileName(fmt.Sprintf("%s.status", filename)).
			WithLogFileName(fmt.Sprintf("%s.log", filename)),
		labels: make(map[string]interface{}),
	}
	z.labels["task.label"] = config.Cfg.Task.Label
	z.labels["task.shard"] = z.arguments.Shard
	z.labels["task.shards"] = z.arguments.Shards
	return z
}

func (z *ZmapTask) WithArguments(arguments *ZMapArguments) *ZmapTask {
	z.arguments = arguments
	return z
}

func (z *ZmapTask) String() string {
	return z.arguments.String()
}

func (z *ZmapTask) Assign(e *secureshell.SSHExecutor) error {
	z.e = e
	return z.e.Connect()
}

func (z *ZmapTask) Prepare() error {
	z.e.RunCommand("docker pull ghcr.io/zmap/zmap:latest")
	return nil
}

func (z *ZmapTask) Start() error {
	arguments := []string{
		"docker", "run",
		"--interactive", "--tty", "--detach",
		"--network", "host",
		"--volume", "/data:/data",
	}
	for k, v := range z.labels {
		arguments = append(arguments, "--label", fmt.Sprintf("%s=%v", k, v))
	}
	arguments = append(arguments, "ghcr.io/zmap/zmap:latest")
	arguments = append(arguments, z.arguments.String())
	stdout, stderr, err := z.e.RunCommand(strings.Join(arguments, " "))
	if err != nil {
		return err
	}
	if stderr != "" {
		return fmt.Errorf(stderr)
	}
	z.containerID = strings.TrimSpace(stdout)
	return nil
}

func (z *ZmapTask) Stop() error {
	_, _, err := z.e.RunCommand(strings.Join([]string{
		z.containerID,
	}, " "))
	return err
}

func (z *ZmapTask) Status() (task.StatusInterface, error) {
	// check if zmap is running
	arguments := []string{
		"docker", "ps",
		"--all",
		"--quiet",
	}
	for k, v := range z.labels {
		arguments = append(arguments, "--filter", fmt.Sprintf("label=%s=%v", k, v))
	}
	stdout, stderr, err := z.e.RunCommand(strings.Join(arguments, " "))
	if err != nil {
		return nil, err
	}
	if stderr != "" {
		return nil, fmt.Errorf(stderr)
	}
	if stdout == "" {
		return PendingProgress, nil
	}

	log.Info("zmap have already been started", "container", stdout)

	// check if the container is running
	stdout, stderr, err = z.e.RunCommand(strings.Join([]string{
		"docker",
		"inspect",
		"--format",
		"{{.State.Running}}",
		z.containerID,
	}, " "))
	if err != nil {
		return nil, err
	}
	if stderr != "" {
		return nil, fmt.Errorf(stderr)
	}
	if strings.TrimSpace(stdout) == "false" {
		// the container is not running, so it must have finished
		return DoneProgress, nil
	}

	// read the status file
	stdout, stderr, err = z.e.RunCommand(strings.Join([]string{
		"tail",
		"-n",
		"1",
		filepath.Join("/data", z.arguments.StatusUpdateFileName),
	}, " "))
	if err != nil {
		return PendingProgress, nil
	}
	if stderr != "" {
		return PendingProgress, nil
	}

	// parse the status file
	progress, err := NewZMapProgress(stdout)
	if err != nil {
		return nil, err
	}
	return progress, nil
}

func (z *ZmapTask) Download() error {
	z.e.DownloadFile(filepath.Join("/data", z.arguments.OutputFileName), filepath.Base(z.arguments.OutputFileName))
	z.e.DownloadFile(filepath.Join("/data", z.arguments.StatusUpdateFileName), filepath.Base(z.arguments.StatusUpdateFileName))
	z.e.DownloadFile(filepath.Join("/data", z.arguments.LogFileName), filepath.Base(z.arguments.LogFileName))
	return nil
}
