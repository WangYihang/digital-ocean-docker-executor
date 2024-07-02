package main

import (
	"os"

	zmap_task "github.com/WangYihang/digital-ocean-docker-executor/examples/zmap/pkg/model/task"
	"github.com/WangYihang/digital-ocean-docker-executor/examples/zmap/pkg/option"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider/api"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/scheduler"
	gojob_utils "github.com/WangYihang/gojob/pkg/utils"
	"github.com/charmbracelet/log"
)

func init() {
	log.SetLevel(log.DebugLevel)
	fd, err := os.OpenFile(option.Opt.LogFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Error("failed to open log file", "error", err, "path", option.Opt.LogFilePath)
		os.Exit(1)
	}
	log.SetOutput(gojob_utils.NewTeeWriterCloser(os.Stdout, fd))
}

func main() {
	log.Info("starting", "options", option.Opt)
	s := scheduler.New(option.Opt.Name).
		WithProvider(provider.Use("digitalocean", option.Opt.DigitalOceanToken)).
		WithCreateServerOptions(
			api.NewCreateServerOptions().
				WithName(option.Opt.Name).
				WithTag(option.Opt.Name).
				WithRegion(option.Opt.DropletRegion).
				WithSize(option.Opt.DropletSize).
				WithImage(option.Opt.DropletImage).
				WithPrivateKeyPath(option.Opt.DropletPrivateKeyPath).
				WithPublicKeyPath(option.Opt.DropletPublicKeyPath).
				WithPublicKeyName(option.Opt.Name),
		).
		WithMaxConcurrency(option.Opt.NumDroplets)
	for t := range zmap_task.Generate(option.Opt.Name, option.Opt.Port, option.Opt.BandWidth) {
		s.Submit(t)
	}
	s.Wait()
}
