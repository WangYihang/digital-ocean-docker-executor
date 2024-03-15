package main

import (
	"os"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/broker"
	zmap_task "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/zmap"
	gojob_utils "github.com/WangYihang/gojob/pkg/utils"
	"github.com/charmbracelet/log"
)

func init() {
	log.SetLevel(log.DebugLevel)
	fd, err := os.OpenFile("dode-1.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(gojob_utils.NewTeeWriterCloser(os.Stdout, fd))
}

func main() {
	b := broker.New().WithMaxConcurrency(1)
	for t := range zmap_task.Generate() {
		b.Submit(t)
	}
	b.Wait()
}
