package zmap_task

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/WangYihang/digital-ocean-docker-executor/examples/zmap/pkg/option"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/executor/secureshell"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	"github.com/charmbracelet/log"
)

type ZmapTask struct {
	e            *secureshell.SSHExecutor
	image        string
	containerID  string
	labels       map[string]interface{}
	arguments    *ZMapArguments
	outputFolder string
}

func Generate(name string, port int, bandwidth string) <-chan *ZmapTask {
	out := make(chan *ZmapTask)
	go func() {
		defer close(out)
		shards := 254
		for shard := range shards {
			if shard >= 4 {
				break
			}
			out <- New(port, shard, shards, name, bandwidth)
		}
	}()
	return out
}

func New(port, shard, shards int, label, bandwidth string) *ZmapTask {
	folder := fmt.Sprintf("/data/zmap/%d", port)
	path := fmt.Sprintf("zmap-%d-%d-%d", port, shard, shards)
	z := &ZmapTask{
		arguments: NewZmapArguments().
			WithTargetPort(port).
			WithShard(shard).
			WithShards(shards).
			WithBandWidth(bandwidth).
			WithOutputFileName(fmt.Sprintf("%s.json", path)).
			WithStatusUpdateFileName(fmt.Sprintf("%s.status", path)).
			WithLogFileName(fmt.Sprintf("%s.log", path)),
		labels: make(map[string]interface{}),
		image:  "ghcr.io/zmap/zmap:latest",
	}
	z.outputFolder = folder
	z.labels["task.label"] = label
	z.labels["task.port"] = port
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
	images := []string{
		"amazon/aws-cli",
		z.image,
	}
	for _, image := range images {
		z.e.RunCommand(strings.Join([]string{
			"docker", "pull", image,
		}, " "))
	}
	z.e.RunCommand(fmt.Sprintf(
		"mkdir -p %s && wget -O %s/ipinfo-%d-%d.json https://ipinfo.io/json",
		z.outputFolder,
		z.outputFolder,
		z.arguments.Shard,
		z.arguments.Shards,
	))
	return nil
}

func (z *ZmapTask) Start() error {
	arguments := []string{
		"docker", "run",
		"--interactive", "--tty", "--detach",
		"--network", "host",
		"--volume", fmt.Sprintf(
			"%s:/data",
			z.outputFolder,
		),
	}
	for k, v := range z.labels {
		arguments = append(arguments, "--label", fmt.Sprintf("%s=%v", k, v))
	}
	arguments = append(arguments, z.image)
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
	log.Info("zmap status", "stdout", stdout, "stderr", stderr, "err", err, "task", z.String())
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
	z.containerID = strings.TrimSpace(stdout)

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
		filepath.Join(z.outputFolder, z.arguments.StatusUpdateFileName),
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
	// upload to amazon s3
	if option.Opt.S3AccessKey != "" {
		today := time.Now().Format("2006-01-02")
		z.e.RunCommand(fmt.Sprintf("docker run --rm -v ~/.aws:/root/.aws amazon/aws-cli configure set aws_access_key_id %s", option.Opt.S3Option.S3AccessKey))
		z.e.RunCommand(fmt.Sprintf("docker run --rm -v ~/.aws:/root/.aws amazon/aws-cli configure set aws_secret_access_key %s", option.Opt.S3Option.S3SecretKey))
		z.e.RunCommand(fmt.Sprintf("docker run --rm -v ~/.aws:/root/.aws amazon/aws-cli configure set default.region %s", option.Opt.S3Option.S3Region))
		z.e.RunCommand(fmt.Sprintf("docker run --rm -v %s:/data -v ~/.aws:/root/.aws amazon/aws-cli s3 cp /data/ s3://dode/%s/%d/%s/ --recursive", z.outputFolder, option.Opt.Name, option.Opt.Port, today))
	}

	// Download to local
	z.e.DownloadFile(filepath.Join("/data", z.arguments.OutputFileName), filepath.Join("data", filepath.Base(z.arguments.OutputFileName)))
	z.e.DownloadFile(filepath.Join("/data", z.arguments.StatusUpdateFileName), filepath.Join("data", filepath.Base(z.arguments.StatusUpdateFileName)))
	z.e.DownloadFile(filepath.Join("/data", z.arguments.LogFileName), filepath.Join("data", filepath.Base(z.arguments.LogFileName)))
	return nil
}
