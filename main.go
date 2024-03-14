package main

import (
	"os"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/scheduler"
	task "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/cdn-domain-harvest/brute-dns-resolve"
	gojob_utils "github.com/WangYihang/gojob/pkg/utils"
	"github.com/charmbracelet/log"
)

func init() {
	log.SetLevel(log.DebugLevel)
	fd, err := os.OpenFile("dode.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(gojob_utils.NewTeeWriterCloser(os.Stdout, fd))
}

func main() {
	s := scheduler.New().WithMaxConcurrency(1)
	for t := range task.Generate(config.Cfg.Task.Label) {
		s.SubmitDockerTask(t)
	}
	s.Wait()
}
