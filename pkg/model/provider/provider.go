package provider

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider/alibaba"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider/digitalocean"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/server"
)

type CloudServiceProvider interface {
	CreateKeyPair(name string, pub string) error
	CreateServer(name string, tag string) (server.Server, error)
	ListServers() []server.Server
	ListServersByName(name string) []server.Server
	ListServersByTag(tag string) []server.Server
	DestroyServerByName(name string) error
	DestroyServerByTag(tag string) error
}

func Use(name string) CloudServiceProvider {
	switch name {
	case "alibaba":
		return alibaba.New()
	case "digitalocean":
		return digitalocean.NewProvider()
	default:
		return digitalocean.NewProvider()
	}
}

func Default() CloudServiceProvider {
	return digitalocean.NewProvider()
}
