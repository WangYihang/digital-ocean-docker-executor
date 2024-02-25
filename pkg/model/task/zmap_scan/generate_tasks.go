package zmap_scan

import "fmt"

func GenerateTasks(numTasks int) []*Task {
	tasks := []*Task{}
	port := 80
	for i := 0; i < numTasks; i++ {
		task_name := fmt.Sprintf("zmap-tcp-%d", port)
		shards := numTasks
		shard := i
		t := NewTask(
			task_name,
			ZMapArguments{
				TargetPort:           port,
				OutputFileName:       fmt.Sprintf("%s-%d-%d.txt", task_name, shard, shards),
				LogFileName:          fmt.Sprintf("%s-%d-%d.log", task_name, shard, shards),
				StatusUpdateFileName: fmt.Sprintf("%s-%d-%d.status", task_name, shard, shards),
				BandWidth:            "90M",
				Seed:                 0,
				Shard:                shard,
				Shards:               shards,
			},
			"examples/zmap/Dockerfile",
		)
		tasks = append(tasks, t)
	}
	return tasks
}
