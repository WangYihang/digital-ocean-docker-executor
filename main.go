package main

import (
	"log/slog"
	"sync"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/executor/secureshell"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/http_grab"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/util/digitalocean"
)

func main() {
	numShards := 4
	// Create tasks
	tasks := http_grab.GenerateTasks(numShards)
	// Create executor
	executors := secureshell.GenerateExecutors(numShards)
	// Assign tasks to executor
	for i := 0; i < numShards; i++ {
		task := tasks[i]
		executor := executors[i]
		task.Assign(executor)
		slog.Info("task assigned to executor", slog.String("task", task.String()), slog.String("executor", executor.String()))
	}
	// Start unstarted tasks
	for _, task := range tasks {
		if !task.IsStarted() {
			slog.Info("task has not been started", slog.String("task", task.String()))
			slog.Info("starting task", slog.String("task", task.String()))
			err := task.Start()
			if err != nil {
				slog.Error("error occured when starting task", slog.String("error", err.Error()))
				continue
			}
		}
	}
	// Wait for all tasks to finish
	wg := sync.WaitGroup{}
	wg.Add(len(tasks))
	for _, task := range tasks {
		go func(t *http_grab.Task) {
			slog.Info("waiting for task to finish", slog.String("task", t.String()))
			t.Wait()
			slog.Info("task finished", slog.String("task", t.String()))
			wg.Done()
		}(task)
	}
	wg.Wait()
	// Retrieve results
	for _, task := range tasks {
		err := task.RetrieveResults()
		if err != nil {
			slog.Error("error occured when retrieving results", slog.String("error", err.Error()))
		}
	}
	// Destroy droplets
	if config.Cfg.DestroyDropletsAfterTaskFinished {
		slog.Info("destroying droplets")
		digitalocean.DestroyDroplets(digitalocean.LoadDropletsFromRemoteAPI(config.Cfg.Tag))
	}
}
