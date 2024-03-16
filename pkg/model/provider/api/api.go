package api

type CreateServerOptions struct {
	Name           string
	Tag            string
	Region         string
	Size           string
	Image          string
	PublicKeyName  string
	PublicKeyPath  string
	PrivateKeyPath string
}

func NewCreateServerOptions() *CreateServerOptions {
	return &CreateServerOptions{
		Name:           "default",
		Tag:            "default",
		Region:         "nyc1",
		Size:           "s-1vcpu-1gb",
		Image:          "ubuntu-20-04-x64",
		PublicKeyPath:  "id_rsa.pub",
		PrivateKeyPath: "id_rsa",
	}
}

func (cso *CreateServerOptions) WithName(name string) *CreateServerOptions {
	cso.Name = name
	return cso
}

func (cso *CreateServerOptions) WithTag(tag string) *CreateServerOptions {
	cso.Tag = tag
	return cso
}

func (cso *CreateServerOptions) WithRegion(region string) *CreateServerOptions {
	cso.Region = region
	return cso
}

func (cso *CreateServerOptions) WithSize(size string) *CreateServerOptions {
	cso.Size = size
	return cso
}

func (cso *CreateServerOptions) WithImage(image string) *CreateServerOptions {
	cso.Image = image
	return cso
}

func (cso *CreateServerOptions) WithPrivateKeyPath(privateKeyPath string) *CreateServerOptions {
	cso.PrivateKeyPath = privateKeyPath
	return cso
}

func (cso *CreateServerOptions) WithPublicKeyName(publicKeyName string) *CreateServerOptions {
	cso.PublicKeyName = publicKeyName
	return cso
}

func (cso *CreateServerOptions) WithPublicKeyPath(publicKeyPath string) *CreateServerOptions {
	cso.PublicKeyPath = publicKeyPath
	return cso
}
