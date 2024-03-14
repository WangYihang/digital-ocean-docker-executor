package cdn_domain_harvest

import (
	"fmt"
	"path/filepath"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
)

func Generate(image string, label string, additionalArguments map[string]string) chan *task.DockerTask {
	out := make(chan *task.DockerTask)
	go func() {
		defer close(out)
		numShards := 1024
		numWorkers := 256
		for shard := 0; shard < numShards; shard++ {
			if shard > 1 {
				break
			}
			hostInputFileFolder := filepath.Join("/root", ".tranco/")
			hostOutputFileFolder := filepath.Join("/root", label)
			containerInputFileFolder := "/tranco/"
			containerOutputFileFolder := "/data/"
			t := task.NewDockerTask().
				WithDockerImage(image).
				WithPrivileged().
				WithNetwork("host").
				WithDetach().WithInteractive().WithTty().
				WithMount(hostInputFileFolder, containerInputFileFolder).
				WithMount(hostOutputFileFolder, containerOutputFileFolder).
				WithLabel("dode.task", label).
				WithLabel("dode.shard", fmt.Sprintf("%d", shard)).
				WithLabel("dode.num-shard", fmt.Sprintf("%d", numShards))
			for k, v := range additionalArguments {
				t.WithLabel(fmt.Sprintf("dode.%s", k), v)
				t.WithArguments(fmt.Sprintf("--%s", k), v)
			}
			outputFolderName := t.GetOutputFolderName()
			containerInputFilePath := filepath.Join(containerInputFileFolder, "2024-01-01_fqdn_full_G6Y6K.csv")
			containerOutputFolderPath := filepath.Join(containerOutputFileFolder, outputFolderName)
			out <- t.WithArguments(
				"-i", containerInputFilePath,
				"--input-file-format", "tranco",
				"-o", containerOutputFolderPath,
				"--shard", fmt.Sprintf("%d", shard),
				"--num-shards", fmt.Sprintf("%d", numShards),
				"--num-workers", fmt.Sprintf("%d", numWorkers),
			).WithLabel("dode.output", outputFolderName)
		}
	}()
	return out
}
