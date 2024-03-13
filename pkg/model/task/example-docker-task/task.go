package example_docker_task

import (
	"fmt"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
)

func Generate(label string) chan *task.DockerTask {
	out := make(chan *task.DockerTask)
	go func() {
		defer close(out)
		numShards := 1
		for shard := 0; shard < numShards; shard++ {
			out <- task.NewDockerTask().
				WithDockerImage("centos:latest").
				WithArguments("ping", "-c", "60", "8.8.8.8").
				WithDetach().WithInteractive().WithTty().
				WithLabel("dode.task", label).
				WithLabel("dode.shard", fmt.Sprintf("%d", shard)).
				WithLabel("dode.num-shard", fmt.Sprintf("%d", numShards))
		}
	}()
	return out
}
