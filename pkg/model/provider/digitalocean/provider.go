package digitalocean

import (
	"os"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/provider/api"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/server"
)

type Provider struct {
	do *DigitalOcean
}

func NewProvider(token string) *Provider {
	return &Provider{
		do: newDigitalOcean(token),
	}
}

func (p *Provider) ListServers() []server.Server {
	servers := []server.Server{}
	remoteServers := p.do.ListDroplets()
	for _, remoteServer := range remoteServers {
		servers = append(servers, NewServer(remoteServer))
	}
	return servers
}

func (p *Provider) ListServersByName(name string) []server.Server {
	servers := []server.Server{}
	remoteServers := p.ListServers()
	for _, remoteServer := range remoteServers {
		if remoteServer.Name() == name {
			servers = append(servers, remoteServer)
		}
	}
	return servers
}

func (p *Provider) ListServersByTag(tag string) []server.Server {
	servers := []server.Server{}
	remoteServers := p.ListServers()
	for _, remoteServer := range remoteServers {
		for _, remoteServerTag := range remoteServer.Tags() {
			if remoteServerTag == tag {
				servers = append(servers, remoteServer)
			}
		}
	}
	return servers
}

func (p *Provider) CreateKeyPair(name string, pubkey string) error {
	_, err := p.do.CreateSSHKeyPair(name, pubkey)
	return err
}

func (p *Provider) CreateServer(cso *api.CreateServerOptions) (server.Server, error) {
	pubkey, err := os.ReadFile(cso.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	err = p.CreateKeyPair(cso.Name, string(pubkey))
	if err != nil {
		return nil, err
	}

	droplet, err := p.do.CreateDroplet(
		cso.Name,
		cso.Region,
		cso.Size,
		cso.Image,
		string(pubkey),
		cso.Tag,
	)
	if err != nil {
		return nil, err
	}
	server := NewServer(*droplet)
	return server, nil
}

func (p *Provider) DestroyServerByName(name string) error {
	return p.do.DestroyDropletByName(name)
}

func (p *Provider) DestroyServerByTag(tag string) error {
	return p.do.DestroyDropletByTag(tag)
}
