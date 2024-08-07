package option

import (
	"os"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/option"
	"github.com/WangYihang/gojob/pkg/version"
	"github.com/jessevdk/go-flags"
)

type ZMapOption struct {
	Port      int    `long:"port" description:"Port" required:"true"`
	BandWidth string `long:"bandwidth" description:"Bandwidth" required:"true" default:"1M"`
}

type Option struct {
	option.S3Option
	option.DigitalOceanOption
	option.DropletOption
	option.MetaOption
	ZMapOption
}

var Opt Option

func init() {
	Opt.Version = version.PrintVersion
	if _, err := flags.Parse(&Opt); err != nil {
		os.Exit(1)
	}
}
