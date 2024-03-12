package option

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	ConfigFilePath string `short:"c" long:"config" description:"Path to the config file" required:"true" default:"config.toml"`
}

var Opts Options

func init() {
	if _, err := flags.Parse(&Opts); err != nil {
		os.Exit(1)
	}
}
