package scheduler

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/executor/secureshell"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/server"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	"github.com/charmbracelet/log"
)

type Scheduler struct {
	maxConcurrency        int
	sshPublicKeyFilePath  string
	sshPrivateKeyFilePath string
	region                string
	image                 string
	size                  string
	provider              provider.CloudServiceProvider
	wg                    *sync.WaitGroup
}

func New() *Scheduler {
	return &Scheduler{
		maxConcurrency: 2,
		sshPublicKeyFilePath: filepath.Join(
			config.Cfg.DigitalOcean.SSH.Key.Folder,
			fmt.Sprintf("%s.pub", config.Cfg.DigitalOcean.SSH.Key.Name),
		),
		sshPrivateKeyFilePath: filepath.Join(
			config.Cfg.DigitalOcean.SSH.Key.Folder,
			config.Cfg.DigitalOcean.SSH.Key.Name,
		),
		region:   config.Cfg.DigitalOcean.Droplet.Region,
		image:    config.Cfg.DigitalOcean.Droplet.Image,
		size:     config.Cfg.DigitalOcean.Droplet.Size,
		provider: provider.Default(),
		wg:       &sync.WaitGroup{},
	}
}

func (pm *Scheduler) WithMaxConcurrency(maxConcurrency int) *Scheduler {
	pm.maxConcurrency = maxConcurrency
	return pm
}

func (pm *Scheduler) WithSSHKey(sshPublicKeyFilePath string, sshPrivateKeyFilePath string) *Scheduler {
	pm.sshPublicKeyFilePath = sshPublicKeyFilePath
	pm.sshPrivateKeyFilePath = sshPrivateKeyFilePath
	return pm
}

func (pm *Scheduler) WithRegion(region string) *Scheduler {
	pm.region = region
	return pm
}

func (pm *Scheduler) WithImage(image string) *Scheduler {
	pm.image = image
	return pm
}

func (pm *Scheduler) WithSize(size string) *Scheduler {
	pm.size = size
	return pm
}

func (pm *Scheduler) WithProvider(provider provider.CloudServiceProvider) *Scheduler {
	pm.provider = provider
	return pm
}

func (pm *Scheduler) Wait() {
	// Wait for all tasks to complete
	pm.wg.Wait()
	// Destroy all servers
	// pm.provider.DestroyServerByTag(config.Cfg.DigitalOcean.Droplet.Tag)
}

func (pm *Scheduler) GetTaskRunningServer(task *task.DockerTask) (server.Server, error) {
	tag := config.Cfg.DigitalOcean.Droplet.Tag
	for _, server := range pm.provider.ListServersByTag(tag) {
		log.Debug("checking if task running the server", "server", server.IPv4(), "task", task)
		e := secureshell.NewSSHExecutor(
			server.IPv4(),
			config.Cfg.DigitalOcean.SSH.Port,
			config.Cfg.DigitalOcean.SSH.User,
			filepath.Join(
				config.Cfg.DigitalOcean.SSH.Key.Folder,
				config.Cfg.DigitalOcean.SSH.Key.Name,
			),
		)
		e.Connect()
		for {
			// Check if the current task is already running on the current server
			cmd := task.DockerPsAllTaskContainersCommand()
			stdout, _, err := e.RunCommand(cmd)
			if err != nil {
				log.Error("error occurred when running command", "error", err, "cmd", cmd)
				continue
			}
			if stdout != "" {
				log.Warn("task is already assigned", "server", server.IPv4(), "task", task)
				return server, nil
			}
			log.Debug("task is not running", "server", server.IPv4(), "task", task)
			break
		}
	}
	return nil, fmt.Errorf("task is not running on any server")
}

