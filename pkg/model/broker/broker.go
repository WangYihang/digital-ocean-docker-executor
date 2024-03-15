package broker

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/executor/secureshell"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	"github.com/charmbracelet/log"
)

type Scheduler struct {
	maxConcurrency int
	provider       provider.CloudServiceProvider
	wg             *sync.WaitGroup
}

func New() *Scheduler {
	return &Scheduler{
		maxConcurrency: 1,
		provider:       provider.Default(),
		wg:             &sync.WaitGroup{},
	}
}

func (pm *Scheduler) WithMaxConcurrency(maxConcurrency int) *Scheduler {
	pm.maxConcurrency = maxConcurrency
	return pm
}

func (pm *Scheduler) WithProvider(provider provider.CloudServiceProvider) *Scheduler {
	pm.provider = provider
	return pm
}

func (pm *Scheduler) FindOrCreateAnIdleExecutor() (*secureshell.SSHExecutor, error) {
	tag := config.Cfg.DigitalOcean.Droplet.Tag
	for _, server := range pm.provider.ListServersByTag(tag) {
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
		isIdle := false
		for {
			stdout, _, err := e.RunCommand(strings.Join([]string{
				"docker",
				"ps",
				"--quiet",
				"--filter",
				fmt.Sprintf("label=task.label=%s", config.Cfg.Task.Label),
			}, " "))
			if err != nil {
				log.Error("failed to run command", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}
			if strings.TrimSpace(stdout) == "" {
				isIdle = true
				break
			}
			time.Sleep(5 * time.Second)
		}
		log.Warn("find an idle server", "server", server.IPv4())
		if isIdle {
			return e, nil
		}
	}
	servers := pm.provider.ListServersByTag(tag)
	if len(servers) < pm.maxConcurrency {
		server, err := pm.provider.CreateServer(fmt.Sprintf(
			"%s-%d",
			config.Cfg.DigitalOcean.Droplet.Name,
			len(servers)+1,
		), tag)
		if err != nil {
			log.Error("failed to create server", "error", err)
			return nil, fmt.Errorf("failed to create server: %s", err.Error())
		}
		return secureshell.NewSSHExecutor(
			server.IPv4(),
			config.Cfg.DigitalOcean.SSH.Port,
			config.Cfg.DigitalOcean.SSH.User,
			filepath.Join(
				config.Cfg.DigitalOcean.SSH.Key.Folder,
				config.Cfg.DigitalOcean.SSH.Key.Name,
			),
		), nil
	}
	return nil, fmt.Errorf("task is not running on any server")
}

// A task is marked NeedRun if and only if it is not in [task.RUNNING, task.FINISHED] on any listed servers
func (pm *Scheduler) NeedRun(t task.TaskInterface) bool {
	for _, server := range pm.provider.ListServersByTag(config.Cfg.DigitalOcean.Droplet.Tag) {
		e := secureshell.NewSSHExecutor(
			server.IPv4(),
			config.Cfg.DigitalOcean.SSH.Port,
			config.Cfg.DigitalOcean.SSH.User,
			filepath.Join(
				config.Cfg.DigitalOcean.SSH.Key.Folder,
				config.Cfg.DigitalOcean.SSH.Key.Name,
			),
		)
		err := t.Assign(e)
		if err != nil {
			log.Error("failed to assign task to executor", "error", err)
			continue
		}
		status, err := t.Status()
		if err != nil {
			log.Error("failed to get task status", "error", err)
			continue
		}
		if status.GetStatus() == task.RUNNING || status.GetStatus() == task.FINISHED {
			return false
		}
	}
	return true
}

func (pm *Scheduler) Submit(t task.TaskInterface) error {
	pm.wg.Add(1)
	// Check if the task is already assigned to a server
	if !pm.NeedRun(t) {
		log.Warn("task already started", t)
		// Wait task to finish
		go pm.WaitTask(t)
		return nil
	}
	// Now the task is pending state on any server
	// Find or create an idle server
	e, err := pm.FindOrCreateAnIdleExecutor()
	if err != nil {
		return err
	}
	// Assign the task to the server (executer)
	err = t.Assign(e)
	if err != nil {
		return err
	}
	// Prepare task prerequisites
	for {
		err := t.Prepare()
		if err != nil {
			log.Error("prepare failed", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Info("prepare succeed")
		break
	}
	// Start the task
	for {
		err := t.Start()
		if err != nil {
			log.Error("start failed", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Info("start succeed")
		break
	}
	// Wait task to finish
	go pm.WaitTask(t)
	return nil
}

func (pm *Scheduler) WaitTask(t task.TaskInterface) {
	for {
		// Wait task status become task.FINISHED
		status, err := t.Status()
		if err != nil {
			log.Error("task status failed", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debug("waiting task", "status", status)
		if status.GetStatus() == task.FINISHED {
			break
		}
		time.Sleep(5 * time.Second)
	}
	for {
		// Download task output files
		err := t.Download()
		if err != nil {
			log.Error("task output download failed", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Info("task output download succeed")
		break
	}
	pm.wg.Done()
}

func (pm *Scheduler) Wait() {
	// Wait for all tasks to complete
	pm.wg.Wait()
	// Destroy all servers
	pm.provider.DestroyServerByTag(config.Cfg.DigitalOcean.Droplet.Tag)
}
