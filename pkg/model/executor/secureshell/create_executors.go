package secureshell

import (
	"log/slog"
	"path/filepath"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/util/digitalocean"
)

func GenerateExecutors(numExecutors int) []*SSHExecutor {
	executors := []*SSHExecutor{}
	for _, droplet := range digitalocean.EnsureDroplets(numExecutors) {
		ip, err := droplet.PublicIPv4()
		if err != nil {
			slog.Error("error occured when retrieving droplet ip", slog.String("error", err.Error()))
			continue
		}
		e := NewSSHExecutor(
			ip,
			config.Cfg.SSHPort,
			config.Cfg.SSHUser,
			filepath.Join(config.Cfg.SSHKeyFolder, config.Cfg.SSHKeyName),
		)
		executors = append(executors, e)
	}
	return executors
}
