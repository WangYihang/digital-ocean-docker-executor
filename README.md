## Usage

```bash
$ mkdir -p .ssh
$ ssh-keygen -t rsa -b 4096 -N "" -f .ssh/id_rsa
```

```bash
$ go run examples/zmap/main.go \
    --droplet-public-key-path .ssh/id_rsa.pub \
    --droplet-private-key-path .ssh/id_rsa \
    --do-token dop_v1_**************************************************************** \
    --num-droplets 1 \
    --s3-access-key=******************** \
    --s3-secret-key=**************************************** \
    --s3-bucket=cdnmon-zmap \
    --port 80 \
    --bandwidth=100M
```