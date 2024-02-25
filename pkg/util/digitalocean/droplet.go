package digitalocean

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/config"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/util/sshutil"
	"github.com/digitalocean/godo"
)

var Client *godo.Client
var key *godo.Key
var err error

func init() {
	Client = godo.NewFromToken(config.Cfg.APIToken)
	key, err = CreateSSHKeyPair(config.Cfg.SSHKeyFolder, config.Cfg.SSHKeyName)
	if err != nil {
		panic(err)
	}
}

func CreateSSHKeyPair(folder, name string) (*godo.Key, error) {
	// Get public key fingerprint
	_, pubkey, err := sshutil.LoadOrCreateSSHKeyPair(folder, name)
	if err != nil {
		return nil, err
	}
	slog.Info("ssh key pair loaded", slog.String("name", name), slog.String("folder", folder))
	fingerprint, err := sshutil.GetSSHPublicKeyFingerprintMD5(pubkey)
	if err != nil {
		return nil, err
	}
	slog.Info("public key fingerprint", slog.String("fingerprint", fingerprint))
	// Check if key already exists
	key, _, err := Client.Keys.GetByFingerprint(context.Background(), fingerprint)
	if err == nil {
		slog.Info("ssh key already exists", slog.String("name", key.Name), slog.String("fingerprint", key.Fingerprint))
		return key, nil
	}
	// If key does not exist, create it
	key, _, err = Client.Keys.Create(context.Background(), &godo.KeyCreateRequest{
		Name:      name,
		PublicKey: pubkey,
	})
	if err != nil {
		return nil, err
	}
	slog.Info("ssh key created", slog.String("name", key.Name), slog.String("fingerprint", key.Fingerprint))
	return key, nil
}

func CreateDroplet(name, region, size, image, fingerprint string, tags []string) (*godo.Droplet, error) {
	var err error
	slog.Info("retrieving ssh key", slog.String("fingerprint", fingerprint))
	key, _, err := Client.Keys.GetByFingerprint(context.Background(), fingerprint)
	if err != nil {
		slog.Error("error occured while retrieving ssh key", slog.String("error", err.Error()))
		return nil, err
	}
	slog.Info("creating droplet", slog.String("name", name), slog.String("region", region), slog.String("size", size), slog.String("image", image))
	var d *godo.Droplet
	d, _, err = Client.Droplets.Create(context.Background(), &godo.DropletCreateRequest{
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
		slog.Error("error occured while creating droplet", slog.String("error", err.Error()))
		return nil, err
	}
	slog.Info("droplet created", slog.Int("droplet_id", d.ID))
	var status string = d.Status
	for status != "active" {
		d, _, err = Client.Droplets.Get(context.Background(), d.ID)
		if err != nil {
			slog.Error("error occured while getting droplet", slog.String("error", err.Error()))
			continue
		}
		ip, _ := d.PublicIPv4()
		status = d.Status
		slog.Info("waiting", slog.Int("droplet_id", d.ID), slog.String("ip", ip), slog.String("status", status))
		time.Sleep(1 * time.Second)
	}
	return d, nil
}

func LoadDropletsFromRemoteAPI(tag string) []godo.Droplet {
	droplets, _, err := Client.Droplets.ListByTag(context.Background(), tag, &godo.ListOptions{})
	if err != nil {
		slog.Error("error occured when listing droplets", slog.String("error", err.Error()))
		return droplets
	}
	return droplets
}

func LoadDropletsFromLocalFile(path string) (droplets []godo.Droplet) {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("error occured when reading droplets file", slog.String("error", err.Error()))
		return droplets
	}
	err = json.Unmarshal(data, &droplets)
	if err != nil {
		slog.Error("error occured when unmarshalling droplets", slog.String("error", err.Error()))
		return droplets
	}
	return droplets
}

func SaveDropletsToLocalFile(droplets []godo.Droplet, path string) error {
	data, err := json.Marshal(droplets)
	if err != nil {
		slog.Error("error occured when marshalling droplets", slog.String("error", err.Error()))
		return err
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		slog.Error("error occured when writing droplets file", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func DestroyDroplets(droplets []godo.Droplet) {
	for _, droplet := range droplets {
		ip, _ := droplet.PublicIPv4()
		slog.Info("destroying droplet", slog.String("ip", ip))
		_, err := Client.Droplets.Delete(context.Background(), droplet.ID)
		if err != nil {
			slog.Error("error occured when deleting droplet", slog.String("error", err.Error()))
			continue
		}
		slog.Info("droplet destroyed", slog.String("ip", ip))
	}
}

func EnsureDroplets(numDroplets int) []godo.Droplet {
	for {
		droplets := LoadDropletsFromRemoteAPI(config.Cfg.Tag)
		slog.Info("retrieved droplets", slog.Int("num_droplets", len(droplets)))
		if len(droplets) >= numDroplets {
			slog.Info("all required droplets created")
			break
		}
		slog.Info("required droplets not satisfied, creating new droplet", slog.Int("remaining", numDroplets-len(droplets)))
		name := fmt.Sprintf("%s-%02d", config.Cfg.Name, len(droplets)+1)
		d, err := CreateDroplet(
			name,
			config.Cfg.Region,
			config.Cfg.Size,
			config.Cfg.Image,
			key.Fingerprint,
			[]string{config.Cfg.Tag},
		)
		if err != nil {
			slog.Error("error occured when creating droplet", slog.String("error", err.Error()))
			continue
		}
		ip, err := d.PublicIPv4()
		if err != nil {
			slog.Error("error occured when retrieving droplet ip", slog.String("error", err.Error()))
			continue
		}
		slog.Info("droplet created", slog.String("name", d.Name), slog.String("ip", ip))
	}

	droplets := LoadDropletsFromRemoteAPI(config.Cfg.Tag)
	for _, droplet := range droplets {
		ip, err := droplet.PublicIPv4()
		if err != nil {
			slog.Error("error occured when retrieving droplet ip", slog.String("error", err.Error()))
			continue
		}
		slog.Info("droplet", slog.String("ip", ip))
	}
	return droplets
}
