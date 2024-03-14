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
		for _, ip := range map[string]string{
			"akamai-enhanced-tls": "23.43.51.134",
			"akamai-standard-tls": "96.7.129.39",
			"aliabab-dcdn":        "121.194.7.13",
			"alibaba-cdn":         "121.194.7.229",
			"alibaba-waf":         "47.110.175.195",
			"baidu-origin-leak":   "112.46.4.36",
			"baidu":               "222.199.191.41",
			"cloudflare":          "104.21.30.14",
			"ctyun":               "36.110.220.111",
			"fastly":              "151.101.108.249",
			"frontdoor":           "13.107.246.74",
			"huawei":              "182.118.39.151",
			"jd":                  "119.188.208.2",
			"ksyun-origin-leak":   "39.175.124.168",
			"ksyun":               "183.61.168.1",
			"qiniu":               "112.65.203.41",
			"tencent":             "219.151.137.59",
			"ucloud":              "42.81.55.129",
			"upyun":               "61.139.65.250",
		} {
			additionalArguments := map[string]string{
				"ip": ip,
			}
			for task := range cdn_domain_harvest.Generate(image, label, additionalArguments) {
				out <- task
			}
		}
	}()
	return out
}
