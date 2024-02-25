package zmap_scan

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/executor/secureshell"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	"github.com/jszwec/csvutil"
)

type ZMapProgress struct {
	RealTime              string  `csv:"real-time" json:"real-time"`
	TimeElapsed           int     `csv:"time-elapsed" json:"time-elapsed"`
	TimeRemaining         int     `csv:"time-remaining" json:"time-remaining"`
	PercentComplete       float64 `csv:"percent-complete" json:"percent-complete"`
	HitRate               float64 `csv:"hit-rate" json:"hit-rate"`
	ActiveSendThreads     int     `csv:"active-send-threads" json:"active-send-threads"`
	SentTotal             int     `csv:"sent-total" json:"sent-total"`
	SentLastOneSec        int     `csv:"sent-last-one-sec" json:"sent-last-one-sec"`
	SentAvgPerSec         int     `csv:"sent-avg-per-sec" json:"sent-avg-per-sec"`
	RecvSuccessTotal      int     `csv:"recv-success-total" json:"recv-success-total"`
	RecvSuccessLastOneSec int     `csv:"recv-success-last-one-sec" json:"recv-success-last-one-sec"`
	RecvSuccessAvgPerSec  int     `csv:"recv-success-avg-per-sec" json:"recv-success-avg-per-sec"`
	RecvTotal             int     `csv:"recv-total" json:"recv-total"`
	RecvTotalLastOneSec   int     `csv:"recv-total-last-one-sec" json:"recv-total-last-one-sec"`
	RecvTotalAvgPerSec    int     `csv:"recv-total-avg-per-sec" json:"recv-total-avg-per-sec"`
	PcapDropTotal         int     `csv:"pcap-drop-total" json:"pcap-drop-total"`
	DropLastOneSec        int     `csv:"drop-last-one-sec" json:"drop-last-one-sec"`
	DropAvgPerSec         int     `csv:"drop-avg-per-sec" json:"drop-avg-per-sec"`
	SendtoFailTotal       int     `csv:"sendto-fail-total" json:"sendto-fail-total"`
	SendtoFailLastOneSec  int     `csv:"sendto-fail-last-one-sec" json:"sendto-fail-last-one-sec"`
	SendtoFailAvgPerSec   int     `csv:"sendto-fail-avg-per-sec" json:"sendto-fail-avg-per-sec"`
	Line                  string
}

func NewZMapProgress(message string) *ZMapProgress {
	content := strings.Join([]string{
		"real-time,time-elapsed,time-remaining,percent-complete,hit-rate,active-send-threads,sent-total,sent-last-one-sec,sent-avg-per-sec,recv-success-total,recv-success-last-one-sec,recv-success-avg-per-sec,recv-total,recv-total-last-one-sec,recv-total-avg-per-sec,pcap-drop-total,drop-last-one-sec,drop-avg-per-sec,sendto-fail-total,sendto-fail-last-one-sec,sendto-fail-avg-per-sec",
		message,
	}, "\n")
	var progresses []ZMapProgress
	if err := csvutil.Unmarshal([]byte(content), &progresses); err != nil {
		slog.Error("error occured while parsing progress", slog.String("error", err.Error()))
		return &ZMapProgress{}
	}
	if len(progresses) == 0 {
		return &ZMapProgress{}
	} else {
		progress := progresses[0]
		progress.Line = message
		return &progress
	}
}

func (z ZMapProgress) String() string {
	return fmt.Sprintf("%s (%f%%)", time.Duration(z.TimeRemaining)*time.Second, z.PercentComplete)
}

func (z ZMapProgress) Done() bool {
	return z.PercentComplete >= 100
}

type ZMapArguments struct {
	TargetPort           int
	OutputFileName       string
	LogFileName          string
	StatusUpdateFileName string
	BandWidth            string
	Seed                 int
	Shards               int
	Shard                int
}

func (z *ZMapArguments) String() string {
	return fmt.Sprintf(
		"--target-port=%d --output-file=/data/%s --log-file=/data/%s --status-updates-file=/data/%s --bandwidth=%s --seed=%d --shards=%d --shard=%d",
		z.TargetPort,
		z.OutputFileName,
		z.LogFileName,
		z.StatusUpdateFileName,
		z.BandWidth,
		z.Seed,
		z.Shards,
		z.Shard,
	)
}

type Task struct {
	// Basic
	Name                 string
	TaskRootPath         string
	DataFolder           string
	DataFileName         string
	DataFilePath         string
	LogFilePath          string
	StatusUpdateFilePath string
	// Zmap arguments
	Arguments ZMapArguments
	// Docker
	LocalDockerfilePath  string
	RemoteDockerfilePath string
	DockerImageName      string
	// Executor
	Executor *secureshell.SSHExecutor
	// Progress
	Progress chan task.Progress
}

func NewTask(name string, arguments ZMapArguments, dockerfilePath string) *Task {
	return &Task{
		Name:                name,
		Arguments:           arguments,
		LocalDockerfilePath: dockerfilePath,
		DockerImageName:     name,
		Executor:            nil,
		Progress:            make(chan task.Progress),
	}
}

