package http_grab

import (
	"fmt"
	"path/filepath"
)

// GenerateTasks

func GenerateTasks(numTasks int) []*Task {
	tasks := []*Task{}
	port := 80
	for i := 0; i < numTasks; i++ {
		task_name := fmt.Sprintf("http-grab-%d", port)
		shards := numTasks
		shard := i
		inputFilePath := "data/zmap-tcp-80/zmap-tcp-80.txt"
		inputFileName := filepath.Base(inputFilePath)
		t := NewTask(
			task_name,
			"examples/http-grab/Dockerfile",
			inputFilePath,
			HTTPGrabArguments{
				InputFilePath:        filepath.Join("/data", inputFileName),
				OutputFilePath:       fmt.Sprintf("/data/%s-%d-%d.txt", task_name, shard, shards),
				StatusUpdateFilePath: fmt.Sprintf("/data/%s-%d-%d.status", task_name, shard, shards),
				NumWorkers:           4096,
				Timeout:              8,
				Port:                 port,
				Path:                 "/verification.html",
				Host:                 "",
				MaxTries:             1,
				NumShards:            int64(shards),
				Shard:                int64(shard),
			},
		)
		tasks = append(tasks, t)
	}
	return tasks
}
