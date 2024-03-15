package zmap_task

import (
	"fmt"
	"strings"
)

type ZMapArguments struct {
	TargetPort           int
	Subnets              []string
	OutputFileName       string
	LogFileName          string
	StatusUpdateFileName string
	BandWidth            string
	Seed                 int
	Shards               int
	Shard                int
}

func NewZmapArguments() *ZMapArguments {
	return &ZMapArguments{
		TargetPort:           80,
		Subnets:              []string{},
		OutputFileName:       "zmap.output",
		LogFileName:          "zmap.log",
		StatusUpdateFileName: "zmap.status",
		BandWidth:            "1M",
		Seed:                 0,
		Shards:               1,
		Shard:                0,
	}
}

func (z *ZMapArguments) WithTargetPort(port int) *ZMapArguments {
	z.TargetPort = port
	return z
}

func (z *ZMapArguments) WithSubnet(subnet string) *ZMapArguments {
	z.Subnets = append(z.Subnets, subnet)
	return z
}

func (z *ZMapArguments) WithOutputFileName(fileName string) *ZMapArguments {
	z.OutputFileName = fileName
	return z
}

func (z *ZMapArguments) WithLogFileName(fileName string) *ZMapArguments {
	z.LogFileName = fileName
	return z
}

func (z *ZMapArguments) WithStatusUpdateFileName(fileName string) *ZMapArguments {
	z.StatusUpdateFileName = fileName
	return z
}

func (z *ZMapArguments) WithBandWidth(bandWidth string) *ZMapArguments {
	z.BandWidth = bandWidth
	return z
}

func (z *ZMapArguments) WithSeed(seed int) *ZMapArguments {
	z.Seed = seed
	return z
}

func (z *ZMapArguments) WithShards(shards int) *ZMapArguments {
	z.Shards = shards
	return z
}

func (z *ZMapArguments) WithShard(shard int) *ZMapArguments {
	z.Shard = shard
	return z
}

func (z *ZMapArguments) String() string {
	arguments := []string{
		"--target-port", fmt.Sprintf("%d", z.TargetPort),
		"--output-file", fmt.Sprintf("/data/%s", z.OutputFileName),
		"--log-file", fmt.Sprintf("/data/%s", z.LogFileName),
		"--status-updates-file", fmt.Sprintf("/data/%s", z.StatusUpdateFileName),
		"--bandwidth", z.BandWidth,
		"--seed", fmt.Sprintf("%d", z.Seed),
		"--shards", fmt.Sprintf("%d", z.Shards),
		"--shard", fmt.Sprintf("%d", z.Shard),
	}
	arguments = append(arguments, z.Subnets...)
	return strings.Join(arguments, " ")
}
