package option

import (
	"os"

	"github.com/WangYihang/digital-ocean-docker-executor/pkg/option"
	"github.com/WangYihang/digital-ocean-docker-executor/pkg/version"
	"github.com/jessevdk/go-flags"
)

type HTTPGrabOption struct {
	Timeout int `long:"timeout" description:"Timeout" required:"true" default:"16"`
}

type Option struct {
	option.S3Option
	option.DigitalOceanOption
	option.DropletOption
	option.MetaOption
	HTTPGrabOption
}

var Opt Option

func init() {
	Opt.Version = version.PrintVersion
	if _, err := flags.Parse(&Opt); err != nil {
		os.Exit(1)
	}
}
