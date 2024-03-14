package brute_dns_resolve_task

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	cdn_domain_harvest "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/cdn-domain-harvest"
)

func Generate(label string) chan *task.DockerTask {
	out := make(chan *task.DockerTask)
	go func() {
		defer close(out)
		image := "ghcr.io/wangyihang/cdn-domain-harvest/brute-dns-resolve:v0.0.7"
		for _, qtype := range []string{
			"A", "AAAA", "TXT", "NS", "CNAME",
		} {
			additionalArguments := map[string]string{
				"qtype": qtype,
			}
			for task := range cdn_domain_harvest.Generate(image, label, additionalArguments) {
				out <- task
			}
		}
	}()
	return out
}
