package scheduler

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/executor/secureshell"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider/api"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	"github.com/charmbracelet/log"
)

type Scheduler struct {
	name           string
	maxConcurrency int
	provider       provider.CloudServiceProvider
	cso            *api.CreateServerOptions
	wg             *sync.WaitGroup
}

func New(name string) *Scheduler {
	return &Scheduler{
		name:           name,
		maxConcurrency: 1,
		wg:             &sync.WaitGroup{},
	}
}

func (s *Scheduler) WithCreateServerOptions(cso *api.CreateServerOptions) *Scheduler {
	s.cso = cso
	return s
}

func (s *Scheduler) WithMaxConcurrency(maxConcurrency int) *Scheduler {
	s.maxConcurrency = maxConcurrency
	return s
}

func (s *Scheduler) WithProvider(provider provider.CloudServiceProvider) *Scheduler {
	s.provider = provider
	return s
}

func (s *Scheduler) FindOrCreateAnIdleExecutor() (*secureshell.SSHExecutor, error) {
	for {
		// Check if there is an idle server
		for _, server := range s.provider.ListServersByTag(s.name) {
			e := secureshell.NewSSHExecutor().
				WithIP(server.IPv4()).
				WithPrivateKeyPath(s.cso.PrivateKeyPath)
			err := e.Connect()
			if err != nil {
				log.Error("failed to connect to server", "error", err)
				continue
			}
			stdout, _, err := e.RunCommand(strings.Join([]string{
				"docker",
				"ps",
				"--quiet",
				"--filter",
				fmt.Sprintf("label=task.label=%s", s.name),
			}, " "))
			if err != nil {
				log.Error("failed to run command", "error", err)
				time.Sleep(5 * time.Second)
				continue
			}
			log.Warn("find an idle server", "server", server.IPv4())
			if strings.TrimSpace(stdout) == "" {
				return e, nil
			}
		}
		// Check if the number of servers is less than max concurrency
		servers := s.provider.ListServersByTag(s.name)
		if len(servers) < s.maxConcurrency {
			// Create a new server
			log.Info("create a new server because of no idle server and not reach max concurrency")
			server, err := s.provider.CreateServer(s.cso.WithName(fmt.Sprintf("%s-%d", s.name, len(servers))))
			if err != nil {
				log.Error("failed to create server", "error", err)
				return nil, fmt.Errorf("failed to create server: %s", err.Error())
			}
			log.Warn("sleep 5 seconds to avoid digital ocean firewall", "server", server.IPv4())
			time.Sleep(5 * time.Second)
			return secureshell.NewSSHExecutor().
				WithIP(server.IPv4()).
				WithPrivateKeyPath(s.cso.PrivateKeyPath), nil
		}
		time.Sleep(5 * time.Second)
	}
}

// A task is marked NeedRun if and only if it is not in [task.RUNNING, task.FINISHED] on any listed servers
func (s *Scheduler) NeedRun(t task.TaskInterface) bool {
	for _, server := range s.provider.ListServersByTag(s.name) {
		log.Info("check task status", "task", t, "server", server.IPv4())
		e := secureshell.NewSSHExecutor().
			WithIP(server.IPv4()).
			WithPrivateKeyPath(s.cso.PrivateKeyPath)
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

func (s *Scheduler) Submit(t task.TaskInterface) error {
	log.Info("submitting task", "task", t.String())
	s.wg.Add(1)
	// Check if the task is already assigned to a server
	if !s.NeedRun(t) {
		log.Warn("task already started", t)
		// Wait task to finish
		go s.WaitTask(t)
		return nil
	}
	// Now the task is pending state on any server
	// Find or create an idle server
	e, err := s.FindOrCreateAnIdleExecutor()
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
	go s.WaitTask(t)
	return nil
}

func (s *Scheduler) WaitTask(t task.TaskInterface) {
	defer s.wg.Done()

	for {
		// Wait task status become task.FINISHED
		status, err := t.Status()
		if err != nil {
			log.Error("task status failed", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Debug("waiting task", "status", status, "task", t.String())
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
}

func (s *Scheduler) Wait() {
	// Wait for all tasks to complete
	s.wg.Wait()
	// Destroy all servers
	s.provider.DestroyServerByTag(s.name)
}
