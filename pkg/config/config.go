package config

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/option"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DigitalOcean DigitalOcean `json:"digitalocean" toml:"digitalocean" yaml:"digitalocean"`
}

type DigitalOcean struct {
	Token   string        `json:"token" toml:"token" yaml:"token"`
	Destroy bool          `json:"destroy" toml:"destroy" yaml:"destroy"`
	Droplet DropletConfig `json:"droplet" toml:"droplet" yaml:"droplet"`
	SSH     SSHConfig     `json:"ssh" toml:"ssh" yaml:"ssh"`
}

type DropletConfig struct {
	Name   string `json:"name" toml:"name" yaml:"name"`
	Region string `json:"region" toml:"region" yaml:"region"`
	Size   string `json:"size" toml:"size" yaml:"size"`
	Image  string `json:"image" toml:"image" yaml:"image"`
	Tag    string `json:"tag" toml:"tag" yaml:"tag"`
}

type SSHConfig struct {
	Key  SSHKeyConfig `json:"ssh_key" toml:"ssh_key" yaml:"ssh_key"`
	Port int          `json:"port" toml:"port" yaml:"port"`
	User string       `json:"user" toml:"user" yaml:"user"`
}

type SSHKeyConfig struct {
	Name   string `json:"name" toml:"name" yaml:"name"`
	Folder string `json:"folder" toml:"folder" yaml:"folder"`
}

var Cfg Config

func loadToml(path string) {
	if _, err := toml.DecodeFile(path, &Cfg); err != nil {
		slog.Error("error occured while decoding config file", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func readFile(path string) ([]byte, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	content, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func loadJson(path string) {
	content, err := readFile(path)
	if err != nil {
		slog.Error("error occured while reading config file", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := json.Unmarshal(content, &Cfg); err != nil {
		slog.Error("error occured while decoding config file", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func loadYaml(path string) {
	content, err := readFile(path)
	if err != nil {
		slog.Error("error occured while reading config file", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := yaml.Unmarshal(content, &Cfg); err != nil {
		slog.Error("error occured while decoding config file", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func init() {
	configFilePath := option.Opts.ConfigFilePath
	extension := filepath.Ext(configFilePath)
	switch extension {
	case ".toml":
		slog.Info("loading toml config file", slog.String("path", configFilePath))
		loadToml(configFilePath)
	case ".json":
		slog.Info("loading json config file", slog.String("path", configFilePath))
		loadJson(configFilePath)
	case ".yaml":
		slog.Info("loading yaml config file", slog.String("path", configFilePath))
		loadYaml(configFilePath)
	default:
		slog.Error("unsupported config file extension", slog.String("extension", extension))
		os.Exit(1)
	}
}