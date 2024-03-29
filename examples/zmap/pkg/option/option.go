package option

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type S3Option struct {
	S3AccessKey string `long:"s3-access-key" description:"AWS access key"`
	S3SecretKey string `long:"s3-secret-key" description:"AWS secret key"`
	S3Region    string `long:"s3-region" description:"AWS region" default:"us-west-2"`
}

type DigitalOceanOption struct {
	DigitalOceanToken string `long:"do-token" description:"DigitalOcean token" required:"true"`
	NumDroplets       int    `long:"num-droplets" description:"Number of droplets" required:"true"`
}

type DropletOption struct {
	DropletName           string `long:"droplet-name" description:"Droplet name" required:"true" default:"zmap"`
	DropletSize           string `long:"droplet-size" description:"Droplet size" required:"true" default:"s-1vcpu-1gb"`
	DropletImage          string `long:"droplet-image" description:"Droplet image" required:"true" default:"docker-20-04"`
	DropletRegion         string `long:"droplet-region" description:"Droplet region" required:"true" default:"sfo2"`
	DropletPublicKeyPath  string `long:"droplet-public-key-path" description:"Public key path" required:"true"`
	DropletPrivateKeyPath string `long:"droplet-private-key-path" description:"Private key path" required:"true"`
}

type ZMapOption struct {
	Port      int    `long:"port" description:"Port" required:"true"`
	BandWidth string `long:"bandwidth" description:"Bandwidth" required:"true" default:"1M"`
}

type Option struct {
	S3Option
	DigitalOceanOption
	DropletOption
	ZMapOption
	Name        string `long:"name" description:"Task name" required:"true" default:"zmap-task"`
	LogFilePath string `long:"log-file-path" description:"Log file path" required:"true" default:"zmap-task.log"`
}

var Opt Option

func init() {
	if _, err := flags.Parse(&Opt); err != nil {
		os.Exit(1)
	}
}
