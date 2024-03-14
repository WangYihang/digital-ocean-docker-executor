package brute_dns_resolve_task

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task"
	cdn_domain_harvest "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/task/cdn-domain-harvest"
)

func Generate(label string) chan *task.DockerTask {
	out := make(chan *task.DockerTask)
	go func() {
		defer close(out)
		image := "ghcr.io/wangyihang/cdn-domain-harvest/brute-cname-subdomain:v0.0.7"
		for _, suffix := range map[string]string{
			"akamai-enhanced-tls": "edgekey.net",
			"akamai-standard-tls": "edgesuite.net",
			"alibaba-cdn":         "w.cdngslb.com",
			"alibaba-dcdn":        "w.kunlunpi.com",
			"baidu":               "a.bdydns.com",
			"cloudflare":          "cdn.cloudflare.net",
			"ctyun":               "ctadns.cn",
			"jd":                  "galileo.jcloud-cdn.com",
			"kingsoft":            "download.ks-cdn.com",
			"tencent":             "dsa.dnsv1.com.cn",
			"ucloud":              "ucloud.com.cn",
		} {
			additionalArguments := map[string]string{
				"suffix": suffix,
			}
			for task := range cdn_domain_harvest.Generate(image, label, additionalArguments) {
				out <- task
			}
		}
	}()
	return out
}