func (t *Task) String() string {
	return fmt.Sprintf(
		"Task<Name: %s, DockerfilePath: %s> (Executor: %s)",
		t.Name,
		t.LocalDockerfilePath,
		t.Executor,
	)
}

func (t *Task) Assign(e *secureshell.SSHExecutor) {
	t.Executor = e
	err := t.Executor.Connect()
	if err != nil {
		slog.Error("error occured while connecting to executor", slog.String("error", err.Error()))
	}
	t.TaskRootPath = filepath.Join("/root", t.Name)
	t.RemoteDockerfilePath = filepath.Join(t.TaskRootPath, "Dockerfile")
	t.DataFolder = filepath.Join(t.TaskRootPath, "data")
	t.DataFilePath = filepath.Join(t.DataFolder, t.Arguments.OutputFileName)
	t.LogFilePath = filepath.Join(t.DataFolder, t.Arguments.LogFileName)
	t.StatusUpdateFilePath = filepath.Join(t.DataFolder, t.Arguments.StatusUpdateFileName)
}

func (t *Task) GetDockerContainerID() string {
	stdout, _, err := t.Executor.RunCommand(fmt.Sprintf("docker ps --quiet --filter ancestor=%s", t.DockerImageName), true)
	if err != nil {
		slog.Error("error occured while retrieving docker container id", slog.String("error", err.Error()))
		return ""
	}
	return strings.TrimSpace(stdout)
}

func (t *Task) Start() error {
	t.Executor.InstallDocker()
	_, _, err := t.Executor.RunCommand(fmt.Sprintf("mkdir -p %s", t.TaskRootPath), true)
	if err != nil {
		return err
	}
	slog.Info("uploading dockerfile", slog.String("from", t.LocalDockerfilePath), slog.String("to", t.RemoteDockerfilePath))
	err = t.Executor.UploadFile(t.LocalDockerfilePath, t.RemoteDockerfilePath)
	if err != nil {
		return err
	}
	slog.Info("building docker image")
	_, _, err = t.Executor.RunCommand(fmt.Sprintf("docker build -t %s %s", t.DockerImageName, t.TaskRootPath), true)
	if err != nil {
		return err
	}
	slog.Info("creating data folder", slog.String("path", t.DataFolder))
	_, _, err = t.Executor.RunCommand(fmt.Sprintf("mkdir -p %s", t.DataFolder), true)
	if err != nil {
		return err
	}
	command := fmt.Sprintf(
		"docker run --network host -v %s:/data --interactive --tty --detach %s %s",
		t.DataFolder,
		t.DockerImageName,
		t.Arguments.String(),
	)
	slog.Info("starting docker container", slog.String("command", command))
	_, _, err = t.Executor.RunCommand(command, true)
	if err != nil {
		return err
	}
	return nil
}

func (t *Task) Wait() {
	for !t.IsFinished() {
		time.Sleep(1 * time.Second)
	}
	t.Executor.Close()
}

func (t *Task) IsStarted() bool {
	_, _, err := t.Executor.RunCommand(fmt.Sprintf("stat %s", t.StatusUpdateFilePath), false)
	return err == nil
}

func (t *Task) IsFinished() bool {
	time.Sleep(1 * time.Second)

	isStarted := t.IsStarted()

	// Read status update file
	var progress *ZMapProgress
	for {
		stdout, _, err := t.Executor.RunCommand(fmt.Sprintf("tail -n 1 %s", t.StatusUpdateFilePath), false)
		if err != nil {
			slog.Error("error occured while reading log file", slog.String("error", err.Error()))
			return false
		}
		if stdout != "" {
			progress = NewZMapProgress(stdout)
			break
		}
	}

	// Check if log file is completed
	stdout, _, err := t.Executor.RunCommand(fmt.Sprintf("tail -n 1 %s", t.LogFilePath), false)
	if err != nil {
		slog.Error("error occured while reading log file", slog.String("error", err.Error()))
		return false
	}
	logCompleted := strings.Contains(stdout, "zmap: completed")
	if logCompleted {
		progress.PercentComplete = 100
	}

	slog.Info("running", slog.Int("shard", t.Arguments.Shard), slog.Int("shards", t.Arguments.Shards), slog.String("progress", progress.String()))
	return isStarted && (progress.Done() || logCompleted)
}

func (t *Task) RetrieveResults() error {
	command := fmt.Sprintf(
		`rsync -e "ssh -o StrictHostKeyChecking=no -i %s/%s" -avz "%s@%s:/root/%s/data/" ./%s/`,
		config.Cfg.SSHKeyFolder,
		config.Cfg.SSHKeyName,
		config.Cfg.SSHUser,
		t.Executor.IP,
		t.Name,
		t.Name,
	)
	slog.Info("retrieving results", slog.String("command", command))
	return exec.Command("bash", "-c", command).Run()
}
