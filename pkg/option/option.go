package option

type S3Option struct {
	S3AccessKey string `long:"s3-access-key" description:"AWS access key"`
	S3SecretKey string `long:"s3-secret-key" description:"AWS secret key"`
	S3Region    string `long:"s3-region" description:"AWS region" default:"us-west-2"`
	S3Bucket    string `long:"s3-bucket" description:"Bucket name" required:"true"`
}

type DigitalOceanOption struct {
	DigitalOceanToken string `long:"do-token" description:"DigitalOcean token" required:"true"`
	NumDroplets       int    `long:"num-droplets" description:"Number of droplets" required:"true" default:"2"`
}

type DropletOption struct {
	DropletSize           string `long:"droplet-size" description:"Droplet size" required:"true" default:"s-1vcpu-1gb"`
	DropletImage          string `long:"droplet-image" description:"Droplet image" required:"true" default:"docker-20-04"`
	DropletRegion         string `long:"droplet-region" description:"Droplet region" required:"true" default:"sfo2"`
	DropletPublicKeyPath  string `long:"droplet-public-key-path" description:"Public key path" required:"true"`
	DropletPrivateKeyPath string `long:"droplet-private-key-path" description:"Private key path" required:"true"`
}

type MetaOption struct {
	Name        string `long:"name" description:"Task name" required:"true"`
	LogFilePath string `long:"log-file-path" description:"Log file path" required:"true"`
	Version     func() `long:"version" description:"print version and exit" json:"-"`
}

type Option struct {
	S3Option
	DigitalOceanOption
	DropletOption
	MetaOption
}
