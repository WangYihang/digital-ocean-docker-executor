package main

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider/digitalocean"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/scheduler"
	brute_dns_resolve_task "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/brute-dns-resolve"
	"github.com/charmbracelet/log"
)

func main() {
	log.SetLevel(log.DebugLevel)
	p := digitalocean.NewProvider()
	s := scheduler.New().WithProvider(p)
	for t := range brute_dns_resolve_task.Generate(config.Cfg.Task.Label) {
		s.SubmitDockerTask(t)
	}
	s.Wait()
}
