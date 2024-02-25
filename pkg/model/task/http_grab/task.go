package http_grab

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
)

type Task struct {
	// Basic
	Name           string
	RootPath       string
	InputFilePath  string
	StatusFilePath string
	// Command line tool arguments
	Arguments HTTPGrabArguments
	// Docker
	DockerfilePath  string
	DockerImageName string
	// Executor
	Executor *secureshell.SSHExecutor
	// Progress
	Progress chan task.Progress
}

func NewTask(name string, dockerfilePath string, inputFilePath string, arguments HTTPGrabArguments) *Task {
	return &Task{
		Name:           name,
		RootPath:       filepath.Join("/root", name),
		InputFilePath:  inputFilePath,
		StatusFilePath: filepath.Join("/root", name, "data", filepath.Base(arguments.StatusUpdateFilePath)),

		// Command line tool arguments
		Arguments: arguments,
		// Docker
		DockerfilePath:  dockerfilePath,
		DockerImageName: name,

		// Executor
		Executor: nil,
		// Progress
		Progress: make(chan task.Progress),
	}
}

func (t *Task) String() string {
	return fmt.Sprintf(
		"Task<Name: %s, DockerfilePath: %s> (Executor: %s)",
		t.Name,
		t.DockerfilePath,
		t.Executor,
	)
}

func (t *Task) Assign(e *secureshell.SSHExecutor) {
	t.Executor = e
	for {
		err := t.Executor.Connect()
		if err != nil {
			slog.Error("error occured while connecting to executor", slog.String("error", err.Error()))
			continue
		}
		break
	}
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
	// Install Docker
	t.Executor.InstallDocker()
	// Create Root Path
	_, _, err := t.Executor.RunCommand(fmt.Sprintf("mkdir -p %s", filepath.Join(t.RootPath, "data")), true)
	if err != nil {
		return err
	}
	remoteDockerfilePath := filepath.Join(t.RootPath, "Dockerfile")
	// Upload Dockerfile
	slog.Info("uploading dockerfile", slog.String("from", t.DockerfilePath), slog.String("to", remoteDockerfilePath))
	err = t.Executor.UploadFile(t.DockerfilePath, remoteDockerfilePath)
	if err != nil {
		return err
	}
	// Build Docker Image
	slog.Info("building docker image")
	_, _, err = t.Executor.RunCommand(fmt.Sprintf("docker build -t %s %s", t.DockerImageName, t.RootPath), true)
	if err != nil {
		return err
	}
	// Upload Input File
	inputFileName := filepath.Base(t.InputFilePath)
	remoteInputFilePath := filepath.Join(t.RootPath, "data", inputFileName)
	slog.Info("uploading input file", slog.String("from", t.InputFilePath), slog.String("to", remoteInputFilePath))
	t.UploadFile(t.InputFilePath, remoteInputFilePath)
	// Start Docker Container
	remoteOutputFolder := filepath.Join(t.RootPath, "data")
	command := fmt.Sprintf(
		"docker run --network host -v %s:/data --interactive --tty --detach %s %s",
		remoteOutputFolder,
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
	_, _, err := t.Executor.RunCommand(fmt.Sprintf("stat %s", t.StatusFilePath), false)
	return err == nil
}

func (t *Task) IsFinished() bool {
	isStarted := t.IsStarted()
	var progress *HTTPGrabProgress
	for {
		stdout, _, err := t.Executor.RunCommand(fmt.Sprintf("tail -n 1 %s", t.StatusFilePath), false)
		if err != nil {
			slog.Error("error occured while reading log file", slog.String("error", err.Error()))
			return false
		}
		if stdout != "" {
			progress = NewHTTPGrabProgress(stdout)
			slog.Info("current progress", slog.String("ip", t.Executor.IP), slog.String("progress", progress.String()))
			break
		}
	}
	return isStarted && progress.Done()
}

func (t *Task) RetrieveResults() error {
	command := fmt.Sprintf(
		`rsync -e "ssh -o StrictHostKeyChecking=no -i %s/%s" -avz "%s@%s:/root/%s/data/" ./data/%s/`,
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

func (t *Task) UploadFile(src string, dst string) error {
	command := fmt.Sprintf(
		`rsync -e "ssh -o StrictHostKeyChecking=no -i %s/%s" -avz "%s" "%s@%s:%s"`,
		config.Cfg.SSHKeyFolder,
		config.Cfg.SSHKeyName,
		src,
		config.Cfg.SSHUser,
		t.Executor.IP,
		dst,
	)
	slog.Info("uploading file", slog.String("command", command))
	return exec.Command("bash", "-c", command).Run()
}
