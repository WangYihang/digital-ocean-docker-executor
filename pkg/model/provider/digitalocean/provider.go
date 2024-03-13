package digitalocean

import (
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/model/server"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/util/sshutil"
)

type Provider struct {
	do *DigitalOcean
}

func NewProvider() *Provider {
	return &Provider{
		do: newDigitalOcean(config.Cfg.DigitalOcean.Token),
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

func (p *Provider) CreateServer(name string, tag string) (server.Server, error) {
	pubkey, _, err := sshutil.LoadOrCreateSSHKeyPair(config.Cfg.DigitalOcean.SSH.Key.Folder, config.Cfg.DigitalOcean.SSH.Key.Name)
	if err != nil {
		return nil, err
	}

	p.CreateKeyPair(name, pubkey)

	droplet, err := p.do.CreateDroplet(
		name,
		config.Cfg.DigitalOcean.Droplet.Region,
		config.Cfg.DigitalOcean.Droplet.Size,
		config.Cfg.DigitalOcean.Droplet.Image,
		pubkey,
		tag,
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
