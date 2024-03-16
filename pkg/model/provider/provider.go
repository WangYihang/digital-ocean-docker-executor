package provider

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider/alibaba"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider/api"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider/digitalocean"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/server"
)

type CloudServiceProvider interface {
	CreateKeyPair(name string, pub string) error
	CreateServer(*api.CreateServerOptions) (server.Server, error)
	ListServers() []server.Server
	ListServersByName(name string) []server.Server
	ListServersByTag(tag string) []server.Server
	DestroyServerByName(name string) error
	DestroyServerByTag(tag string) error
}

func Use(name, token string) CloudServiceProvider {
	providers := map[string]func(string) CloudServiceProvider{
		"alibaba": func(token string) CloudServiceProvider {
			return alibaba.NewProvider(token)
		},
		"digitalocean": func(token string) CloudServiceProvider {
			return digitalocean.NewProvider(token)
		},
	}
	if provider, ok := providers[name]; ok {
		return provider(token)
	}
	panic("unsupported cloud service provider")
}
