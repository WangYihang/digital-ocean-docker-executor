package brute_dns_resolve_task

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	cdn_domain_harvest "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/cdn-domain-harvest"
)

func Generate(label string) chan *task.DockerTask {
	out := make(chan *task.DockerTask)
	go func() {
		defer close(out)
		image := "ghcr.io/wangyihang/cdn-domain-harvest/brute-dns-verification:v0.0.7"
		for _, prefix := range map[string]string{
			"alibaba-cdn": "verification",
			"alibaba-waf": "wafdnscheck",
			"baidu":       "bdy-verify",
			"cloudflare":  "cloudflare-verify",
			"ctyun":       "dnsverify",
			"frontdoor":   "_dnsauth",
			"huawei":      "cdn_verification",
			"jd":          "_cdnautover",
			"ksyun":       "ksy-cdnauth",
			"tencent":     "_cdnauth",
			"upyun":       "upyun-verify",
		} {
			additionalArguments := map[string]string{
				"prefix": prefix,
			}
			for task := range cdn_domain_harvest.Generate(image, label, additionalArguments) {
				out <- task
			}
		}
	}()
	return out
}
