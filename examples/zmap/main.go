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
	fd, err := os.OpenFile("zmap-task.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(gojob_utils.NewTeeWriterCloser(os.Stdout, fd))
}

func main() {
	s := scheduler.New(option.Opt.Name).
		WithProvider(provider.Use("digitalocean", option.Opt.DigitalOceanToken)).
		WithCreateServerOptions(
			api.NewCreateServerOptions().
				WithName(option.Opt.DropletName).
				WithTag(option.Opt.DropletName).
				WithRegion(option.Opt.DropletRegion).
				WithSize(option.Opt.DropletSize).
				WithImage(option.Opt.DropletImage).
				WithPrivateKeyPath(option.Opt.DropletPrivateKeyPath).
				WithPublicKeyPath(option.Opt.DropletPublicKeyPath).
				WithPublicKeyName("zmap"),
		).
		WithMaxConcurrency(1)
	for t := range zmap_task.Generate(option.Opt.Name) {
		s.Submit(t)
	}
	s.Wait()
}
