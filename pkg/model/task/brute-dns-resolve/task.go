package brute_dns_resolve_task

import (
	"fmt"
	"path/filepath"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	"github.com/charmbracelet/log"
)

func Generate(label string) chan *task.DockerTask {
	out := make(chan *task.DockerTask)
	go func() {
		defer close(out)
		numShards := 1024
		numWorkers := 256
		for shard := 0; shard < numShards; shard++ {
			hostInputFileFolder := filepath.Join("/root", ".tranco/")
			hostOutputFileFolder := filepath.Join("/root", label)
			containerInputFileFolder := "/tranco/"
			containerOutputFileFolder := "/data/"
			t := task.NewDockerTask().
				WithDockerImage("ghcr.io/wangyihang/cdn-domain-harvest/brute-dns-resolve:v0.0.4").
				WithPrivileged().
				WithNetwork("host").
				WithDetach().WithInteractive().WithTty().
				WithMount(hostInputFileFolder, containerInputFileFolder).
				WithMount(hostOutputFileFolder, containerOutputFileFolder).
				WithLabel("dode.task", label).
				WithLabel("dode.shard", fmt.Sprintf("%d", shard)).
				WithLabel("dode.num-shard", fmt.Sprintf("%d", numShards))
			outputFolderName := t.GetOutputFolderName()
			containerInputFilePath := filepath.Join(containerInputFileFolder, "2024-01-01_fqdn_full_G6Y6K.csv")
			containerOutputFolderPath := filepath.Join(containerOutputFileFolder, outputFolderName)
			containerConfigFilePath := "config/config.yaml"
			out <- t.WithArguments(
				"-i", containerInputFilePath,
				"-o", containerOutputFolderPath,
				"--shard", fmt.Sprintf("%d", shard),
				"--num-shards", fmt.Sprintf("%d", numShards),
				"--num-workers", fmt.Sprintf("%d", numWorkers),
				"--config-file-path", containerConfigFilePath,
			).WithLabel("dode.output", outputFolderName)
			if shard > 8 {
				log.Warn("Shard is greater than 8, breaking the loop")
				break
			}
		}
	}()
	return out
}
