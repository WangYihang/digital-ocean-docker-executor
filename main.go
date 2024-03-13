package main

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/scheduler"
	task "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/cdn-domain-harvest/brute-dns-resolve"
	"github.com/charmbracelet/log"
)

func main() {
	log.SetLevel(log.DebugLevel)
	s := scheduler.New().WithMaxConcurrency(4)
	for t := range task.Generate(config.Cfg.Task.Label) {
		s.SubmitDockerTask(t)
	}
	s.Wait()
}
