package digitalocean

import (
	"context"
	"time"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/util/sshutil"
	"github.com/charmbracelet/log"
	"github.com/digitalocean/godo"
)

type DigitalOcean struct {
	client *godo.Client
}

func newDigitalOcean(token string) *DigitalOcean {
	return &DigitalOcean{
		client: godo.NewFromToken(token),
	}
}

func (d *DigitalOcean) CreateSSHKeyPair(name string, pubkey string) (*godo.Key, error) {
	fingerprint, err := sshutil.GetSSHPublicKeyFingerprintMD5(pubkey)
	if err != nil {
		return nil, err
	}
	log.Info("public key fingerprint", "fingerprint", fingerprint)
	// Check if key already exists
	key, _, err := d.client.Keys.GetByFingerprint(context.Background(), fingerprint)
	if err == nil {
		log.Info("ssh key already exists", "name", key.Name, "fingerprint", key.Fingerprint)
		return key, nil
	}
	// If key does not exist, create it
	key, _, err = d.client.Keys.Create(context.Background(), &godo.KeyCreateRequest{
		Name:      name,
		PublicKey: pubkey,
	})
	if err != nil {
		return nil, err
	}
	log.Info("ssh key created", "name", key.Name, "fingerprint", key.Fingerprint)
	return key, nil
}

func (d *DigitalOcean) CreateDroplet(name, region, size, image, pubkey string, tags []string) (*godo.Droplet, error) {
	var err error
	fingerprint, err := sshutil.GetSSHPublicKeyFingerprintMD5(pubkey)
	if err != nil {
		return nil, err
	}
	log.Info("retrieving ssh key", "fingerprint", fingerprint)
	key, _, err := d.client.Keys.GetByFingerprint(context.Background(), fingerprint)
	if err != nil {
		log.Error("error occured while retrieving ssh key", "error", err.Error())
		return nil, err
	}
	log.Info("creating droplet", "name", name, "region", region, "size", size, "image", image)
	var gd *godo.Droplet
	gd, _, err = d.client.Droplets.Create(context.Background(), &godo.DropletCreateRequest{
		Name:   name,
		Region: region,
		Size:   size,
		Image: godo.DropletCreateImage{
			Slug: image,
		},
		SSHKeys: []godo.DropletCreateSSHKey{
			{
				Fingerprint: key.Fingerprint,
			},
		},
		Tags: tags,
	})
	if err != nil {
		log.Error("error occured while creating droplet", "error", err.Error())
		return nil, err
	}
	log.Info("droplet created", "droplet_id", gd.ID)
	var status string = gd.Status
	var numTries int = 0
	for status != "active" {
		gd, _, err = d.client.Droplets.Get(context.Background(), gd.ID)
		if err != nil {
			log.Error("error occured while getting droplet", "error", err.Error())
			continue
		}
		ip, _ := gd.PublicIPv4()
		status = gd.Status
		log.Warn("waiting", "droplet_id", gd.ID, "ip", ip, "status", status, "num_tries", numTries)
		time.Sleep(1 * time.Second)
		numTries++
	}
	return gd, nil
}

func (d *DigitalOcean) ListDroplets() []godo.Droplet {
	droplets, _, err := d.client.Droplets.List(context.Background(), &godo.ListOptions{})
	if err != nil {
		log.Error("error occured when listing droplets", "error", err.Error())
		return droplets
	}
	return droplets
}

func (d *DigitalOcean) DestroyDropletByName(name string) error {
	droplets := d.ListDroplets()
	for _, droplet := range droplets {
		if droplet.Name == name {
			ip, _ := droplet.PublicIPv4()
			log.Info("destroying droplet", "ip", ip)
			_, err := d.client.Droplets.Delete(context.Background(), droplet.ID)
			if err != nil {
				log.Error("error occured when deleting droplet", "error", err.Error())
				return err
			}
			log.Info("droplet destroyed", "ip", ip)
		}
	}
	return nil
}

func (d *DigitalOcean) DestroyDropletByTag(tag string) error {
	droplets := d.ListDroplets()
	for _, droplet := range droplets {
		for _, t := range droplet.Tags {
			if t == tag {
				ip, _ := droplet.PublicIPv4()
				log.Info("destroying droplet", "ip", ip)
				_, err := d.client.Droplets.Delete(context.Background(), droplet.ID)
				if err != nil {
					log.Error("error occured when deleting droplet", "error", err.Error())
					return err
				}
				log.Info("droplet destroyed", "ip", ip)
			}
		}
	}
	return nil
}
