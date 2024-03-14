package brute_dns_resolve_task

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	cdn_domain_harvest "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/cdn-domain-harvest"
)

func Generate(label string) chan *task.DockerTask {
	out := make(chan *task.DockerTask)
	go func() {
		defer close(out)
		image := "ghcr.io/wangyihang/cdn-domain-harvest/brute-http-verification:v0.0.7"
		for _, path := range map[string]string{
			"alibaba": "/verification.html",
			"ksyun":   "/ksy-cdnauth.html",
		} {
			additionalArguments := map[string]string{
				"path": path,
			}
			for task := range cdn_domain_harvest.Generate(image, label, additionalArguments) {
				out <- task
			}
		}
	}()
	return out
}
