package config

import (
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DigitalOcean DigitalOcean `json:"digitalocean" toml:"digitalocean"`
}

type DigitalOcean struct {
	Token   string        `json:"token" toml:"token"`
	Destroy bool          `json:"destroy" toml:"destroy"`
	Droplet DropletConfig `json:"droplet" toml:"droplet"`
	SSH     SSHConfig     `json:"ssh" toml:"ssh"`
}

type DropletConfig struct {
	Name   string `json:"name" toml:"name"`
	Region string `json:"region" toml:"region"`
	Size   string `json:"size" toml:"size"`
	Image  string `json:"image" toml:"image"`
	Tag    string `json:"tag" toml:"tag"`
}

type SSHConfig struct {
	Key  SSHKeyConfig `json:"ssh_key" toml:"ssh_key"`
	Port int          `json:"port" toml:"port"`
	User string       `json:"user" toml:"user"`
}

type SSHKeyConfig struct {
	Name   string `json:"name" toml:"name"`
	Folder string `json:"folder" toml:"folder"`
}

var Cfg *Config

func init() {
	var config Config
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		slog.Error("error occured while decoding config file", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
