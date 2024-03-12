package alibaba

import "github.com/WangYihang/digital-ocean-docker-executor/pkg/model/server"

type AlibabaProvider struct {
}

func New() *AlibabaProvider {
	return &AlibabaProvider{}
}

func (a *AlibabaProvider) ListServers() []server.Server {
	panic("not implemented")
}

func (a *AlibabaProvider) ListServersByName(name string) []server.Server {
	panic("not implemented")
}

func (a *AlibabaProvider) ListServersByTag(tag string) []server.Server {
	panic("not implemented")
}

func (a *AlibabaProvider) CreateKeyPair(name string, pubkey string) error {
	panic("not implemented")
}

func (a *AlibabaProvider) CreateServer(name string, tag string) (server.Server, error) {
	panic("not implemented")
}

func (a *AlibabaProvider) DestroyServerByName(name string) error {
	panic("not implemented")
}

func (a *AlibabaProvider) DestroyServerByTag(tag string) error {
	panic("not implemented")
}
