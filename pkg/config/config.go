package config

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	// Digital Ocean API Configuration
	APIToken                         string `yaml:"api_token"`
	DestroyDropletsAfterTaskFinished bool   `yaml:"destroy_droplets_after_task_finished"`
	// SSH Key Configuration
	SSHKeyFolder string `yaml:"ssh_key_folder"`
	SSHKeyName   string `yaml:"ssh_key_name"`
	SSHPort      int    `yaml:"ssh_port"`
	SSHUser      string `yaml:"ssh_user"`
	// Droplet Configuration
	Name   string `yaml:"name"`
	Tag    string `yaml:"tag"`
	Region string `yaml:"region"`
	Size   string `yaml:"size"`
	Image  string `yaml:"image"`
}

var Cfg *Config

func init() {
	body, err := os.ReadFile("config.yaml")
	if err != nil {
		slog.Error("error occured while reading config file", slog.String("error", err.Error()))
		os.Exit(1)
	}
	err = yaml.Unmarshal(body, &Cfg)
	if err != nil {
		slog.Error("error occured while unmarshalling config file", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