func (pm *Scheduler) WaitTask(task *task.DockerTask, server server.Server) {
	defer pm.wg.Done()
	// Initialize a new SSH executor to interact with the server
	e := secureshell.NewSSHExecutor(
		server.IPv4(),
		config.Cfg.DigitalOcean.SSH.Port,
		config.Cfg.DigitalOcean.SSH.User,
		filepath.Join(
			config.Cfg.DigitalOcean.SSH.Key.Folder,
			config.Cfg.DigitalOcean.SSH.Key.Name,
		),
	)
	e.Connect()
	for {
		// Log waiting status for the task
		log.Warn("waiting for task to complete", "server", server.IPv4())
		// Run command to check if Docker containers for the task are still running
		stdout, _, err := e.RunCommand(task.DockerPsRunningTaskContainersCommand())
		if err != nil {
			continue
		}
		if stdout == "" {
			// Task is complete when there are no running containers
			log.Warn("task done", "server", server.IPv4(), "task", task)
			// Download output file from the server
			task.RetrieveOutput(server.IPv4())
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func (pm *Scheduler) SubmitDockerTask(task *task.DockerTask) {
	pm.wg.Add(1)

	// Log information about scheduling the task
	log.Info("scheduling task", "task", task)

	// Check if the task is already running or finished on any server
	tag := config.Cfg.DigitalOcean.Droplet.Tag
	taskRunningServer, err := pm.GetTaskRunningServer(task)

	// If task is already running, return immediately
	if err == nil {
		go pm.WaitTask(task, taskRunningServer)
		return
	}

	// Find an idle server to run the task
	idleServer := func() server.Server {
		for {
			// Iterate through servers and return the first idle one
			for _, server := range pm.provider.ListServersByTag(tag) {
				log.Debug("checking if server is idle", "server", server.IPv4())
				e := secureshell.NewSSHExecutor(
					server.IPv4(),
					config.Cfg.DigitalOcean.SSH.Port,
					config.Cfg.DigitalOcean.SSH.User,
					filepath.Join(
						config.Cfg.DigitalOcean.SSH.Key.Folder,
						config.Cfg.DigitalOcean.SSH.Key.Name,
					),
				)
				e.Connect()
				for {
					cmd := task.DockerPsAllRelatedRunningContainersCommand()
					stdout, _, err := e.RunCommand(cmd)
					if err != nil {
						log.Error("error occurred when running command", "error", err, "cmd", cmd)
						continue
					}
					if stdout == "" {
						// Server is idle if no related Docker containers are running
						log.Debug("server is idle", "server", server.IPv4())
						return server
					}
					log.Debug("server is not idle", "server", server.IPv4())
					break
				}
			}

			// If no idle servers are available, create a new one if not exceeding the concurrency limit
			servers := pm.provider.ListServersByTag(tag)
			if len(servers) < pm.maxConcurrency {
				server, err := pm.provider.CreateServer(fmt.Sprintf(
					"%s-%d",
					config.Cfg.DigitalOcean.Droplet.Name,
					len(servers)+1,
				), tag)

				if err != nil {
					continue
				}
				e := secureshell.NewSSHExecutor(
					server.IPv4(),
					config.Cfg.DigitalOcean.SSH.Port,
					config.Cfg.DigitalOcean.SSH.User,
					filepath.Join(
						config.Cfg.DigitalOcean.SSH.Key.Folder,
						config.Cfg.DigitalOcean.SSH.Key.Name,
					),
				)
				e.Connect()
				// Run initialization scripts on the new server
				e.RunExecutable("./assets/scripts/ubuntu-22-04-x64/add-swap.sh")

				// Docker pull
				e.RunCommand(task.DockerPullCommand())
			}

			// Wait before checking again for idle servers
			time.Sleep(5 * time.Second)
		}
	}()

	// Execute the task on the found idle server
	e := secureshell.NewSSHExecutor(
		idleServer.IPv4(),
		config.Cfg.DigitalOcean.SSH.Port,
		config.Cfg.DigitalOcean.SSH.User,
		filepath.Join(
			config.Cfg.DigitalOcean.SSH.Key.Folder,
			config.Cfg.DigitalOcean.SSH.Key.Name,
		),
	)
	e.Connect()

	// Download tranco list
	e.RunCommand("docker pull ghcr.io/wangyihang/tranco-go-package:main")
	e.RunCommand("docker run -v /root/.tranco:/root/.tranco ghcr.io/wangyihang/tranco-go-package:main --date 2024-01-01")

	// Run the Docker task command
	cmd := task.DockerRunCommand()
	stdout, stderr, err := e.RunCommand(cmd)
	if err != nil {
		log.Error("error occurred when running command", "error", err, "cmd", cmd, "stdout", stdout, "stderr", stderr)
		return
	}

	// Start waiting for the task to complete in a separate goroutine
	go pm.WaitTask(task, idleServer)
	log.Warn("task assigned", "server", idleServer.IPv4(), "task", task, "stdout", stdout, "stderr", stderr)
}
