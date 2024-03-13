package brute_dns_resolve_task

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	cdn_domain_harvest "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/cdn-domain-harvest"
)

func Generate(label string) chan *task.DockerTask {
	return cdn_domain_harvest.Generate("ghcr.io/wangyihang/cdn-domain-harvest/brute-dns-resolve:v0.0.5", label)
}
