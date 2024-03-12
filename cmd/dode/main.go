package main

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider/digitalocean"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/scheduler"
	docker_task "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/docker-task"
)

func main() {
	p := digitalocean.NewProvider()
	s := scheduler.New().WithProvider(p)
	for t := range docker_task.Generate(config.Cfg.Project.Name) {
		s.SubmitDockerTask(t)
	}
	s.Wait()
}
